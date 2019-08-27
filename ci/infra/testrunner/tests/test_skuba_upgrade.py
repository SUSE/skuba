import pytest
import time


PREVIOUS_VERSION = "1.14.1"
CURRENT_VERSION = "1.15.2"


@pytest.fixture
def setup(request, platform, skuba):
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)

    platform.provision(num_master=3, num_worker=3)


def setup_kubernetes_version(platform, skuba, kubectl, kubernetes_version=None):
    """
    Initialize the cluster with the given kubernetes_version, bootstrap it and
    join nodes.
    """

    skuba.cluster_init(kubernetes_version)
    skuba.node_bootstrap()

    masters = platform.get_num_nodes("master")
    for n in range (1, masters):
        skuba.node_join("master", n)

    workers = platform.get_num_nodes("worker")
    for n in range (0, workers):
        skuba.node_join("worker", n)

    kubectl.run_kubectl("wait --for=condition=ready nodes --all --timeout=5m")


def node_is_upgraded(kubectl, platform, node_name):
    for attempt in range(10):
        if platform.all_apiservers_responsive():
            # kubernetes might be a little bit slow with updating the NodeVersionInfo
            version = kubectl.run_kubectl("get nodes {} -o jsonpath='{{.status.nodeInfo.kubeletVersion}}'".format(node_name))
            if version.find(PREVIOUS_VERSION) == 0:
                time.sleep(2)
            else:
                break
        else:
            time.sleep(2)
    return kubectl.run_kubectl("get nodes {} -o jsonpath='{{.status.nodeInfo.kubeletVersion}}'".format(node_name)).find(CURRENT_VERSION) != -1


def node_is_ready(kubectl, node_name):
    return kubectl.run_kubectl("get nodes {} -o jsonpath='{{range @.status.conditions[*]}}{{@.type}}={{@.status}};{{end}}'".format(node_name)).find("Ready=True") != -1


@pytest.mark.disruptive
def test_upgrade_plan_all_fine(setup, skuba, kubectl, platform):
    """
    Starting from a up-to-date cluster, check what cluster/node plan report.
    """

    setup_kubernetes_version(platform, skuba, kubectl)
    out = skuba.cluster_upgrade_plan()

    assert out.find(
        "Congratulations! You are already at the latest version available"
    ) != -1


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
    for n in range (0, masters):
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
    for n in range (0, workers):
        worker = skuba.node_upgrade("plan", "worker", n)
        assert worker.find(
            "Current Kubernetes cluster version: {pv}".format(pv=PREVIOUS_VERSION))
        assert worker.find("Latest Kubernetes version: {cv}".format(
            cv=CURRENT_VERSION)) != -1
        # If the control plane nodes are not upgraded yet, skuba disallows upgrading a worker
        assert worker.find("Node my-worker-{} is up to date".format(n))


@pytest.mark.disruptive
def test_upgrade_apply_all_fine(setup, platform, skuba, kubectl):
    """
    Starting from a up-to-date cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl)

    # node upgrade apply
    masters = platform.get_num_nodes("master")
    for n in range (0, masters):
        master = skuba.node_upgrade("plan", "master", n)
        assert master.find(
            "Node my-master-{} is up to date".format(n)
        ) != -1

    workers = platform.get_num_nodes("worker")
    for n in range (0, workers):
        worker = skuba.node_upgrade("plan", "worker", n)
        assert worker.find(
            "Node my-worker-{} is up to date".format(n)
        ) != -1


@pytest.mark.disruptive
def test_upgrade_apply_from_previous(setup, platform, skuba, kubectl):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

    masters = platform.get_num_nodes("master")
    for n in range (0, masters):
        node = "my-master-{}".format(n)
        assert node_is_ready(kubectl, node)
        master = skuba.node_upgrade("apply", "master", n)
        assert master.find("successfully upgraded") != -1
        assert node_is_upgraded(kubectl, platform, node)

    workers = platform.get_num_nodes("worker")
    for n in range (0, workers):
        node = "my-worker-{}".format(n)
        worker = skuba.node_upgrade("apply", "worker", n)
        assert worker.find("successfully upgraded") != -1
        assert node_is_upgraded(kubectl, platform, node)


@pytest.mark.disruptive
def test_upgrade_apply_user_lock(setup, platform, kubectl, skuba):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

    # lock kured
    kubectl.run_kubectl("-n kube-system annotate ds kured weave.works/kured-node-lock='{\"nodeID\":\"manual\"}'")

    masters = platform.get_num_nodes("master")
    for n in range (0, masters):
        node = "my-master-{}".format(n)
        # disable skuba-update.timer
        platform.ssh_run("master", n, "sudo systemctl disable --now skuba-update.timer")
        assert node_is_ready(kubectl, node)
        master = skuba.node_upgrade("apply", "master", n)
        assert master.find("successfully upgraded") != -1
        assert node_is_upgraded(kubectl, platform, node)
        assert platform.ssh_run("master", n, "sudo systemctl is-enabled skuba-update.timer || :").find("disabled") != -1

    workers = platform.get_num_nodes("worker")
    for n in range (0, workers):
        node = "my-worker-{}".format(n)
        # disable skuba-update.timer
        platform.ssh_run("worker", n, "sudo systemctl disable --now skuba-update.timer")
        assert node_is_ready(kubectl, node)
        worker = skuba.node_upgrade("apply", "worker", n)
        assert worker.find("successfully upgraded") != -1
        assert node_is_upgraded(kubectl, platform, node)
        assert platform.ssh_run("worker", n, "sudo systemctl is-enabled skuba-update.timer || :").find("disabled") != -1

    assert kubectl.run_kubectl(r"-n kube-system get ds/kured -o jsonpath='{.metadata.annotations.weave\.works/kured-node-lock}'").find("manual") != -1
