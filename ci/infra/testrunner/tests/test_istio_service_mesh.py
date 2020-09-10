import logging
import pytest
import time

CLUSTERROLEBINDING = ("""
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: suse:caasp:psp:privileged
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
- kind: ServiceAccount
  name: httpbin
  namespace: default
- kind: ServiceAccount
  name: bookinfo-productpage
  namespace: default
- kind: ServiceAccount
  name: bookinfo-reviews
  namespace: default
- kind: ServiceAccount
  name: bookinfo-ratings 
  namespace: default
- kind: ServiceAccount
  name: bookinfo-details
  namespace: default
- kind: ServiceAccount
  name: sleep
  namespace: default
EOF
""")

ISTIO_VERSION = "1.5"
ISTIO_VERSION_PATCH = ISTIO_VERSION + ".9"
ISTIO_URL = "https://raw.githubusercontent.com/istio/istio/release-" + ISTIO_VERSION + "/samples"

def _istio_bookinfo_setup(kubectl):
    istioctl = ("""
                istioctl --kubeconfig={config} manifest apply \
                         --set profile=default \
                         --set addonComponents.prometheus.enabled=false \
                         --set hub=registry.suse.de/devel/caasp/4.5/containers/containers/caasp/v4.5 \
                         --set tag={version} \
                         --set values.pilot.image=istio-pilot \
                         --set values.global.proxy.image=istio-proxyv2 \
                         --set values.global.proxy_init.image=istio-proxyv2
                 """.format(config=kubectl.get_kubeconfig(), version=ISTIO_VERSION_PATCH))

    kubectl.utils.runshellcommand(istioctl)
    kubectl.run_kubectl("-n istio-system wait --for=condition=available deploy/istio-ingressgateway --timeout=3m")

    # Activate service in namespace default
    kubectl.run_kubectl("label namespace default istio-injection=enabled --overwrite")

    # Create the clusterrolebinding to run everything
    kubectl.run_kubectl("apply -f - << EOF " + CLUSTERROLEBINDING)

    # Deploy bookinfo application
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/platform/kube/bookinfo.yaml")
    kubectl.run_kubectl("-n default wait --for=condition=available deploy/productpage-v1 --timeout=3m")

    # Deploy the gateway and destination rules
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/networking/bookinfo-gateway.yaml")
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/networking/destination-rule-all.yaml")

    # Deploy the sleep pod that we use as client
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/sleep/sleep.yaml")
    kubectl.run_kubectl("-n default wait --for=condition=available deploy/sleep --timeout=3m")


def _cleanup(kubectl):
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/bookinfo/platform/kube/bookinfo.yaml")
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/bookinfo/networking/bookinfo-gateway.yaml")
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/bookinfo/networking/destination-rule-all.yaml")
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/bookinfo/networking/virtual-service-reviews-v3.yaml")
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/sleep/sleep.yaml")
    kubectl.run_kubectl(f"delete -f {ISTIO_URL}/httpbin/httpbin.yaml")
    kubectl.run_kubectl("label namespace default istio-injection-")


    istioctl_delete = ("""
                       istioctl --kubeconfig={config} manifest generate \
                                --set profile=default \
                                --set addonComponents.prometheus.enabled=false \
                                --set hub=registry.suse.de/devel/caasp/4.5/containers/containers/caasp/v4.5 \
                                --set tag={version} \
                                --set values.pilot.image=istio-pilot \
                                --set values.global.proxy.image=istio-proxyv2 \
                                --set values.global.proxy_init.image=istio-proxyv2 \
                                | kubectl --kubeconfig={config} delete -f - || true
                        """.format(config=kubectl.get_kubeconfig(), version=ISTIO_VERSION_PATCH))
    kubectl.utils.runshellcommand(istioctl_delete)


def _test_traffic_shift(kubectl, logger):
    '''
    It tests the traffic shift feature of service mesh. There are two versions of the review service, v1 and v3. We will run three subtest:
    1 - All traffic goes to v1
    2 - 50% of traffic goes to v1 and 50% to v3
    3 - All traffic goes to v3

    The productpage queries the review service and returns the string "glyphicon glyphicon-star" when it uses v3. That's what we will search in the output
    '''
    logger.info("Create the traffic shift config")

    v3_string = "glyphicon glyphicon-star"
    sleep_pod = kubectl.run_kubectl("get pod -l app=sleep -n default -o 'jsonpath={.items..metadata.name}'")

    # Create the virtual service that sends 100% of traffic to v1
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/networking/virtual-service-all-v1.yaml")
    time.sleep(30)

    # We shouldn't find the v3 string because all traffic goes to v1
    for i in range(5):
        output = kubectl.run_kubectl("exec {pod} -c sleep -n default -- curl -s http://istio-ingressgateway.istio-system/productpage".format(pod=sleep_pod))
        assert output.find(v3_string) == -1
        time.sleep(5)

    # Now v1 and v3 have 50% weight
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/networking/virtual-service-reviews-50-v3.yaml")
    time.sleep(60)

    v1 = 0
    v3 = 0
    for i in range(5):
        output = kubectl.run_kubectl("exec {pod} -c sleep -n default -- curl -s http://istio-ingressgateway.istio-system/productpage".format(pod=sleep_pod))
        if output.find(v3_string) > 0:
            v3 += 1
        else:
            v1 += 1
        time.sleep(5)

    # We must have reached at least once v1 and v3
    assert v1 > 0
    assert v3 > 0

    # Now v3 has 100% weight
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/bookinfo/networking/virtual-service-reviews-v3.yaml")
    time.sleep(60)

    for i in range(5):
        output = kubectl.run_kubectl("exec {pod} -c sleep -n default -- curl -s http://istio-ingressgateway.istio-system/productpage".format(pod=sleep_pod))
        assert output.find(v3_string) != -1
        time.sleep(5)

def _test_mTLS(kubectl, logger):
    '''
    It tests that mTLS is active (default behaviour) and thus authentication is happening in the service mesh

    To do so, we check that there is a client certificate in the HTTP request
    '''
    logger.info("Add httpbin deployment")
    kubectl.run_kubectl(f"apply -f {ISTIO_URL}/httpbin/httpbin.yaml")
    kubectl.run_kubectl("-n default wait --for=condition=available deploy/httpbin --timeout=3m")

    sleep_pod = kubectl.run_kubectl("get pod -l app=sleep -n default -o 'jsonpath={.items..metadata.name}'")
    httpbin_port = kubectl.run_kubectl("get service -l app=httpbin -o 'jsonpath={.items..spec.ports..port}'")
    output = kubectl.run_kubectl("exec {pod} -c sleep -n default -- curl -s http://httpbin.default:{port}/headers".format(pod=sleep_pod, port=httpbin_port))

    assert output.find("X-Forwarded-Client-Cert") != -1


def test_istio_service_mesh(deployment, platform, skuba, kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Deploying istio and the bookinfo app")
    _istio_bookinfo_setup(kubectl)

    logger.info("Testing the traffic shifting")
    _test_traffic_shift(kubectl, logger)

    logger.info("Testing the authentication")
    _test_mTLS(kubectl, logger)

    _cleanup(kubectl)
