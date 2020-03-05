import os

from timeout_decorator import timeout

from platforms.terraform import Terraform
from utils import Format


class VMware(Terraform):
    def __init__(self, conf):
        super().__init__(conf, 'vmware')
        if not os.path.isfile(conf.vmware.env_file):
            msg = (f'Your VMware env file path "{conf.vmware.env_file}" does not exist.\n\t    '
                   'Check the VMware env file path in your configured yaml file.')
            raise ValueError(Format.alert(msg))
        self.platform_new_vars = {}

    def _env_setup_cmd(self):
        return f"source {self.conf.vmware.env_file}"

    @timeout(600)
    def _cleanup_platform(self):
        variables = [f"stack_name={self.conf.terraform.stack_name}"]

        self.destroy(variables)

    def _get_platform_logs(self):
        # Get logs from the VMware LB
        node_ip = self.get_lb_ipaddr()
        logs = {
            "files": ["/var/run/cloud-init/status.json",
                      "/var/log/cloud-init-output.log",
                      "/var/log/cloud-init.log"],
            "dirs": [],
            "services": ["haproxy"]
        }

        node_log_dir = self._create_node_log_dir(node_ip, "load_balancer", self.conf.log_dir)
        logging_error = self.utils.collect_remote_logs(node_ip, logs, node_log_dir)

        return logging_error

    def get_lb_ipaddr(self):
        self._load_tfstate()
        if self.state["version"] == 3:
            return self.state["modules"][0]["outputs"]["ip_load_balancer"]["value"]["{}-lb-0".format(self.stack_name())]
        elif self.state["version"] == 4:
            return self.state["outputs"]["ip_load_balancer"]["value"]["{}-lb-0".format(self.stack_name())]
