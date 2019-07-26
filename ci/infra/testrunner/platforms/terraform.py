import hcl
import json
import logging
import os
import subprocess

from timeout_decorator import timeout

from utils import (Format, step, Utils)

logger = logging.getLogger('testrunner')


class Terraform:
    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(conf)
        self.tfdir = os.path.join(self.conf.terraform.tfdir, platform)
        self.tfjson_path = os.path.join(conf.workspace, "tfout.json")
        self.tfout_path = os.path.join(self.conf.workspace, "tfout")
        self.state = None

    def _env_setup_cmd(self):
        """Returns the command for setting up the platform environment"""
        return ""

    def _cleanup_platform(self):
        """Platform specific cleanup. Expected to be overridden by platforms"""

    def _get_platform_logs(self):
        """Platform specific logs to collect. Expected to be overridden by platforms"""
        return False

    def cleanup(self):
        """ Clean up """
        try:
            self._cleanup_platform()
        except Exception as ex:
            cleanup_failure = True
            logger.warning("Received the following error: '{}'\n"
                           "Attempting to finish cleanup".format(ex))
            raise Exception("Failure(s) during cleanup") from ex
        finally:
            dirs = [os.path.join(self.conf.workspace, "tfout"),
                self.tfjson_path]
            for dir in dirs:
                try:
                    self.utils.runshellcommand("rm -rf {}".format(dir))
                except Exception as ex:
                    logger.warning("Received the following error: '{}'\n"
                                   "Attempting to finish cleanup".format(ex))

    def destroy(self, variables):
        cmd = "destroy -auto-approve"

        for var in variables:
            cmd += f" -var {var}"

        self._run_terraform_command(cmd)

    @timeout(600)
    @step
    def gather_logs(self):
        logging_errors = False

        node_ips = {"master": self.get_nodes_ipaddrs("master"),
                    "worker": self.get_nodes_ipaddrs("worker")}
        logs = {"files": ["/var/run/cloud-init/status.json",
                          "/var/log/cloud-init-output.log",
                          "/var/log/cloud-init.log"],
                "dirs": ["/var/log/pods"],
                "services": ["kubelet"]}

        if not os.path.isdir(self.conf.log_dir):
            os.mkdir(self.conf.log_dir)
            logger.info(f"Created log dir {self.conf.log_dir}")

        for node_type in node_ips:
            for ip_address in node_ips[node_type]:
                node_log_dir = self._create_node_log_dir(ip_address, node_type, self.conf.log_dir)
                logging_error = self.utils.collect_remote_logs(ip_address, logs, node_log_dir)

                if logging_error:
                    logging_errors = logging_error

        platform_log_error = self._get_platform_logs()

        if platform_log_error:
            logging_errors = platform_log_error

        return logging_errors

    @step
    def provision(self, num_master=-1, num_worker=-1):
        """ Create and apply terraform plan"""
        if num_master > -1 or num_worker > -1:
            logger.warning("Overriding number of nodes")
            if num_master > -1:
                self.conf.master.count = num_master
                logger.warning("   Masters:{} ".format(num_master))

            if num_worker > -1:
                self.conf.worker.count = num_worker
                logger.warning("   Workers:{} ".format(num_worker))

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

        # TODO: define the number of retries as a configuration parameter
        for retry in range(1, 5):
            self._run_terraform_command(plan_cmd)
            try:
                self._run_terraform_command(apply_cmd)
                break

            except Exception as ex :
                logger.warning("Failed terraform apply attempt {}/5".format(retry))
                if retry == 4:
                    raise Exception("Failed Openstack Terraform deployment") from ex
            finally:
                self._fetch_terraform_output()

    @staticmethod
    def _create_node_log_dir(ip_address, node_type, log_dir_path):
        node_log_dir_path = os.path.join(log_dir_path, f"{node_type}_{ip_address.replace('.', '_')}")

        if not os.path.isdir(node_log_dir_path):
            os.mkdir(node_log_dir_path)
            logger.info(f"Created log dir {node_log_dir_path}")

        return node_log_dir_path

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


    def ssh_run(self, role, nr, cmd):
        ip_addrs = self.get_nodes_ipaddrs(role)
        if nr >= len(ip_addrs):
            raise ValueError(f'Node {role}-{nr} not deployed in platform')

        self.utils.ssh_run(ip_addrs[nr], cmd)

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
