import pytest

from tests.utils import PREVIOUS_VERSION, CURRENT_VERSION, setup_kubernetes_version, node_is_ready, node_is_upgraded


@pytest.mark.disruptive
def test_upgrade_plan_from_previous_with_upgraded_control_plane(setup, skuba, kubectl, platform):
    """
    Starting from an updated control plane, check what cluster/node plan report.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

    masters = platform.get_num_nodes("master")
    for n in range(0, masters):
        assert node_is_ready(platform, kubectl, "master", n)
        master = skuba.node_upgrade("apply", "master", n)
        assert master.find("successfully upgraded") != -1
        assert node_is_upgraded(kubectl, platform, "master", n)

    workers = platform.get_num_nodes("worker")
    for n in range(0, workers):
        worker = skuba.node_upgrade("plan", "worker", n)
        assert worker.find(
            "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
        assert worker.find("Latest Kubernetes version: {cv}".format(
            cv=CURRENT_VERSION)) != -1
        assert worker.find(" - kubelet: {pv} -> {cv}".format(
            pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1