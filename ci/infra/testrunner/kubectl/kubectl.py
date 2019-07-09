import os

from platforms.platform import Platform
from utils.format import Format
from utils.utils import (step, Utils)

class Kubectl:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.platform = Platform.get_platform(conf, platform)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)


    def get_pods(self):
        print("gets pods using kubectl")
        try:
            self._run_kubectl("get po -o wide")
        except Exception as ex:
            print("Received the following error {}".format(ex))


    def create_deployment(self, name, image):
        print("create a new deployment {}".format(name))
        try:
            self._run_kubectl("create deployment {name}  --image={image}"
                             .format(name=name, image=image))
        except Exception as ex:
            print("Received the following error {}".format(ex))


    def scale_deployment(self, name, replicas):
        print("scale deployment {}".format(name))
        try:
            self._run_kubectl("scale deployment {name} --replicas={replicas}"
                             .format(name=name, replicas=replicas))
        except Exception as ex:
            print("Received the following error {}".format(ex))


    def expose_deployment(self, name, port, nodeType="NodePort"):
        print("expose deployment {}".format(name))
        try:
            self._run_kubectl("expose deployment {name} --port={port} --type={nodeType}"
                             .format(name=name, port=port, nodeType=nodeType))
        except Exception as ex:
            print("Received the following error {}".format(ex))


    def wait_deployment(self, name, timeout):
        print("wait deployment {}".format(name))
        try:
            self._run_kubectl("wait --for=condition=available deploy/{name} --timeout={timeout}m"
                             .format(name=name, timeout=timeout))
        except Exception as ex:
            print("Received the following error {}".format(ex))


    def count_available_replicas(self, name):
        print("count available replicas of deployment {}".format(name))
        try:
            result = self._run_kubectl("get deployment/{name} | jq '.status.availableReplicas'"
                                       .format(name=name))
            return int(result)
        except Exception as ex:
            print("Received the following error {}".format(ex))
            return 0


    def get_service_port(self, name):
        print("get port for deployment {}".format(name))
        try:
            result = self._run_kubectl("get service/{name} | jq '.spec.ports[0].nodePort'"
                                       .format(name=name))
            return int(result)
        except Exception as ex:
            print("Received the following error {}".format(ex))
            return 0

    def test_service(self, name, path="/"):
        print("curl deployment {}".format(name))
        ip_address = self.platform.get_nodes_ipaddrs("worker")
        port = self.get_service_port(name)
        try:
            return self.utils.runshellcommand_withoutput("curl {ip}:{port}{path}"
                                                    .format(ip=ip_address[0],port=port,path=path), False)
        except Exception as ex:
            print("Received the following error {}".format(ex))
            return ""


    def _run_kubectl(self, command):
        try:
            return self.utils.runshellcommand_withoutput("kubectl --kubeconfig={cwd}/test-cluster/admin.conf -o json {command}"
                                                  .format(command=command,cwd=self.conf.workspace), False)
        except Exception as ex:
            print("Received the following error {}".format(ex))
            return ""