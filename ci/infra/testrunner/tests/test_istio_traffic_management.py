import pytest
import requests
import subprocess
import logging
import time
from istioctl.istioctl import (Istioctl)


_istio_version = "1.5.4"
_sample_path = "https://raw.githubusercontent.com/istio/istio/%s/samples" % _istio_version
_binfo_vserv = "bookinfo/networking/virtual-service"
   
HTTPBIN_RULE = ("""
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: httpbin
spec:
  host: httpbin
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 1
      http:
        http1MaxPendingRequests: 1
        maxRequestsPerConnection: 1
    outlierDetection:
      consecutiveErrors: 1
      interval: 1s
      baseEjectionTime: 3m
      maxEjectionPercent: 100
EOF
""")

HTTPBIN_V1_DEPLOYMENT = ("""
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin-v1
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
      version: v1
  template:
    metadata:
      labels:
        app: httpbin
        version: v1
    spec:
      containers:
      - image: docker.io/kennethreitz/httpbin
        imagePullPolicy: IfNotPresent
        name: httpbin
        command: ["gunicorn", "--access-logfile", "-", "-b", "0.0.0.0:80", "httpbin:app"]
        ports:
        - containerPort: 80
EOF
""")

HTTPBIN_V2_DEPLOYMENT = ("""
---   
apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin-v2
spec:
  replicas: 1
  selector:
    matchLabels:
      app: httpbin
      version: v2
  template:
    metadata:
      labels:
        app: httpbin
        version: v2
    spec:
      containers:
      - image: docker.io/kennethreitz/httpbin
        imagePullPolicy: IfNotPresent
        name: httpbin
        command: ["gunicorn", "--access-logfile", "-", "-b", "0.0.0.0:80", "httpbin:app"]
        ports:
        - containerPort: 80
EOF
""")

HTTPBIN_SERVICE = ("""
---
apiVersion: v1
kind: Service
metadata:
  name: httpbin
  labels:
    app: httpbin
spec:
  ports:
  - name: http
    port: 8000
    targetPort: 80
  selector:
    app: httpbin
EOF
""")

SLEEP_DEPLOYMENT = ("""
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
    spec:
      containers:
      - name: sleep
        image: tutum/curl
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: IfNotPresent
EOF
""")

HTTPBIN_SERVICE_RULE = ("""
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: httpbin
spec:
  hosts:
    - httpbin
  http:
  - route:
    - destination:
        host: httpbin
        subset: v1
      weight: 100
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: httpbin
spec:
  host: httpbin
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
EOF
""")

HTTPBIN_SERVICE_MIRROR = ("""
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: httpbin
spec:
  hosts:
    - httpbin
  http:
  - route:
    - destination:
        host: httpbin
        subset: v1
      weight: 100
    mirror:
      host: httpbin
      subset: v2
    mirror_percent: 100
EOF
""")

def _cleanup_cluster(kubectl):
    logger.info('Removing all Istio Traffic test artifacts from the cluster')

    tempfile = tempfile.TemporaryDirectory("$HOME/istio")
    cmd = "istioctl manifest generate > %s/generated-manifest.yaml" % tempfile

    kubectl.utils.runshellcommand(cmd)

    cmd = "istioctl verify-install -f %s/generated-manifest.yaml" % tempfile
    kubectl.utils.runshellcommand(cmd)

def _test_setup_istio(kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Checking if Istio is installed and setup correctly")
    kubeconfig = kubectl.get_kubeconfig()

    # Enable injection only for default namespace here. 
    # Tests will need explicitly enabling injection if different namespace
    # is used (for e.g: TCP Traffic Shifting Test)
    kubectl.run_kubectl("label --overwrite namespace default istio-injection=enabled")

    istioctl = ("""
            istioctl --kubeconfig={config} manifest apply \
                     --set profile=default \
                     --set addonComponents.prometheus.enabled=false \
                     --set hub={hub} \
                     --set tag={tag} \
                     --set values.pilot.image=istio-pilot \
                     --set values.global.proxy.image=istio-proxyv2 \
                     --set values.global.proxy_init.image=istio-proxyv2
             """.format(config=kubeconfig,
                        hub=kubeconfig.istioctl.container_hub,
                        tag=kubeconfig.istioctl.istio_version))

    logger.info("Running command: + " + istioctl)
    kubectl.utils.runshellcommand(istioctl)

    kubectl.run_kubectl("-n istio-system wait --for=condition=available deploy/istio-ingressgateway --timeout=3m")

def _test_istio_cleanup(kubectl):
    logger = logging.getLogger("testrunner")
    kubeconfig = kubectl.get_kubeconfig()

    kubectl.run_kubectl("delete --ignore-not-found=true -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml")
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
                    """.format(config=kubeconfig,
                               hub=kubeconfig.istioctl.container_hub,
                               tag=kubeconfig.istioctl.istio_version))
    
    logger.info("Running command: + " + istioctl_delete)
    kubectl.utils.runshellcommand(istioctl_delete)

def _get_ingress_address(kubectl):
    logger = logging.getLogger("testrunner")
    wrk_idx = 0
    ip_addresses = platform.get_nodes_ipaddrs("worker")
    assert 0 < len(ip_addresses)
    
    worker_ip = ip_addresses[wrk_idx]
    
    port = kubectl.run_kubectl("-n istio-system get service/istio-ingressgateway -o jsonpath='{ .spec.ports[2].nodePort }'")
    assert 30000 <= int(port) <= 32767
    
    logger.info("Obtain Ingress Path: %s:%s" % (worker_ip, port))
    return "%s:%s" % (worker_ip, port)


def _setup_bookinfo(kubectl):
    kubectl.run_kubectl("apply -f %s/bookinfo/platform/kube/bookinfo.yaml" % _sample_path)
    #wait until the bookinfo setup is complete
    time.sleep(30)

def _cleanup_bookinfo(kubectl):
    kubectl.run_kubectl("delete --ignore-not-found=true -f %s/%s-all-v1.yaml" % (_sample_path, _binfo_vserv)) 

def _test_istio_request_routing(kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Testing Istio Request Routing")
    _cleanup_bookinfo(kubectl)
    _setup_bookinfo(kubectl)

    kubectl.run_kubectl("apply -f %s/%s-all-v1.yaml" % (_sample_path, _binfo_vserv))
    time.sleep(30)
    kubectl.run_kubectl("apply -f %s/%s-reviews-test-v2.yaml" % (_sample_path, _binfo_vserv))
    time.sleep(30)

    #TODO: Need a way to test if the applications run correctly
    _cleanup_bookinfo(kubectl)
    logger.info("Request Routing Test PASSED")


def _test_istio_traffic_shifting(kubectl, worker_ip):
    logger = logging.getLogger("testrunner")
    networking_service_list = ['all-v1',
                               'reviews-50-v3',
                               'reviews-v3']
    tcp_service_list = ['tcp-echo-services', 'tcp-echo-all-v1']
    tcp_ns = "istio-io-tcp-traffic-shifting"

    # Create a clean setup of bookinfo application
    _cleanup_bookinfo(kubectl)
    _setup_bookinfo(kubectl)

    for _service in networking_service_list:
        kubectl.run_kubectl("apply -f %s/%s-%s.yaml" % (_sample_path,
                                                        _binfo_vserv,
                                                        _service))

    cmd = "kubectl get namespaces | grep istio-io-tcp-traffic-shifting | awk '{print $1}'"
    logger.info("Running COMMAND: %s" % cmd)
    is_namespace = kubectl.utils.runshellcommand(cmd)

    logger.info("IS NAMESPACE: %s" % is_namespace)
    if not is_namespace:
        kubectl.run_kubectl("create namespace %s" % tcp_ns)
    kubectl.run_kubectl("label --overwrite namespace %s istio-injection=enabled" % tcp_ns)
    
    for _service in tcp_service_list:
        kubectl.run_kubectl("apply -f %s/tcp-echo/%s.yaml -n %s" % (_sample_path,
                                                                    _service,
                                                                    tcp_ns))

    # Wait for istio to start and configure these services
    time.sleep(60)
    
    INGRESS_PORT = kubectl.run_kubectl("-n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name==\"tcp\")].port}'")

    assert 30000 <= int(INGRESS_PORT) <= 32767

    docker_run = ("""
                  docker run -e INGRESS_HOST={ingress_host} \
                  -e INGRESS_PORT={ingress_port} --rm busybox sh \
                  -c "(date; sleep 1)"
                  """.format(ingress_host=worker_ip, ingress_port=INGRESS_PORT))
 
    for i in range(10):
        kubectl.utils.runshellcommand(docker_run)

    kubectl.run_kubectl("apply -f %s/tcp-echo/tcp-echo-20-v2.yaml -n istio-io-tcp-traffic-shifting" % _sample_path)
    kubectl.run_kubectl("get virtualservice tcp-echo -o yaml -n istio-io-tcp-traffic-shifting")

    docker_run = ("""
                  docker run -e INGRESS_HOST={ingress_host} \
                  -e INGRESS_PORT={ingress_port} --rm busybox sh \
                  -c "(date; sleep 1)"
                  """.format(ingress_host=worker_ip, ingress_port=INGRESS_PORT))

    for i in range(10):
        kubectl.utils.runshellcommand(docker_run)

    # Cleanup
    for _service in tcp_service_list:
        kubectl.run_kubectl("delete --ignore-not-found=true -f %s/tcp-echo/%s.yaml -n %s" % (_sample_path,
                                                                     _service,
                                                                     tcp_ns))
    kubectl.run_kubectl("delete --ignore-not-found=true namespace %s" % tcp_ns)

    logger.info("Traffic Shifting test PASSED")


def _test_istio_circuit_breaking(kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Testing Istio Circuit Breaking")
    curlpath = "http://httpbin:8000/get"
    fortiobin = "/usr/bin/fortio"

    kubectl.run_kubectl("apply -f - << EOF " + HTTPBIN_RULE)

    kubectl.run_kubectl("""apply -f {path}/httpbin/sample-client/fortio-deploy.yaml""".format(path=_sample_path))
   
    # Wait for fortio to be ready for use.
    time.sleep(60)

    FORTIO_POD = kubectl.utils.runshellcommand("""kubectl get pod | grep pod | grep fortio | awk '{print $1}'""")
    logger.info("FORTIO POD: %s" % FORTIO_POD)

    #kubectl.run_kubectl("""exec -it {pod} -c fortio {fbin} \
    #                       -- load -curl {cpath}\
    #                    """.format(pod=FORTIO_POD, fbin=fortiobin, cpath=curlpath))

    exec_cmd = "kubectl exec -it %s -c fortio %s -- load -curl %s " % (FORTIO_POD, fortiobin, curlpath)
    logger.info("Executing command: %s" % exec_cmd)
    kubectl.utils.runshellcommand(exec_cmd)

    kubectl.run_kubectl("""exec -it {pod}  -c fortio {fbin} \
                           -- load -c 2 -qps 0 -n 20 -loglevel Warning {cpath}\
                        """.format(pod=FORTIO_POD, fbin=fortiobin, cpath=curlpath))

    kubectl.run_kubectl("""exec -it {pod  -c fortio {fbin} \
                           -- load -c 3 -qps 0 -n 30 -loglevel Warning {cpath}\
                        """.format(pod=FORTIO_POD, fbin=fortiobin, cpath=curlpath))

    kubectl.run_kubectl("""exec -it {pod} -c istio-proxy \
                           -- pilot-agent request GET stats | grep httpbin | grep pending
                        """.format(pod=FORTIO_POD))

    # Cleanup cluster
    kubectl.run_kubectl("delete --ignore-not-found=true destinationrule httpbin")
    kubectl.run_kubectl("delete --ignore-not-found=true deploy httpbin fortio-deploy")
    kubectl.run_kubectl("delete --ignore-not-found=true svc httpbin fortio")
    
    logger.info("Circuit Breaking Test PASSED")

def _test_istio_mirroring(kubectl):
    logger = logging.getLogger("testrunner")
    logger.info("Testing Istio Mirroring")

    # Create a clean setup of bookinfo application
    _cleanup_bookinfo(kubectl)
    _setup_bookinfo(kubectl)

    kubectl.run_kubectl("apply -f %s/bookinfo/platform/kube/bookinfo.yaml" % _sample_path)

    # cat << EOF | istioctl kube-inject -f - | kubectl create -f -
    # httpbin_v1
    kubectl.run_kubectl("apply -f - << EOF " + HTTPBIN_V1)
    time.sleep(30)

    # cat <<EOF | istioctl kube-inject -f - | kubectl create -f -
    # httpbin_v2
    kubectl.run_kubectl("apply -f - << EOF " + HTTPBIN_V2)
    time.sleep(30)

    # kubectl create -f - <<EOF
    # httpbin Service
    kubectl.run_kubectl("apply -f - << EOF " + HTTPBIN_SERVICE)
    time.sleep(30)

    # cat <<EOF | istioctl kube-inject -f - | kubectl create -f -
    # sleep deployment
    kubectl.run_kubectl("apply -f - << EOF " + SLEEP_DEPLOYMENT)
    time.sleep(30)

    # kubectl apply -f - <<EOF
    # httpbin vserv rule
    kubectl.run_kubectl("apply -f - << EOF " + HTTPBIN_SERVICE_RULE)

    time.sleep(60)

    SLEEP_POD = kubectl.run_kubectl("""kubectl get pod -l app=sleep -o jsonpath={.items..metadata.name}""")

    kubectl.run_kubectl("exec -it $SLEEP_POD -c sleep -- sh -c 'curl  http://httpbin:8000/headers' | python -m json.tool")

    kubectl.run_kubectl("logs -f $V1_POD -c httpbin")

    V2_POD = kubectl.run_kubectl("""get pod -l app=httpbin,version=v2 -o jsonpath={.items..metadata.name}""")
    

    # kubectl apply -f - <<EOF
    # http service Mirror
    #
    kubectl.run_kubectl("exec -it $SLEEP_POD -c sleep -- sh -c 'curl  http://httpbin:8000/headers' | python -m json.tool")
    kubectl.run_kubectl("logs -f $V1_POD -c httpbin")
    kubectl.run_kubectl("logs -f $V2_POD -c httpbin")

    # Cleanup the cluster
    kubectl.run_kubectl("delete --ignore-not-found=true virtualservice httpbin")
    kubectl.run_kubectl("delete --ignore-not-found=true destinationrule httpbin")
    kubectl.run_kubectl("delete --ignore-not-found=true deploy httpbin-v1 httpbin-v2 sleep")
    kubectl.run_kubectl("delete --ignore-not-found=true svc httpbin")

    logger.info("Traffic Mirroring Test PASSED")

def test_istio_traffic_shaping(deployment, platform, skuba, kubectl, tests=[]):
    logger = logging.getLogger("testrunner")
    logger.info("Testing Istio Traffic Shaping")

    ictl = Istioctl(kubectl)
    #_test_setup_istio(kubectl)
    ictl.istio_setup(kubectl)
    time.sleep(60)
    wrk_idx = 0
    ip_addresses = platform.get_nodes_ipaddrs("worker")
    worker_ip = ip_addresses[wrk_idx]

    all_tests = ["traffic_shifting",
                 "request_routing",
                 "circuit_breaking",
                 "traffic_mirroring"]

    if not tests:
      tests_to_do = [x for x in all_tests]
    else:
      tests_to_do = [x for x in tests]

    logger.info("Tests to do: " + str(tests_to_do))
    # Run all tests by default or selectively if given a choice
    if "traffic_shifting" in tests_to_do:
         _test_istio_traffic_shifting(kubectl, worker_ip)

    if "request_routing" in tests_to_do:
        _test_istio_request_routing(kubectl)

    if "circuit_breaking" in tests_to_do:
        _test_istio_circuit_breaking(kubectl)

    if "traffic_mirroring" in tests_to_do:
        _test_istio_traffic_mirroring(kubectl)

    _test_istio_cleanup(kubectl)

    logger.info("All four Traffic Management Tests PASSED")

