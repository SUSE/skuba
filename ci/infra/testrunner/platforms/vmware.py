import os

from timeout_decorator import timeout

from platforms.terraform import Terraform
from utils import Format


class VMware(Terraform):
    def __init__(self, conf):
        super().__init__(conf)
        if not os.path.isfile(conf.vmware.env_file):
            msg = (f'Your VMware env file path "{conf.vmware.env_file}" does not exist.\n\t    '
                   'Check the VMware env file path in your configured yaml file.')
            raise ValueError(Format.alert(msg))

    def _env_setup_cmd(self):
        return f"source {self.conf.vmware.env_file}"

    @timeout(600)
    def _cleanup_platform(self):
        cmd = (f"source {self.conf.vmware.env_file}; "
               f"terraform destroy -auto-approve -var stack_name={self.conf.jenkins.run_name}")

        self._runshellcommandterraform(cmd)
