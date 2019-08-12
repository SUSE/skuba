import os
import platforms

from timeout_decorator import timeout

from platforms.platform import Platform
from skuba.skuba import Skuba
from utils.format import Format
from utils.utils import (step, Utils)

class Kubectl:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.platform = platforms.get_platform(conf, platform)
        self.skuba = Skuba(conf, platform)


    def create_deployment(self, name, image):
        self.run_kubectl("create deployment {name} --image={image}"
                          .format(name=name, image=image))

    def scale_deployment(self, name, replicas):
        self.run_kubectl("scale deployment {name} --replicas={replicas}"
                          .format(name=name, replicas=replicas))

    def expose_deployment(self, name, port, nodeType="NodePort"):
        self.run_kubectl("expose deployment {name} --port={port} --type={nodeType}"
                          .format(name=name, port=port, nodeType=nodeType))

    def wait_deployment(self, name, timeout):
        self.run_kubectl("wait --for=condition=available deploy/{name} --timeout={timeout}m"
                          .format(name=name, timeout=timeout))

    def count_available_replicas(self, name):
        result = self.run_kubectl("get deployment/{name} | jq '.status.availableReplicas'"
                                   .format(name=name))
        return int(result)

    def get_service_port(self, name):
        result = self.run_kubectl("get service/{name} | jq '.spec.ports[0].nodePort'"
                                   .format(name=name))
        return int(result)

    def test_service(self, name, path="/", worker=0):
        ip_addresses = self.platform.get_nodes_ipaddrs("worker")
        if worker >= len(ip_addresses):
            raise ValueError("Node worker-{} not deployed".format(worker))

        port = self.get_service_port(name)
        shell_cmd = "curl {ip}:{port}{path}".format(ip=ip_addresses[worker],port=port,path=path)
        try:
            return self.utils.runshellcommand(shell_cmd)
        except Exception as ex:
            raise Exception("Error testing service {} with path {} at node {}"
                                .format(name, path, ip_address)) from ex
 
    def lock_kured(self):
        self.run_kubectl("-n kube-system annotate ds kured weave.works/kured-node-lock='{\"nodeID\":\"manual\"}'")


    def unlock_kured(self):
        self.run_kubectl("-n kube-system annotate ds kured weave.works/kured-node-lock-")


    def check_kured_lock(self):
        return self.run_kubectl("-n kube-system get ds/kured",
            outputformat="-o jsonpath='{.metadata.annotations.weave\.works/kured-node-lock}'").find("manual") != -1


    def run_kubectl(self, command):
        kubeconfig = self.skuba.get_kubeconfig()
        
        shell_cmd = "kubectl --kubeconfig={} {}".format(kubeconfig, command)
        try:
            return self.utils.runshellcommand(shell_cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(shell_cmd)) from ex
