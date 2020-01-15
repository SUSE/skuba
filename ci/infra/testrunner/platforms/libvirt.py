import os
import stat

from timeout_decorator import timeout
from platforms.terraform import Terraform
from utils import Format


class Libvirt(Terraform):
    def __init__(self, conf):
        super().__init__(conf, 'libvirt')

    def _env_setup_cmd(self):
        return ":"

    @timeout(600)
    def _cleanup_platform(self):
        self.destroy()

    def get_lb_ipaddr(self):
        self._load_tfstate()
        return self.state["modules"][0]["outputs"]["ip_load_balancer"]["value"]["{}-lb".format(self.conf.terraform.stack_name)]

    def get_nodes_ipaddrs(self, role):
        self._load_tfstate()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role: {}".format(role))
        role_key = "ip_" + role + "s"

        return list(self.state["modules"][0]["outputs"][role_key]["value"].values())
