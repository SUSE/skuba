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

    @staticmethod
    def get_pods(conf):
        print("gets pods using kubectl")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl get po --kubeconfig={cwd}/test-cluster/admin.conf -o wide".format(cwd=conf.workspace))
        except Exception as ex:
            print("Received the following error {}".format(ex))

    @staticmethod
    def create_deployment(conf, name, image):
        print("create a new deployment")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl create deployment {name} --kubeconfig={cwd}/test-cluster/admin.conf --image={image}".format(cwd=conf.workspace, name=name, image=image))
        except Exception as ex:
            print("Received the following error {}".format(ex))

    @staticmethod
    def scale_deployment(conf, name, replicas):
        print("scale a deployment")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl scale deployment {name} --kubeconfig={cwd}/test-cluster/admin.conf --replicas={replicas}".format(cwd=conf.workspace, name=name, replicas=replicas))
        except Exception as ex:
            print("Received the following error {}".format(ex))

    @staticmethod
    def expose_deployment(conf, name, port, nodeType="NodePort"):
        print("expose a deployment")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl expose deployment {name} --kubeconfig={cwd}/test-cluster/admin.conf --port={port} --type={nodeType}".format(cwd=conf.workspace, name=name, port=port, nodeType=nodeType))
        except Exception as ex:
            print("Received the following error {}".format(ex))

    @staticmethod
    def wait_deployment(conf, name, timeout):
        print("wait a deployment")
        utils = Utils(conf)
        try:
            utils.runshellcommand("kubectl wait --for=condition=available deploy/{name} --kubeconfig={cwd}/test-cluster/admin.conf --timeout={timeout}m".format(cwd=conf.workspace, name=name, timeout=timeout))
        except Exception as ex:
            print("Received the following error {}".format(ex))

    @staticmethod
    def count_available_replicas(conf, name):
        print("count available replicas of a deployment")
        utils = Utils(conf)
        try:
            return int(utils.runshellcommand_withoutput("kubectl get deployment/{name} -o json --kubeconfig={cwd}/test-cluster/admin.conf | jq '.status.availableReplicas'".format(cwd=conf.workspace, name=name), False))
        except Exception as ex:
            print("Received the following error {}".format(ex))
            return 0
