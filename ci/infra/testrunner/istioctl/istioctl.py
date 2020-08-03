from utils.utils import (Utils)
from time import sleep
import logging

class Istioctl:

    def __init__(self, conf):
        self.conf = conf
        self.binpath = conf.istioctl.binpath
        self.istio_version = conf.istioctl.istio_version
        self.utils = Utils(self.conf)
        self.container_hub = conf.istioctl.container_hub
        self.logger = logging.getLogger("testrunner")

    def istio_setup(self, kubectl):
        # Enable injection only for default namespace here. 
        # Tests will need explicitly enabling injection if different namespace
        # is used (for e.g: TCP Traffic Shifting Test)
        kubectl.run_kubectl("label --overwrite=true namespace default istio-injection=enabled")

        istioctl = ("""
                istioctl --kubeconfig={config} manifest apply \
                         --set profile=default \
                         --set addonComponents.prometheus.enabled=false \
                         --set hub={hub} \
                         --set tag={tag} \
                         --set values.pilot.image=istio-pilot \
                         --set values.global.proxy.image=istio-proxyv2 \
                         --set values.global.proxy_init.image=istio-proxyv2
                 """.format(config=kubectl.get_kubeconfig(),
                            hub=self.container_hub,
                            tag=self.istio_version))

        self.logger.info("Running command: + " + istioctl)
        kubectl.utils.runshellcommand(istioctl)

        kubectl.run_kubectl("-n istio-system wait --for=condition=available deploy/istio-ingressgateway --timeout=3m")


    def istio_cleanup(self, kubectl):
        kubectl.run_kubectl("delete --ignore-nod-found=true -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")
        istioctl_delete = ("""
                       istioctl --kubeconfig={config} manifest generate \
                                --set profile=default \
                                --set addonComponents.prometheus.enabled=false \
                                --set hub={hub} \
                                --set tag={tag} \
                                --set values.pilot.image=istio-pilot \
                                --set values.global.proxy.image=istio-proxyv2 \
                                --set values.global.proxy_init.image=istio-proxyv2 \
                                | kubectl --kubeconfig={config} delete --ignore-not-found=true -f - || true
                        """.format(config=kubectl.get_kubeconfig(),
                                   hub=self.container_hub,
                                   tag=self.istio_version))
        
        self.logger.info("Running command: + " + istioctl_delete)
        kubectl.utils.runshellcommand(istioctl_delete)


    def get_ingress_address(self, kubectl):
        wrk_idx = 0
        ip_addresses = platform.get_nodes_ipaddrs("worker")
        assert 0 < len(ip_addresses)

        worker_ip = ip_addresses[wrk_idx]

        port = kubectl.run_kubectl("-n istio-system get service/istio-ingressgateway -o jsonpath='{ .spec.ports[2].nodePort }'")
        assert 30000 <= int(port) <= 32767

        self.logger.info("Obtain Ingress Path: %s:%s" % (worker_ip, port))
        return "%s:%s" % (worker_ip, port)
