import os

from timeout_decorator import timeout

from platforms.platform import Platform
from utils.format import Format
from utils.utils import (step, Utils)

class Kubectl:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.platform = Platform.get_platform(conf, platform)


    def create_deployment(self, name, image):
        try:
            self._run_kubectl("create deployment {name}  --image={image}"
                                .format(name=name, image=image))
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex


    def scale_deployment(self, name, replicas):
        try:
            self._run_kubectl("scale deployment {name} --replicas={replicas}"
                             .format(name=name, replicas=replicas))
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex


    def expose_deployment(self, name, port, nodeType="NodePort"):
        try:
            self._run_kubectl("expose deployment {name} --port={port} --type={nodeType}"
                             .format(name=name, port=port, nodeType=nodeType))
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex


    def wait_deployment(self, name, timeout):
        try:
            self._run_kubectl("wait --for=condition=available deploy/{name} --timeout={timeout}m"
                             .format(name=name, timeout=timeout))
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex


    def count_available_replicas(self, name):
        try:
            result = self._run_kubectl("get deployment/{name} | jq '.status.availableReplicas'"
                                       .format(name=name))
            return int(result)
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex


    def get_service_port(self, name):
        try:
            result = self._run_kubectl("get service/{name} | jq '.spec.ports[0].nodePort'"
                                       .format(name=name))
            return int(result)
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex

    def test_service(self, name, path="/", worker=0):
        ip_addresses = self.platform.get_nodes_ipaddrs("worker")
        if worker >= len(ip_address):
            raise ValueError("Node worker-{} not deployed".format(worker))

        port = self.get_service_port(name)
        shell_cmd = "curl {ip}:{port}{path}".format(ip=ip_address[worker],port=port,path=path)
        try:
            return self.utils.runshellcommand(shell_cmd, output=True)
        except Exception as ex:
            raise Exception("Error testing service {} with path {} at node {}"
                                .format(name, path, ip_address)) from ex
 
    def _run_kubectl(self, command):
        shell_cmd = "kubectl --kubeconfig={cwd}/test-cluster/admin.conf \
                      -o json {command}".format(command=command, cwd=self.conf.workspace)
        try:
            return self.utils.runshellcommand(shell_cmd, output=True)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(shell_cmd)) from ex

