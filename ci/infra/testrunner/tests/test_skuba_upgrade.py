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
    for (r, n) in [("master", 0), ("worker", 0)]:
        node = "my-{}-{}".format(r, n)
        outs[node] = skuba.node_upgrade("plan", r, n)

    master = outs["my-master-0"]
    assert master.find(
        "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
    assert master.find("Latest Kubernetes version: {cv}".format(
        cv=CURRENT_VERSION)) != -1
    assert master.find(" - apiserver: {pv} -> {cv}".format(
        pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1
    assert master.find(" - kubelet: {pv} -> {cv}".format(
        pv=PREVIOUS_VERSION, cv=CURRENT_VERSION)) != -1

    worker = outs["my-worker-0"]
    assert worker.find(
        "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
    assert worker.find("Latest Kubernetes version: {cv}".format(
        cv=CURRENT_VERSION)) != -1
    # If the control plane nodes are not upgraded yet, skuba disallows upgrading a worker
    assert worker.find("Node my-worker-0 is up to date")


def test_upgrade_apply_all_fine(setup, platform, skuba):
    """
    Starting from a up-to-date cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(skuba)

    # node upgrade apply
    outs = {}
    for (r, n) in [("master", 0), ("worker", 0)]:
        node = "my-{}-{}".format(r, n)
        outs[node] = skuba.node_upgrade("apply", r, n)

    master = outs["my-master-0"]
    assert master.find(
        "Node my-master-0 is up to date"
    ) != -1

    worker = outs["my-worker-0"]
    assert worker.find(
        "Node my-worker-0 is up to date"
    ) != -1


def test_upgrade_apply_from_previous(setup, platform, skuba):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(skuba, PREVIOUS_VERSION)

    outs = {}
    for (r, n) in [("master", 0), ("worker", 0)]:
        node = "my-{}-{}".format(r, n)
        outs[node] = skuba.node_upgrade("apply", r, n)

    master = outs["my-master-0"]
    assert master.find("successfully upgraded") != -1

    worker = outs["my-worker-0"]
    assert worker.find("successfully upgraded") != -1


def test_upgrade_apply_user_lock(setup, platform, kubectl, skuba):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(skuba, PREVIOUS_VERSION)

    # lock kured
    kubectl.lock_kured()

    outs = {}
    for (r, n) in [("master", 0), ("worker", 0)]:
        node = "my-{}-{}".format(r, n)
        # disable skuba-update.timer
        platform.ssh_run(r, n, "sudo systemctl disable --now skuba-update.timer")
        outs[node] = skuba.node_upgrade("apply", r, n)
        assert platform.ssh_run(r, n, "sudo systemctl is-enabled skuba-update.timer || :").find("disabled") != -1

    assert kubectl.check_kured_lock() == True

    master = outs["my-master-0"]
    assert master.find("successfully upgraded") != -1

    worker = outs["my-worker-0"]
    assert worker.find("successfully upgraded") != -1
