import platforms

from skuba.skuba import Skuba
from utils.utils import (Utils)


class Kubectl:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.platform = platforms.get_platform(conf, platform)
        self.skuba = Skuba(conf, platform)

    def run_kubectl(self, command):
        kubeconfig = self.skuba.get_kubeconfig()

        shell_cmd = "kubectl --kubeconfig={} {}".format(kubeconfig, command)
        try:
            return self.utils.runshellcommand(shell_cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(shell_cmd)) from ex
