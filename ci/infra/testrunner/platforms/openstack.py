import os
from timeout_decorator import timeout

from platforms.terraform import Terraform
from utils import Format


class Openstack(Terraform):
    def __init__(self, conf):
        super().__init__(conf)
        if  not os.path.isfile(conf.openstack.openrc):
            raise ValueError(Format.alert("Your openrc file path \"{}\" does not exist.\n\t    "
                                 "Check your openrc file path in a configured yaml file".format(conf.openstack.openrc)))
        self.osconf = conf.openstack

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
