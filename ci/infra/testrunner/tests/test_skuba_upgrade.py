import pytest
import time


PREVIOUS_VERSION = "1.14.1"
CURRENT_VERSION = "1.15.0"


@pytest.fixture
def setup(request, platform, skuba):
    platform.provision(num_master=1, num_worker=1)

    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)


def setup_kubernetes_version(skuba, kubernetes_version=None):
    """
    Initialize the cluster with the given kubernetes_version, bootstrap it and
    join worker nodes.
    """

    skuba.cluster_init(kubernetes_version)
    skuba.node_bootstrap()
    skuba.node_join(role="worker", nr=0)


def test_upgrade_plan_all_fine(setup, skuba):
    """
    Starting from a up-to-date cluster, check what cluster/node plan report.
    """

    setup_kubernetes_version(skuba)
    out = skuba.cluster_upgrade_plan()

    assert out.find(
        "Congratulations! You are already at the latest version available"
    ) != -1


def test_upgrade_plan_from_previous(setup, skuba):
    """
    Starting from an outdated cluster, check what cluster/node plan report.
    """

    setup_kubernetes_version(skuba, PREVIOUS_VERSION)

    # cluster upgrade plan
    out = skuba.cluster_upgrade_plan()
    assert out.find("Current Kubernetes cluster version: {pv}".format(
        pv=PREVIOUS_VERSION)) != -1
    assert out.find("Latest Kubernetes version: {cv}".format(
        cv=CURRENT_VERSION)) != -1
    assert out.find(
        "Upgrade path to update from {pv} to {cv}:\n - {pv} -> {cv}".format(
            pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)
    ) != -1

    # node upgrade plan
    outs = {}
    for (r, n) in [("master", 0), ("worker",1)]:
        node = "my-{}-{}".format(n,r)
        outs[node] = skuba.node_upgrade_plan(r, n)

    master = outs["my-master-0"]
    assert master.find(
        "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
    assert master.find("Latest Kubernetes version: {cv}".format(
        cv=CURRENT_VERSION)) != -1
    assert master.find(" - apiserver: {pv} -> {cv}".format(
        pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1
    assert master.find("  - kubelet: {pv} -> {cv}".format(
        pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1

    worker = outs["my-worker-0"]
    assert worker.find(
        "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
    assert worker.find("Latest Kubernetes version: {cv}".format(
        cv=CURRENT_VERSION)) != -1
    assert worker.find(
        "  - kubelet: {pv} -> {cv}".format(pv=PREVIOUS_VERSION,
                                           cv=CURRENT_VERSION)) != -1
