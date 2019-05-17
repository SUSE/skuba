import os, sys, pprint
from shutil import copyfile
from timeout_decorator import timeout
from utils import step
from utils import Utils
from constants import Constant


class Openstack:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(conf)


    @timeout(600)
    def _cleanup_openstack_deployment(self):
        cmd = 'mkdir -p {}/logs'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        cmd = 'chmod a+x {}'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        cmd = ("source {orc};"
            " terraform destroy -auto-approve"
            " -var internal_net=net-{run}"
            " -var stack_name={run}").format(orc=self.conf.openstack.openrc, run=self.conf.jenkins.run_name)
        self.utils.runshellcommandterraform(cmd)


    def cleanup(self):
        cleanup_failure = False

        try:
            self._cleanup_openstack_deployment()
        except Exception as ex:
            cleanup_failure = True
            print("Received the following error {}".format(ex))
            print("Attempting to finish cleanup")

        workspace_dirs = [os.path.join(self.conf.workspace, "test-cluster"),
                          os.path.join(self.conf.workspace, "go"),
                          os.path.join(self.conf.workspace, "logs"),
                          os.path.join(self.conf.workspace, "ssh-agent-sock"),
                          os.path.join(self.conf.workspace, "test-cluster"),
                          os.path.join(self.conf.workspace, "tfout"),
                          os.path.join(self.conf.workspace, "tfout.json"),
                          os.path.join(self.conf.terraform_dir, "terraform.tfstate")]

        for workspace_dir in workspace_dirs:
            try:
                self.utils.runshellcommand("rm -rf {}".format(workspace_dir))
            except Exception as ex:
                cleanup_failure = True
                print("Received the following error {}".format(ex))
                print("Attempting to finish cleanup")

        if cleanup_failure:
            raise Exception("Failure(s) during cleanup")


    @step
    def apply_terraform(self):
        print("Init terraform")
        self._check_tf_deployed()
        self.utils.runshellcommandterraform("terraform init")
        self.utils.runshellcommandterraform("terraform version")
        self._generate_tfvars_file()
        plan_cmd = ("source {openrc};"
                    " terraform plan "
                    " -out {tfout}".format(openrc=self.conf.openstack.openrc,
                                           tfout=os.path.join(self.conf.workspace, "tfout"))
                    )
        apply_cmd = ("source {openrc};"
                     "terraform apply -auto-approve"
                     " {tfout}".format(openrc=self.conf.openstack.openrc,
                                       tfout=os.path.join(self.conf.workspace, "tfout"))
                     )
        for retry in range(1, 5):
            print("{}Run terraform plan - execution # {}".format(Constant.BLUE, retry, Constant.COLOR_EXIT))
            self.utils.runshellcommandterraform(plan_cmd)
            try:
                print("{}Run terraform apply - execution # {}".format(Constant.BLUE, retry, Constant.COLOR_EXIT))
                self.utils.runshellcommandterraform(apply_cmd)
                break

            except:
                if retry == 4:
                    raise RuntimeError("{}{}{}".format(Constant.RED, "Failed Openstack Terraform deployment "
                                                       "and destroyed associated resources", Constant.COLOR_EXIT))

        self._fetch_openstack_terraform_output()


    @step
    def _fetch_openstack_terraform_output(self):
        cmd = "source {orc}; terraform output -json > {json_f}".format(
                                orc=self.conf.openstack.openrc, json_f=os.path.join(self.conf.workspace,"tfout.json"))
        self.utils.runshellcommandterraform(cmd)


    def _generate_tfvars_file(self):
        """Generate terraform tfvars file"""
        src_terraform = os.path.join(self.conf.workspace,
                            "caaspctl/ci/infra/{}/{}".format(
                                self.conf.platform, Constant.TERRAFORM_EXAMPLE))

        dir, tfvars, _ = src_terraform.partition("terraform.tfvars")
        dest_terraform = os.path.join(dir, tfvars)
        copyfile(src_terraform, dest_terraform)

        with open(dest_terraform) as f:
            lines = f.readlines()

        for i, line in enumerate(lines):
            if line.startswith("internal_net"):
                lines[i] = 'internal_net = "{}"'.format(self.conf.jenkins.run_name)

            if line.startswith("stack_name"):
                    lines[i] = 'stack_name = "{}"'.format(self.conf.jenkins.run_name)

            if line.startswith("username"):
                lines[i]='username = "{}"'.format(self.conf.nodeuser)

            if line.startswith("masters"):
                lines[i]='masters = {}'.format(self.conf.master.count)

            if line.startswith("workers"):
                lines[i]='workers = {}'.format(self.conf.worker.count)

            if line.startswith("authorized_keys"):
                lines[i]='authorized_keys = [ "{}" ,'.format(self.utils.authorized_keys())

        with open(dest_terraform, "w") as f:
            f.writelines(lines)

    def _check_tf_deployed(self):
        if os.path.exists(os.path.join(os.path.join(self.conf.workspace, "tfout.json"))):
            raise RuntimeError("{}You need to run \"testrunner --cleanup first"
                               " before running \"testrunner --terraform\" commands\"{}".format(Constant.RED, Constant.COLOR_EXIT))