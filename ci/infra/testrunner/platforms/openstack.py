from timeout_decorator import timeout
from terraform import Terraform


class Openstack(Terraform):
    def __init__(self, conf):
        self.osconf = conf.openstack
        super().__init__(conf)

    def _env_setup_cmd(self):
        return "source {openrc}".format(openrc=self.osconf.openrc)

    @timeout(600)
    def _cleanup_platform(self):
        # TODO: this command is here because is passes two openstack
        # specific vars to terraform. Find a way to move the command to 
        # Terraform class and pass the variables from Openstack class.
        cmd = ("source {openrc};"
               " terraform destroy -auto-approve"
               " -var internal_net=net-{run}"
               " -var stack_name={run}".format(
                   openrc=self.conf.openstack.openrc,
                   run=self.conf.jenkins.run_name))

        self.runshellcommandterraform(cmd)
