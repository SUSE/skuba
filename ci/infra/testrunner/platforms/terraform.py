import os
from shutil import copyfile
import subprocess
from utils import step
from utils import Utils
from constants import Constant


class Terraform:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(conf)

    def _env_setup_cmd(self):
        """Returns the command for setting up the platform environment"""
        return ""

    def _cleanup_platform(self):
        """Platform specific cleanup. Expected to be overriden by platforms"""

    def cleanup(self):
        """ Clean up """
        try:
            self._cleanup_platform()
        except Exception as ex:
            cleanup_failure = True
            print("Received the following error {}".format(ex))
            print("Attempting to finish cleaup")

        dirs = [os.path.join(self.conf.workspace, "tfout"),
                os.path.join(self.conf.workspace, "tfout.json")]

        for dir in dirs:
            try:
                self.utils.runshellcommand("rm -rf {}".format(dir))
            except Exception as ex:
                cleanup_failure = True
                print("Received the following error {}".format(ex))
                print("Attempting to finish cleaup")

        if cleanup_failure:
            raise Exception("Failure(s) during cleanup")

    @step
    def apply_terraform(self):
        """ Create and apply terraform plan"""
        print("Init terraform")
        self.runshellcommandterraform("terraform init")
        self.runshellcommandterraform("terraform version")
        self.generate_tfvars_file()
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
            print("Run terraform plan - execution n. %d" % retry)
            self.runshellcommandterraform(plan_cmd)
            print("Running terraform apply - execution n. %d" % retry)
            try:
                self.runshellcommandterraform(apply_cmd)
                break

            except:
                print("Failed terraform apply n. %d" % retry)
                if retry == 4:
                    print("Last failed attempt, exiting")
                    raise Exception("Failed Terraform deployment")

            self.fetch_terraform_output()

    @step
    def fetch_terraform_output(self):
        cmd = ("{env_setup};"
               "terraform output -json >"
               "{workspace}/tfout.json".format(
                   env_setup=self._env_setup_cmd(),
                   workspace=self.conf.workspace))
        self.runshellcommandterraform(cmd)

    def generate_tfvars_file(self):
        """Generate terraform tfvars file"""
        src_terraform = os.path.join(
                            self.conf.workspace,
                            "caaspctl/ci/infra/{}/{}".format(
                                self.conf.platform,
                                Constant.TERRAFORM_EXAMPLE)
                        )

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

        with open(dest_terraform, "w") as f:
            f.writelines(lines)

    def runshellcommandterraform(self, cmd, env=None):
        """Running terraform command in {workspace}/ci/infra/{platform}"""
        cwd = self.conf.terraform_dir
        print("$ {} > {}".format(cwd, cmd))
        subprocess.check_call(cmd, cwd=cwd, shell=True, env=env)
