import pytest

from tests.utils import setup_kubernetes_version, PREVIOUS_VERSION, CURRENT_VERSION


@pytest.mark.disruptive
def test_upgrade_plan_from_previous(setup, skuba, kubectl, platform):
    """
    Starting from an outdated cluster, check what cluster/node plan report.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

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
    masters = platform.get_num_nodes("master")
    for n in range(0, masters):
        master = skuba.node_upgrade("plan", "master", n)
        assert master.find(
            "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
        assert master.find("Latest Kubernetes version: {cv}".format(
            cv=CURRENT_VERSION)) != -1
        assert master.find(" - apiserver: {pv} -> {cv}".format(
            pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1
        assert master.find(" - kubelet: {pv} -> {cv}".format(
            pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1

    workers = platform.get_num_nodes("worker")
    worker_names = platform.get_nodes_names("worker")
    for n in range(0, workers):
        worker = skuba.node_upgrade("plan", "worker", n, ignore_errors=True)
        # If the control plane nodes are not upgraded yet, skuba disallows upgrading a worker
        assert worker.find("Unable to plan node upgrade: {} is not upgradeable until all control plane nodes are upgraded".format(worker_names[n]))