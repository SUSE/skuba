import json
import logging
import time

import pytest

from tests.utils import (check_pods_ready, node_is_ready, wait)

from timeout_decorator import timeout

logger = logging.getLogger("testrunner")


def check_node_is_ready(platform, kubectl, role, nr):
    assert node_is_ready(platform, kubectl, role, nr)


@pytest.mark.pr
@pytest.mark.parametrize('role,node', [('master', 1), ('worker', 0)])
def test_hard_reboot(deployment, platform, skuba, kubectl, role, node):
    """ Reboots master and worker nodes and checks they are back ready.
    For masters, reboot master 1, as master 0 is expected to be the
    cluster leader and rebooting it may introduce transient etcd errors.
    """
    assert skuba.num_of_nodes(role) > node

    platform.ssh_run(role, node, 'sudo reboot &')

    # Allow time for kubernetes to check node readiness and detect it is not
    # ready
    time.sleep(60)

    # wait the node to become ready
    wait(check_node_is_ready,
         platform,
         kubectl,
         role,
         node,
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180,
         wait_allow=(AssertionError))

    # wait pods to become ready
    wait(check_pods_ready,
         kubectl,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180,
         wait_allow=(AssertionError))
