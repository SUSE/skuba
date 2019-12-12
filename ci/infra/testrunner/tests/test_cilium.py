import logging
import re
import time

import pytest

from tests.utils import (wait, daemon_set_is_ready)

logger = logging.getLogger("testrunner")

CILIUM_VERSION = '1.6'


@pytest.fixture()
def deploy_deathstar(request, kubectl):
    logger.info("Deploy deathstar")
    kubectl.run_kubectl(f"create -f https://raw.githubusercontent.com/cilium/cilium/v{CILIUM_VERSION}/examples/minikube/http-sw-app.yaml")

    def cleanup():
        kubectl.run_kubectl("delete deploy/deathstar svc/deathstar pod/tiefighter pod/xwing")

    request.addfinalizer(cleanup)

    wait(kubectl.run_kubectl,
         "wait --for=condition=available deployment/deathstar --timeout=0",
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180)

    wait(kubectl.run_kubectl,
         "wait --for=condition=ready pod/tiefighter pod/xwing --timeout=0",
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180)

    wait(daemon_set_is_ready,
         kubectl,
         "cilium",
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180)


@pytest.fixture()
def deploy_l3_l4_policy(request, kubectl, deploy_deathstar):
    logger.info("Deploy l3 and l4 policy")
    kubectl.run_kubectl(f"create -f https://raw.githubusercontent.com/cilium/cilium/v{CILIUM_VERSION}/examples/minikube/sw_l3_l4_policy.yaml")

    def cleanup():
        kubectl.run_kubectl("delete cnp/rule1")

    request.addfinalizer(cleanup)


@pytest.fixture()
def deploy_l3_l4_l7_policy(request, kubectl, deploy_deathstar):
    logger.info("Deploy l3, l4, and l7 policy")
    kubectl.run_kubectl(f"create -f https://raw.githubusercontent.com/cilium/cilium/v{CILIUM_VERSION}/examples/minikube/sw_l3_l4_l7_policy.yaml")

    def cleanup():
        kubectl.run_kubectl("delete cnp/rule1")

    request.addfinalizer(cleanup)


def get_cilium_pod(kubectl):
    cilium_podlist = kubectl.run_kubectl("get pods -n kube-system -l k8s-app=cilium -o jsonpath='{ .items[0].metadata.name }'").split(" ")
    return cilium_podlist[0]


def test_cilium_version(deployment, kubectl):
    cilium_pod_id = get_cilium_pod(kubectl)
    cilium_version = kubectl.run_kubectl(f"-n kube-system exec {cilium_pod_id} -- cilium version")
    cilium_client_version = re.search(r'(?<=Client:\s)[\d\.]+', cilium_version).group(0)

    logger.info(f'Cilium client version is {cilium_client_version}')
    assert cilium_client_version.startswith(CILIUM_VERSION)


def test_cilium(deployment, kubectl, deploy_l3_l4_policy):
    landing_req = 'curl -sm10 -XPOST deathstar.default.svc.cluster.local/v1/request-landing'

    logger.info("Check with L3/L4 policy")
    tie_out = kubectl.run_kubectl("exec tiefighter -- {}".format(landing_req))
    assert 'Ship landed' in tie_out

    xwing_out = kubectl.run_kubectl("exec xwing -- {} 2>&1 || :".format(landing_req))
    assert 'terminated with exit code 28' in xwing_out

    logger.info("Check status (N/N)")
    node_list = kubectl.run_kubectl("get nodes -o jsonpath='{ .items[*].metadata.name }'")
    node_count = len(node_list.split(" "))
    cilium_pod_id = get_cilium_pod(kubectl)
    cilium_status_cmd = "-n kube-system exec {} -- cilium status".format(cilium_pod_id)
    cilium_status = kubectl.run_kubectl(cilium_status_cmd)
    assert re.search(r'Controller Status:\s+([0-9]+)/\1 healthy', cilium_status) is not None

    for i in range(1, 10):
        cilium_status = kubectl.run_kubectl(cilium_status_cmd)
        all_reachable = re.search(r"Cluster health:\s+({})/\1 reachable".format(node_count), cilium_status)
        if all_reachable:
            break
        time.sleep(30)

    assert all_reachable


def test_cilium_l7(deployment, kubectl, deploy_l3_l4_l7_policy):
    """
    GIVEN cilium is properly configured
    AND an L7 policy that restricts access to an endpoint is applied
    WHEN a unautorized request is made to the endpoint
    THEN Access is denied
    """
    landing_req = 'curl -sm10 -XPOST deathstar.default.svc.cluster.local/v1/request-landing'
    exhaust_port = 'curl -s -XPUT deathstar.default.svc.cluster.local/v1/exhaust-port'

    logger.info("Check with L3/L4/L7 policy")
    landing_req_out = kubectl.run_kubectl(f"exec tiefighter -- {landing_req}")
    assert 'Ship landed' in landing_req_out

    exhaust_port_out = kubectl.run_kubectl(f"exec tiefighter -- {exhaust_port}")
    assert 'Access denied' in exhaust_port_out
