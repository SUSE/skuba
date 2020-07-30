import logging
import pytest
import requests
import tempfile
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

GATEWAY_HTTPBIN_SECURE = ("""
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
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: httpbin-credential
    hosts:
    - 'httpbin.example.com'
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

def _istio_httpbin_setup(kubectl):
    istioctl = ("""
                istioctl --kubeconfig={config} manifest apply \
                         --set profile=default \
                         --set addonComponents.prometheus.enabled=false \
                         --set hub=registry.suse.de/devel/caasp/4.5/containers/containers/caasp/v4.5 \
                         --set tag=1.5.4 \
                         --set values.pilot.image=istio-pilot \
                         --set values.global.proxy.image=istio-proxyv2 \
                         --set values.global.proxy_init.image=istio-proxyv2
                 """.format(config=kubectl.get_kubeconfig()))

    kubectl.utils.runshellcommand(istioctl)
    kubectl.run_kubectl("-n istio-system wait --for=condition=available deploy/istio-ingressgateway --timeout=3m")

    kubectl.run_kubectl("create -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")


def _cleanup(kubectl):
    kubectl.run_kubectl("delete -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")
    istioctl_delete = ("""
                       istioctl --kubeconfig={config} manifest generate \
                                --set profile=default \
                                --set addonComponents.prometheus.enabled=false \
                                --set hub=registry.suse.de/devel/caasp/4.5/containers/containers/caasp/v4.5 \
                                --set tag=1.5.4 \
                                --set values.pilot.image=istio-pilot \
                                --set values.global.proxy.image=istio-proxyv2 \
                                --set values.global.proxy_init.image=istio-proxyv2 \
                                | kubectl --kubeconfig={config} delete -f - || true
                        """.format(config=kubectl.get_kubeconfig()))
    kubectl.utils.runshellcommand(istioctl_delete)


def _test_non_TLS(kubectl, worker_ip, logger):
    logger.info("Create the istio config")
    kubectl.run_kubectl("apply -f - << EOF " + GATEWAY_HTTPBIN)
    kubectl.run_kubectl("apply -f - << EOF " + VIRTUALSERVICE_HTTPBIN)

    # Wait for istio to digest the config
    time.sleep(100)

    nodePort = kubectl.run_kubectl("-n istio-system get service/istio-ingressgateway -o jsonpath='{ .spec.ports[1].nodePort }'")

    assert 30000 <= int(nodePort) <= 32767

    url = "{protocol}://{ip}:{port}{path}".format(protocol="http", ip=str(worker_ip), port=str(nodePort), path="/status/200")
    r = requests.get(url, headers={'host': 'httpbin.example.com'})

    assert 200 == r.status_code


def _test_TLS(kubectl, worker_ip, logger):
    # Create a temporary directory for the CA certificate
    temp_dir = tempfile.TemporaryDirectory()

    logger.info("Create the certificate")
    openssl_list = ["openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout example.com.key -out {directory}/example.com.crt".format(directory=temp_dir.name),
                    'openssl req -out httpbin.example.com.csr -newkey rsa:2048 -nodes -keyout httpbin.example.com.key -subj "/CN=httpbin.example.com/O=httpbin organization"',
                    "openssl x509 -req -days 365 -CA {directory}/example.com.crt -CAkey example.com.key -set_serial 0 -in httpbin.example.com.csr -out httpbin.example.com.crt".format(directory=temp_dir.name)]
    for cmd in openssl_list:
        kubectl.utils.runshellcommand(cmd)

    logger.info("Create the secret")
    kubectl.run_kubectl("-n istio-system create secret tls httpbin-credential --key=httpbin.example.com.key --cert=httpbin.example.com.crt")

    logger.info("Create the istio config")
    kubectl.run_kubectl("apply -f - << EOF " + GATEWAY_HTTPBIN_SECURE)
 
    # Wait for istio to digest the config
    time.sleep(60)

    secure_nodePort = kubectl.run_kubectl("-n istio-system get service/istio-ingressgateway -o jsonpath='{ .spec.ports[2].nodePort }'")

    assert 30000 <= int(secure_nodePort) <= 32767

    url = "{protocol}://{ip}:{port}{path}".format(protocol="https", ip='httpbin.example.com', port=str(secure_nodePort), path="/status/200")
    curl_command = "(curl -v -HHost:httpbin.example.com --resolve 'httpbin.example.com:{port}:{ip}' \
                    --cacert {directory}/example.com.crt \
                    {url}) 2>&1".format(port=secure_nodePort, ip=str(worker_ip), directory=temp_dir.name, url=url)
 
    output = kubectl.utils.runshellcommand(curl_command)

    assert "HTTP/2 200" in output


def test_istio_deployment(deployment, platform, skuba, kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Deploying istio and httpbin")
    _istio_httpbin_setup(kubectl)

    wrk_idx = 0
    ip_addresses = platform.get_nodes_ipaddrs("worker")
    worker_ip = ip_addresses[wrk_idx]

    logger.info("Testing the non-TLS use case")
    _test_non_TLS(kubectl, worker_ip, logger)

    logger.info("Testing now the TLS use case")
    _test_TLS(kubectl, worker_ip, logger)

    _cleanup(kubectl)
