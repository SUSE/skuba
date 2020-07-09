import json
import logging
import os
from urllib.parse import urlparse

import hcl

from platforms.platform import Platform
from utils import (Format, step, Utils)

logger = logging.getLogger('testrunner')


class Terraform(Platform):
    def __init__(self, conf, platform):
        super().__init__(conf)
        if not conf.terraform.stack_name:
            raise ValueError("a terraform stack name must be specified")

        self.tfdir = os.path.join(self.conf.terraform.tfdir, platform)
        self.tfjson_path = os.path.join(self.conf.terraform.workdir, "tfout.json")
        self.tfout_path = os.path.join(self.conf.terraform.workdir, "tfout")
        self.utils = Utils(conf)
        self.state = None

        self.logs["files"] += ["/var/run/cloud-init/status.json",
                               "/var/log/cloud-init-output.log",
                               "/var/log/cloud-init.log"]

        self.tmp_files = [self.tfout_path,
                          self.tfjson_path]

    def destroy(self, variables=[]):
        cmd = "destroy -auto-approve"

        for var in variables:
            cmd += f" -var {var}"

        self._run_terraform_command(cmd)

    def _provision_platform(self, masters=-1, workers=-1):

        if masters > -1:
            self.conf.terraform.master.count = masters
        if workers > -1:
            self.conf.terraform.worker.count = workers

        exception = None
        self._check_tf_deployed()

        init_cmd = "init"
        if self.conf.terraform.plugin_dir:
            logger.info(f"Installing plugins from {self.conf.terraform.plugin_dir}")
            init_cmd += f" -plugin-dir={self.conf.terraform.plugin_dir}"
        else:
            init_cmd += f" -get-plugins=false"
        self._run_terraform_command(init_cmd)

        self._run_terraform_command("version")
        self._generate_tfvars_file()
        plan_cmd = f"plan -out {self.tfout_path}"
        apply_cmd = f"apply -auto-approve {self.tfout_path}"

        self._run_terraform_command(plan_cmd)

        try:
            self._run_terraform_command(apply_cmd)
        except Exception as ex:
            exception = ex
        finally:
            try:
                self._fetch_terraform_output()
            except Exception as inner_ex:
                # don't override original exception if any
                if not exception:
                    exception = inner_ex

        if exception:
            raise exception

    def _load_tfstate(self):
        if self.state is None:
            fn = os.path.join(self.tfdir, "terraform.tfstate")
            logger.debug("Reading configuration from {}".format(fn))
            with open(fn) as f:
                self.state = json.load(f)

    def get_lb_ipaddr(self):
        self._load_tfstate()
        if self.state["version"] == 3:
            return self.state["modules"][0]["outputs"]["ip_load_balancer"]["value"]["{}-lb".format(self.stack_name())]
        elif self.state["version"] == 4:
            return self.state["outputs"]["ip_load_balancer"]["value"]["{}-lb".format(self.stack_name())]

    def get_num_nodes(self, role):
        return len(self.get_nodes_ipaddrs(role))

    def get_nodes_names(self, role):
        stack_name = self.stack_name()
        return [f'caasp-{role}-{stack_name}-{i}' for i in range(self.get_num_nodes(role))]

    def get_nodes_ipaddrs(self, role):
        self._load_tfstate()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role: {}".format(role))

        role_key = "ip_" + role + "s"
        if self.state["version"] == 3:
            return list(self.state["modules"][0]["outputs"][role_key]["value"].values())
        elif self.state["version"] == 4:
            return list(self.state["outputs"][role_key]["value"].values())

    @step
    def _fetch_terraform_output(self):
        cmd = f"output -json >{self.tfjson_path}"
        self._run_terraform_command(cmd)

    def _generate_tfvars_file(self):
        """Generate terraform tfvars file"""
        tfvars_template = os.path.join(self.tfdir, self.conf.terraform.tfvars)
        tfvars_final = os.path.join(self.tfdir, "terraform.tfvars.json")

        with open(tfvars_template) as f:
            if '.json' in os.path.basename(tfvars_template).lower():
                tfvars = json.load(f)
            else:
                tfvars = hcl.load(f)

            self._update_tfvars(tfvars)

            with open(tfvars_final, "w") as f:
                json.dump(tfvars, f)

    # take up to 45 characters from stackname to give room to the fixed part
    # in the node name: caasp-[master|worker]-<stack name>-xxx (total length
    # must be <= 63).
    # Also ensure that only valid character are present and that the string
    # starts and ends with alphanumeric characters and all lowercase.
    def stack_name(self):
        stack_name = self.conf.terraform.stack_name[:45]
        stack_name = stack_name.replace("_", "-").replace("/", "-")
        stack_name = stack_name.strip("-.")
        stack_name = stack_name.lower()

        return stack_name

    def _update_tfvars(self, tfvars):
        new_vars = {
            "internal_net": self.conf.terraform.internal_net,
            "stack_name": self.stack_name(),
            "username": self.utils.ssh_user(),
            "masters": self.conf.terraform.master.count,
            "master_memory": self.conf.terraform.master.memory,
            "master_vcpu": self.conf.terraform.master.cpu,
            "workers": self.conf.terraform.worker.count,
            "worker_memory": self.conf.terraform.worker.memory,
            "worker_vcpu": self.conf.terraform.worker.cpu,
            "lbs": self.conf.terraform.lb.count,
            "lb_memory": self.conf.terraform.lb.memory,
            "lb_vcpu": self.conf.terraform.lb.cpu,
            "authorized_keys": [self.utils.authorized_keys()]
        }

        new_vars.update(self.platform_new_vars)

        for k, v in new_vars.items():
            if tfvars.get(k) is not None:
                if isinstance(v, list):
                    tfvars[k] = tfvars[k] + v
                elif isinstance(v, dict):
                    tfvars[k].update(v)
                else:
                    tfvars[k] = v

        # if registry code specified, repositories are not needed
        if self.conf.packages.registry_code:
            tfvars["caasp_registry_code"] = self.conf.packages.registry_code
            tfvars["repositories"] = {}

        repos = tfvars.get("repositories", {})
        if self.conf.packages.additional_repos:
            for name, url in self.conf.packages.additional_repos.items():
                if not url:
                    logger.warning(f'skipping repository {name} with empty url')
                    continue
                repos[name] = url

        # Update mirror urls
        if self.conf.packages.mirror and repos:
            for name, url in repos.items():
                url_parsed = urlparse(url)
                url_updated = url_parsed._replace(netloc=self.conf.packages.mirror)
                tfvars["repositories"][name] = url_updated.geturl()

        if self.conf.packages.additional_pkgs:
            tfvars["packages"].extend(self.conf.packages.additional_pkgs)

    def _run_terraform_command(self, cmd, env={}):
        """Running terraform command in {terraform.tfdir}/{platform}"""
        env['TF_LOG'] = self.conf.terraform.log_level
        cmd = f'{self._env_setup_cmd()}; terraform {cmd}'
        self.utils.runshellcommand(cmd, cwd=self.tfdir, env=env)

    def _check_tf_deployed(self):
        if os.path.exists(self.tfjson_path):
            raise Exception(Format.alert(f"tf file found. Please run cleanup and try again {self.tfjson_path}"))

    # TODO: this function is currently not used. Identify points where it should
    # be invoked
    def _verify_tf_dependency(self):
        if not os.path.exists(self.tfjson_path):
            raise Exception(Format.alert("tf file not found. Please run terraform and try again{}"))
