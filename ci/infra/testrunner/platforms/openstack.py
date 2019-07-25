import os
from timeout_decorator import timeout

from platforms.terraform import Terraform
from utils import Format


class Openstack(Terraform):
    def __init__(self, conf):
        super().__init__(conf, 'openstack')
        if not os.path.isfile(conf.openstack.openrc):
            raise ValueError(Format.alert("Your openrc file path \"{}\" does not exist.\n\t    "
                                          "Check your openrc file path in a configured yaml file".format(conf.openstack.openrc)))
        self.osconf = conf.openstack

    def _env_setup_cmd(self):
        return "source {openrc}".format(openrc=self.osconf.openrc)

    @timeout(600)
    def _cleanup_platform(self):
        cmd = ("destroy -auto-approve"
               f" -var internal_net=net-{self.conf.terraform.internal_net}"
               f" -var stack_name={self.conf.terraform.stack_name}")

        self._run_terraform_command(cmd)
