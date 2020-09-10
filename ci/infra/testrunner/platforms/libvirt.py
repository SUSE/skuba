from timeout_decorator import timeout
from platforms.terraform import Terraform


class Libvirt(Terraform):
    def __init__(self, conf):
        super().__init__(conf, 'libvirt')
        self.platform_new_vars = {
            "libvirt_uri": self.conf.libvirt.uri,
            "libvirt_keyfile": self.conf.libvirt.keyfile,
        }

    def _env_setup_cmd(self):
        return ":"

    @timeout(600)
    def _cleanup_platform(self):
        self.destroy()
