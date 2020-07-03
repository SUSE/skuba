from utils.utils import (Utils)
from time import sleep


class Kubectl:

    def __init__(self, conf):
        self.conf = conf
        self.binpath = conf.kubectl.binpath
        self.kubeconfig = conf.kubectl.kubeconfig
        self.utils = Utils(self.conf)

    def run_kubectl(self, command, stdin=None):
        shell_cmd = f'{self.binpath} --kubeconfig={self.kubeconfig} {command}'
        try:
            return self.utils.runshellcommand(shell_cmd, stdin=stdin)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(shell_cmd)) from ex

    def get_num_nodes_by_role(self, role):
        """ Returns the number of nodes by role (master/worker)"""
        if role not in ("master", "worker"):
            raise ValueError("Invalid role {}".format(role))

        nodes = self.get_node_names_by_role(role)
        return len(nodes)

    def get_node_names_by_role(self, role):
        """Returns a list of node names for a given role
        Uses selectors to get the nodes. Master nodes have the node-role.kubernetes.io/master="" label, while other
        nodes (workers) dont even have the label.
        """

        if role not in ("master", "worker"):
            raise ValueError("Invalid role {}".format(role))

        roles = {
            "master": "==",
            "worker": "!="
        }
        command = f"get nodes --selector=node-role.kubernetes.io/master{roles.get(role)}"" -o jsonpath='{.items[*].metadata.name}'"
        return self.run_kubectl(command).split()

    def inhibit_kured(self):
        max_attempt = 18
        current_attempt = 0
        while current_attempt <= max_attempt:
            try:
                self.run_kubectl("-n kube-system annotate ds kured weave.works/kured-node-lock='{\"nodeID\":\"manual\"}'")
                break
            except Exception:
                current_attempt += 1
                sleep(10)
