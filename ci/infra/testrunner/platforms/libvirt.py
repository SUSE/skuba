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

