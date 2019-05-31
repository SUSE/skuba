import json
import os
import subprocess
from shutil import copyfile

from utils import (Constant, Format, step, Utils)


class Terraform:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(conf)
        self.tfdir = os.path.join(self.conf.terraform.tfdir,self.conf.platform)
        self.tfjson_path = os.path.join(conf.workspace, "tfout.json")

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
    def apply_terraform(self):
        """ Create and apply terraform plan"""
        print("Init terraform")
        self._runshellcommandterraform("terraform init")
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
        fn = os.path.join(self.tfdir, "terraform.tfstate")
        print("Reading {}".format(fn))
        with open(fn) as f:
            return json.load(f)

    def get_lb_ipaddr(self):
        self.state = self._load_tfstate()
        return self.state["modules"][0]["outputs"]["ip_ext_load_balancer"]["value"]

    def get_masters_ipaddrs(self):
        self.state = self._load_tfstate()
        return self.state["modules"][0]["outputs"]["ip_masters"]["value"]

    def get_workers_ipaddrs(self):
        self.state = self._load_tfstate()
        return self.state["modules"][0]["outputs"]["ip_workers"]["value"]

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
        src_terraform = os.path.join(
                            self.tfdir,
                            self.conf.terraform.tfvars)

        dir, tfvars, _ = src_terraform.partition("terraform.tfvars")
        dest_terraform = os.path.join(dir, tfvars)
        copyfile(src_terraform, dest_terraform)

        with open(dest_terraform) as f:
            lines = f.readlines()

        for i, line in enumerate(lines):
            # TODO: internal_net and stack_name are openstack variables
            #       should move to the Openstack class
            if line.startswith("internal_net"):
                lines[i] = 'internal_net = "{}"'.format(self.conf.jenkins.run_name)

            if line.startswith("stack_name"):
                lines[i] = 'stack_name = "{}"'.format(self.conf.jenkins.run_name)

            if line.startswith("username"):
                lines[i] = 'username = "{}"'.format(self.conf.nodeuser)

            if line.startswith("masters"):
                lines[i] = 'masters = {}'.format(self.conf.master.count)

            if line.startswith("workers"):
                lines[i] = 'workers = {}'.format(self.conf.worker.count)

            if line.startswith("authorized_keys"):
                lines[i] = 'authorized_keys = [ "{}" ,'.format(self.utils.authorized_keys())

            # Switch to US mirror if running on CI
            if "download.suse.de" in line and os.environ.get('JENKINS_URL'):
                lines[i]=line.replace('download.suse.de', 'ibs-mirror.prv.suse.net')

        with open(dest_terraform, "w") as f:
            f.writelines(lines)

    def _runshellcommandterraform(self, cmd, env=None):
        """Running terraform command in {terraform.tfdir}/{platform}"""
        cwd = self.tfdir
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
