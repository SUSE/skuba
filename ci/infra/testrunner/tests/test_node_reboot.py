import logging
import json
import time

import pytest
from timeout_decorator import timeout

logger = logging.getLogger("testrunner")


@timeout(120)
def wait_for_node_reboot(node_type, node_number, platform):
    while True:
        try:
            platform.ssh_run(node_type, node_number, 'echo hello')
        except Exception as ex:
            logger.debug(ex)
        else:
            break

        time.sleep(5)


@timeout(120)
def check_nodes_ready(kubectl):
    while True:
        try:
            kubectl.run_kubectl('wait --for=condition=Ready node --all')
        except Exception as ex:
            logger.debug(ex)
        else:
            break

        time.sleep(5)


@timeout(300)
def check_pods_ready(kubectl):
    while True:
        try:
            pods = json.loads(kubectl.run_kubectl('get pods --all-namespaces -o json'))['items']

            for pod in pods:
                pod_status = pod['status']['phase']
                assert pod_status in ['Running', 'Completed'], f'Pod {pod["metadata"]["name"]} status {pod_status} != Running or Completed'
        except Exception as ex:
            logger.debug(ex)
        else:
            break

        time.sleep(5)


@pytest.mark.parametrize('node_type,node_number', [('master', 1), ('worker', 0)])
def test_hard_reboot(deployment, platform, skuba, kubectl, node_type, node_number):
    assert skuba.num_of_nodes(node_type) > node_number

    logger.info('Wait for all the nodes to be ready')
    kubectl.run_kubectl('wait --for=condition=Ready node --all --timeout=5m')

    logger.info(f'Rebooting {node_type} {node_number}')

    platform.ssh_run(node_type, node_number, 'sudo reboot &')

    # Give the reboot a bit to start
    logger.info(f'Waiting 30s...')
    time.sleep(30)

    wait_for_node_reboot(node_type, node_number, platform)
    check_nodes_ready(kubectl)
    check_pods_ready(kubectl)
