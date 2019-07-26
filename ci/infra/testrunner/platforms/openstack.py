import os
from timeout_decorator import timeout

from platforms.terraform import Terraform
from utils import Format


class Openstack(Terraform):
    def __init__(self, conf):
        super().__init__(conf, 'openstack')
        if not os.path.isfile(conf.openstack.openrc):
            raise ValueError(Format.alert(f"Your openrc file path \"{conf.openstack.openrc}\" does not exist.\n\t    "
                                          "Check your openrc file path in a configured yaml file"))

    def _env_setup_cmd(self):
        return f"source {self.conf.openstack.openrc}"

    @timeout(600)
    def _cleanup_platform(self):
        variables = [f"internal_net=net-{self.conf.terraform.internal_net}",
                     f"stack_name={self.conf.terraform.stack_name}"]

        self.destroy(variables)
