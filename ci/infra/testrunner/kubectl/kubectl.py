import os

from platforms.platform import Platform
from utils.format import Format
from utils.utils import (step, Utils)

class Kubectl:

    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.platform = Platform.get_platform(conf)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)

    @staticmethod
    def get_nodes(conf):
        print("gets nodes using kubectl")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl get nodes --kubeconfig={cwd}/test-cluster/admin.conf".format(cwd=conf.workspace))
        except Exception as ex:
            print("Received the following error {}".format(ex))
