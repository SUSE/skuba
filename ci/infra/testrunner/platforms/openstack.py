import os
import stat

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

    def setup_cloud_provider(self):
        openstack_conf_template = ""
        for root, dirs, files in os.walk(self.conf.workspace):
            for f in files:
                if f == "openstack.conf.template":
                    openstack_conf_template = os.path.join(root, f)

        if not openstack_conf_template:
            raise ValueError("openstack.conf.template file is not found")

        openstack_conf = os.path.join(os.path.dirname(openstack_conf_template), "openstack.conf")
        openrc_vars = {}
        openrc_vars["PRIVATE_SUBNET_ID"] = ""
        openrc_vars["PUBLIC_NET_ID"] = ""
        with open(self.conf.openstack.openrc, mode="rt") as openrc:
            for line in openrc:
                try:
                    name, _, var = line.split()[1].partition("=")
                    if var:
                        openrc_vars[name] = var.strip('"').strip("'")
                except:
                    pass

        with open(openstack_conf_template, mode="rt") as tmp_conf:
            with open(openstack_conf, mode="wt") as conf:
                for line in tmp_conf:
                    line = self._replace_env_vars(line, openrc_vars)
                    conf.write(line)
        os.chmod(openstack_conf, stat.S_IRUSR | stat.S_IWUSR)


    def _replace_env_vars(self, line, openrc_vars):
        """
        openstack.conf.template provides auth-url=<OS_AUTH_URL>.  This method will replace <OS_AUTH_URL> to
        OS_AUTH_URL from openrc file from users
         :param line:
        :param openrc:
        :return:
        """
        name, _, var = line.strip().partition("=")
        var = var.strip().strip("<").strip(">")
        re_line = ""
        if var in openrc_vars:
            line = "=".join((name, openrc_vars[var])) + os.linesep
        return line