import pytest
import time


@pytest.fixture
def setup(request, platform, skuba):
    platform.provision()

    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)


def setup_kubernetes_version(skuba, kubernetes_version=""):
    """
    Initialize the cluster with the given kubernetes_version, bootstrap it and
    join worker nodes.
    """

    skuba.cluster_init(kubernetes_version)
    skuba.node_bootstrap()
    skuba.node_join(role="worker", nr=0)


##
# cluster upgrade plan


def test_cluster_upgrade_plan_all_fine(setup, skuba):
    setup_kubernetes_version(skuba)
    out = skuba._run_skuba("cluster upgrade plan", output=True)

    assert out.find(
        "Congratulations! You are already at the latest version available"
    ) != -1


def test_cluster_upgrade_plan_from_v1_14(setup, skuba):
    setup_kubernetes_version(skuba, "1.14.1")
    out = skuba._run_skuba("cluster upgrade plan", output=True)

    assert out.find("Current Kubernetes cluster version: 1.14.1") != -1
    assert out.find("Latest Kubernetes version: 1.15.0") != -1
    assert out.find(
        "Upgrade path to update from 1.14.1 to 1.15.0:\n - 1.14.1 -> 1.15.0"
    ) != -1


##
# node upgrade plan


def test_node_upgrade_plan_all_fine(setup, skuba):
    setup_kubernetes_version(skuba)
    outs = {}
    for n in ("my-master-0", "my-worker-1"):
        outs[n] = skuba._run_skuba(
            "node upgrade plan {}".format(n), output=True
        )

    for node, out in outs.iteritems():
        assert out.find("Node {} is up to date".format(node)) != -1


def test_node_upgrade_plan_from_v1_14(setup, skuba):
    setup_kubernetes_version(skuba, "1.14.1")
    outs = {}
    for n in ("my-master-0", "my-worker-1"):
        outs[n] = skuba._run_skuba(
            "node upgrade plan {}".format(n), output=True
        )

    master = outs["my-master-0"]
    assert master.find("Current Kubernetes cluster version: 1.14.1")
    assert master.find("Latest Kubernetes version: 1.15.0") != -1
    assert master.find("  - apiserver: 1.14.1 -> 1.15.0") != -1
    assert master.find("  - kubelet: 1.14.1 -> 1.15.0") != -1

    worker = outs["my-worker-0"]
    assert worker.find("Current Kubernetes cluster version: 1.14.1")
    assert worker.find("Latest Kubernetes version: 1.15.0") != -1
    assert worker.find("  - kubelet: 1.14.1 -> 1.15.0") != -1
