import os, sys, pprint
from shutil import copyfile
from timeout_decorator import timeout
from utils import step
from caaspctl import Caaspctl
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
        try:
            self._cleanup_openstack_deployment()
        except:
            pass

        dirs =[os.path.join(self.conf.workspace, "test-cluster"),
               os.path.join(self.conf.workspace, "go"),
               os.path.join(self.conf.workspace, "logs"),
               os.path.join(self.conf.workspace, "ssh-agent-sock"),
               os.path.join(self.conf.workspace, "test-cluster"),
               os.path.join(self.conf.workspace, "tfout"),
               os.path.join(self.conf.workspace, "tfout.json"),
               os.path.join(self.conf.terraform_dir, "terraform.tfstate")]

        for dir in dirs:
            self.utils.runshellcommand("rm -rf {}".format(dir))


    @step
    def apply_terraform(self):
        print("Init terraform")
        self.utils.runshellcommandterraform("terraform init")
        self.utils.runshellcommandterraform("terraform version")
        self.generate_tfvars_file()
        plan_cmd = ("source {openrc};"
                    " terraform plan "
                    " -out {tfout}".format(openrc=self.conf.openstack.openrc,
                                           tfout=os.path.join(self.conf.workspace,"tfout"))
                    )
        apply_cmd = ("source {openrc};"
                     "terraform apply -auto-approve"
                     " {tfout}".format(openrc=self.conf.openstack.openrc,
                                            tfout=os.path.join(self.conf.workspace,"tfout"))
                    )

        for retry in range(1, 5):
            print("{}Run terraform plan - execution n {}".format(Constant.BLUE, retry, Constant.COLOR_EXIT))
            self.utils.runshellcommandterraform(plan_cmd)
            try:
                print("{}Run terraform apply - execution n {}".format(Constant.BLUE, retry, Constant.COLOR_EXIT))
                self.utils.runshellcommandterraform(apply_cmd)
                break

            except:
                print("{}Failed terraform apply -n {}".format(Constant.BLUE, retry, Constant.COLOR_EXIT))
                if retry == 4:
                    self._cleanup_openstack_deployment()
                    raise RuntimeError("{}{}{}".format(Constant.RED, "Failed Openstack Terraform deployment "
                                                       "and destroyed associated resources", Constant.COLOR_EXIT))

            self.fetch_openstack_terraform_output()


    @step
    def fetch_openstack_terraform_output(self):
        cmd = "source {}; terraform output -json > {}/tfout.json".format(
                                self.conf.openstack.openrc, self.conf.workspace)
        self.utils.runshellcommandterraform(cmd)


    def generate_tfvars_file(self):
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

