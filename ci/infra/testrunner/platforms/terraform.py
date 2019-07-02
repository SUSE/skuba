import hcl
import json
import os
import re
import subprocess
from shutil import copyfile

from utils import (Constant, Format, step, Utils)


class Terraform:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(conf)
        self.tfdir = os.path.join(self.conf.terraform.tfdir,self.conf.platform)
        self.tfjson_path = os.path.join(conf.workspace, "tfout.json")
        self.state = None

    def _env_setup_cmd(self):
        """Returns the command for setting up the platform environment"""
        return ""

    def _cleanup_platform(self):
        """Platform specific cleanup. Expected to be overriden by platforms"""

    def cleanup(self):
        """ Clean up """
        cleanup_failure = False
        try:
            self._cleanup_platform()
        except Exception as ex:
            cleanup_failure = True
            print(Format.alert("Received the following error {}".format(ex)))
            print("Attempting to finish cleaup")

        dirs = [os.path.join(self.conf.workspace, "tfout"),
                self.tfjson_path]

        for dir in dirs:
            try:
                self.utils.runshellcommand("rm -rf {}".format(dir))
            except Exception as ex:
                cleanup_failure = True
                print("Received the following error {}".format(ex))
                print("Attempting to finish cleaup")

        if cleanup_failure:
            raise Exception(Format.alert("Failure(s) during cleanup"))

    @step
    def provision(self, num_master=-1, num_worker=-1):
        """ Create and apply terraform plan"""
        if num_master > -1 or num_worker > -1:
            print("Overriding number of nodes")
            if num_master > -1:
                self.conf.master.count = num_master
                print("   Masters:{} ".format(num_master));

            if num_worker > -1:
                self.conf.worker.count = num_worker
                print("   Workers:{} ".format(num_worker));

        print("Init terraform")
        self._check_tf_deployed()
        
        self.utils.setup_ssh()

        init_cmd = "terraform init"
        if self.conf.terraform.plugin_dir:
            print("Installing plugins from {}".format(self.conf.terraform.plugin_dir))
            init_cmd = init_cmd+" -plugin-dir="+self.conf.terraform.plugin_dir
        self._runshellcommandterraform(init_cmd)

        self._runshellcommandterraform("terraform version")
        self._generate_tfvars_file()
        plan_cmd = ("{env_setup};"
                    " terraform plan "
                    " -out {workspace}/tfout".format(
                        env_setup=self._env_setup_cmd(),
                        workspace=self.conf.workspace))
        apply_cmd = ("{env_setup};"
                     "terraform apply -auto-approve {workspace}/tfout".format(
                        env_setup=self._env_setup_cmd(),
                        workspace=self.conf.workspace))

        # TODO: define the number of retries as a configuration parameter
        for retry in range(1, 5):
            print(Format.alert("Run terraform plan - execution # {}".format(retry)))
            self._runshellcommandterraform(plan_cmd)
            print(Format.alert("Run terraform apply - execution # {}".format(retry)))
            try:
                self._runshellcommandterraform(apply_cmd)
                break

            except:
                print("Failed terraform apply n. %d" % retry)
                if retry == 4:
                    print(Format.alert("Failed Openstack Terraform deployment"))
                    raise
            finally:
                self._fetch_terraform_output()

    def _load_tfstate(self):
        if self.state is None:
            fn = os.path.join(self.tfdir, "terraform.tfstate")
            print("Reading {}".format(fn))
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
        cmd = ("{env_setup};"
               "terraform output -json >"
               "{json_f}".format(
                   env_setup=self._env_setup_cmd(),
                   json_f=self.tfjson_path))
        self._runshellcommandterraform(cmd)

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

        # Switch to mirrors  
        repos = tfvars.get("repositories")
        if self.conf.mirror and repos is not None:
            for name, url in repos.items():
                tfvars["repositories"][name] = url.replace("download.suse.de", self.conf.mirror)

    def _runshellcommandterraform(self, cmd, env={}):
        """Running terraform command in {terraform.tfdir}/{platform}"""
        cwd = self.tfdir

        # Terraform needs PATH and SSH_AUTH_SOCK
        sock_fn = self.utils.ssh_sock_fn()
        env["SSH_AUTH_SOCK"] = sock_fn
        env["PATH"] = os.environ['PATH']

        print(Format.alert("$ {} > {}".format(cwd, cmd)))
        subprocess.check_call(cmd, cwd=cwd, shell=True, env=env)

    def _check_tf_deployed(self):
        if os.path.exists(self.tfjson_path):
            raise Exception(Format.alert("tf file found. Please run cleanup and try again{}"))

    # TODO: this function is currently not used. Identify points where it should
    # be invoked
    def _verify_tf_dependency(self):
        if not os.path.exists(self.tfjson_path):
            raise Exception(Format.alert("tf file not found. Please run terraform and try again{}"))
