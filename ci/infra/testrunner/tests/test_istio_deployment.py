import logging
import pytest
import requests
import time


GATEWAY_HTTPBIN = ("""
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: httpbin-gateway
  namespace: default
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - '*'
    port:
      name: http
      number: 80
      protocol: HTTP
EOF
""")

VIRTUALSERVICE_HTTPBIN = ("""
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: httpbin
spec:
  hosts:
  - '*'
  gateways:
  - httpbin-gateway
  http:
  - match:
    - uri:
        prefix: /status
    - uri:
        prefix: /delay
    route:
    - destination:
        port:
          number: 8000
        host: httpbin
EOF
""")

def istio_httpbin_setup(kubectl):
    istioctl = ("""
                istioctl --kubeconfig={config} manifest apply \
                         --set profile=default \
                         --set addonComponents.prometheus.enabled=false \
                         --set hub=registry.suse.de/devel/caasp/5/containers/containers/caasp/v5 \
                         --set tag=1.5.4 \
                         --set values.pilot.image=istio-pilot \
                         --set values.global.proxy.image=istio-proxyv2 \
                         --set values.global.proxy_init.image=istio-proxyv2
                 """.format(config=kubectl.get_kubeconfig()))

    kubectl.utils.runshellcommand(istioctl)
    kubectl.run_kubectl("-n istio-system wait --for=condition=available deploy/istio-ingressgateway --timeout=3m")

    kubectl.run_kubectl("create -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")


def cleanup(kubectl):
    kubectl.run_kubectl("delete -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")
    istioctl_delete = ("""
                       istioctl --kubeconfig={config} manifest generate \
                                --set profile=default \
                                --set addonComponents.prometheus.enabled=false \
                                --set hub=registry.suse.de/devel/caasp/5/containers/containers/caasp/v5 \
                                --set tag=1.5.4 \
                                --set values.pilot.image=istio-pilot \
                                --set values.global.proxy.image=istio-proxyv2 \
                                --set values.global.proxy_init.image=istio-proxyv2 \
                                | kubectl --kubeconfig={config} delete -f - || true
                        """.format(config=kubectl.get_kubeconfig()))
    kubectl.utils.runshellcommand(istioctl_delete)


def test_istio_deployment(deployment, platform, skuba, kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Deploying istio and httpbin")
    istio_httpbin_setup(kubectl)

    logger.info("Create the istio config")
    kubectl.run_kubectl("apply -f - << EOF " + GATEWAY_HTTPBIN)
    kubectl.run_kubectl("apply -f - << EOF " + VIRTUALSERVICE_HTTPBIN)

    # Wait for istio to digest the config
    time.sleep(100)

    nodePort = kubectl.run_kubectl("-n istio-system get service/istio-ingressgateway -o jsonpath='{ .spec.ports[1].nodePort }'")

    assert 30000 <= int(nodePort) <= 32767

    wrk_idx = 0
    ip_addresses = platform.get_nodes_ipaddrs("worker")

    assert "10." in ip_addresses[wrk_idx]

    url = "{protocol}://{ip}:{port}{path}".format(protocol="http", ip=str(ip_addresses[wrk_idx]), port=str(nodePort), path="/status/200")
    r = requests.get(url, headers={'host': 'httpbin.example.com'})

    assert 200 == r.status_code

    cleanup(kubectl)

