import logging
import pytest
import re
import time

logger = logging.getLogger("testrunner")


@pytest.mark.flaky
def test_cillium(deployment, kubectl):

    landing_req='curl -sm10 -XPOST deathstar.default.svc.cluster.local/v1/request-landing'

    logger.info("Deploy deathstar")
    kubectl.run_kubectl("create -f https://raw.githubusercontent.com/cilium/cilium/v1.5/examples/minikube/http-sw-app.yaml")
    kubectl.run_kubectl("wait --for=condition=ready pods --all --timeout=3m")

    # FIXME: this hardcoded wait should be replaces with a (cilum?) condition
    time.sleep(100)

    logger.info("Check with L3/L4 policy")
    kubectl.run_kubectl("create -f https://raw.githubusercontent.com/cilium/cilium/v1.5/examples/minikube/sw_l3_l4_policy.yaml")
    tie_out = kubectl.run_kubectl("exec tiefighter -- {}".format(landing_req))
    assert 'Ship landed' in tie_out

    xwing_out = kubectl.run_kubectl("exec xwing -- {} 2>&1 || :".format(landing_req))
    assert 'terminated with exit code 28' in xwing_out

    logger.info("Check status (N/N)")
    node_list = kubectl.run_kubectl("get nodes -o jsonpath='{ .items[*].metadata.name }'")
    node_count = len(node_list.split(" "))
    cilium_podlist = kubectl.run_kubectl("get pods -n kube-system -l k8s-app=cilium -o jsonpath='{ .items[0].metadata.name }'").split(" ")
    cilium_podid = cilium_podlist[0]
    cilium_status_cmd = "-n kube-system exec {} -- cilium status".format(cilium_podid)
    cilium_status = kubectl.run_kubectl(cilium_status_cmd)
    assert re.search(r'Controller Status:\s+([0-9]+)/\1 healthy', cilium_status) is not None

    for i in range(1, 10):
        cilium_status = kubectl.run_kubectl(cilium_status_cmd) 
        all_reachable = re.search(r"Cluster health:\s+({})/\1 reachable".format(node_count), cilium_status)
        if all_reachable:
           break
        time.sleep(30)

    assert all_reachable
