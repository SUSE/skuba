import hcl
import json
import logging
import os

from platforms.platform import Platform
from utils import (Format, step)

logger = logging.getLogger('testrunner')


class Terraform(Platform):
    def __init__(self, conf, platform):
        super().__init__(conf)
        self.tfdir = os.path.join(self.conf.terraform.tfdir, platform)
        self.tfjson_path = os.path.join(conf.workspace, "tfout.json")
        self.tfout_path = os.path.join(self.conf.workspace, "tfout")
        self.state = None

        self.logs["files"] += ["/var/run/cloud-init/status.json",
                               "/var/log/cloud-init-output.log",
                               "/var/log/cloud-init.log"]

        self.tmp_files = [self.tfout_path,
                          self.tfjson_path]

    def destroy(self, variables):
        cmd = "destroy -auto-approve"

        for var in variables:
            cmd += f" -var {var}"

        self._run_terraform_command(cmd)

    def _provision_platform(self):
        """ Create and apply terraform plan"""
        exception = None
        self._check_tf_deployed()

        self.utils.setup_ssh()

        init_cmd = "init"
        if self.conf.terraform.plugin_dir:
            logger.info(f"Installing plugins from {self.conf.terraform.plugin_dir}")
            init_cmd += f" -plugin-dir={self.conf.terraform.plugin_dir}"
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
            self._fetch_terraform_output()
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
        return self.state["modules"][0]["outputs"]["ip_load_balancer"]["value"]

    def get_nodes_ipaddrs(self, role):
        self._load_tfstate()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role: {}".format(role))

        role_key = "ip_"+role+"s"
        return self.state["modules"][0]["outputs"][role_key]["value"]

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

    def _update_tfvars(self, tfvars):
        new_vars = {
            "internal_net": self.conf.terraform.internal_net,
            "stack_name": self.conf.terraform.stack_name,
            "username": self.conf.nodeuser,
            "masters": self.conf.master.count,
            "workers": self.conf.worker.count,
            "authorized_keys": [self.utils.authorized_keys()]
        }

        for k, v in new_vars.items():
            if tfvars.get(k) is not None:
                if isinstance(v, list):
                    tfvars[k] = tfvars[k] + v
                elif isinstance(v, dict):
                    tfvars[k].update(v)
                else:
                    tfvars[k] = v

        # Update mirror urls
        repos = tfvars.get("repositories")
        if self.conf.terraform.mirror and repos is not None:
            for name, url in repos.items():
                tfvars["repositories"][name] = url.replace("download.suse.de", self.conf.terraform.mirror)

    def _run_terraform_command(self, cmd, env={}):
        """Running terraform command in {terraform.tfdir}/{platform}"""
        cmd = f'{self._env_setup_cmd()}; terraform {cmd}'

        # Terraform needs PATH and SSH_AUTH_SOCK
        sock_fn = self.utils.ssh_sock_fn()
        env["SSH_AUTH_SOCK"] = sock_fn
        env["PATH"] = os.environ['PATH']

        self.utils.runshellcommand(cmd, cwd=self.tfdir, env=env)

    def _check_tf_deployed(self):
        if os.path.exists(self.tfjson_path):
            raise Exception(Format.alert(f"tf file found. Please run cleanup and try again {self.tfjson_path}"))

    # TODO: this function is currently not used. Identify points where it should
    # be invoked
    def _verify_tf_dependency(self):
        if not os.path.exists(self.tfjson_path):
            raise Exception(Format.alert("tf file not found. Please run terraform and try again{}"))
