import json
import logging
import time

import pytest

from tests.utils import wait

from timeout_decorator import timeout

logger = logging.getLogger("testrunner")


def check_pods_ready(kubectl):
    pods = json.loads(kubectl.run_kubectl('get pods --all-namespaces -o json'))['items']
    for pod in pods:
        pod_status = pod['status']['phase']
        pod_name   = pod['metadata']['name']
        assert pod_status in ['Running', 'Completed'], f'Pod {pod_name} status {pod_status} != Running or Completed'


@pytest.mark.pr
@pytest.mark.parametrize('node_type,node_number', [('master', 1), ('worker', 0)])
def test_hard_reboot(deployment, platform, skuba, kubectl, node_type, node_number):
    assert skuba.num_of_nodes(node_type) > node_number

    logger.info('Wait for all the nodes to be ready')
    wait(kubectl.run_kubectl, 
         'wait --for=condition=Ready node --all --timeout=0',
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=300,
         wait_allow=(RuntimeError))

    logger.info(f'Rebooting {node_type} {node_number}')

    platform.ssh_run(node_type, node_number, 'sudo reboot &')

    # wait the node to reboot
    wait(platform.ssh_run,
         node_type,
         node_number,
         'echo hello',
         wait_backoff=30,
         wait_timeout=30,
         wait_elapsed=300,
         wait_allow=(RuntimeError))

    # wait the node to become ready
    wait(kubectl.run_kubectl,
         'wait --for=condition=Ready node --all --timeout=0',
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180,
         wait_allow=(RuntimeError))

    # wait pods to get become ready
    wait(check_pods_ready,
         kubectl,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180,
         wait_allow=(AssertionError))
