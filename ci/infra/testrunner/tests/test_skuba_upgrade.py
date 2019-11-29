import time
import pytest

from tests.utils import (check_nodes_ready, wait)

PREVIOUS_VERSION = "1.15.2"
CURRENT_VERSION = "1.16.2"

@pytest.fixture
def setup(request, platform, skuba):
    def cleanup():
        platform.gather_logs()
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

    skuba.join_nodes()

    wait(check_nodes_ready,
         kubectl,
         wait_delay=60,
         wait_backoff=30,
         wait_elapsed=60*10,
         wait_allow=(AssertionError))


def node_is_upgraded(kubectl, platform, role, nr):
    node_name = platform.get_nodes_names(role)[nr]
    for attempt in range(20):
        if platform.all_apiservers_responsive():
            # kubernetes might be a little bit slow with updating the NodeVersionInfo
            cmd = ("get nodes {} -o jsonpath="
                   "'{{.status.nodeInfo.kubeletVersion}}'").format(node_name)
            version = kubectl.run_kubectl(cmd)
            if version.find(PREVIOUS_VERSION) == 0:
                time.sleep(2)
            else:
                break
        else:
            time.sleep(2)

    cmd = "get nodes {} -o jsonpath='{{.status.nodeInfo.kubeletVersion}}'".format(node_name)
    return kubectl.run_kubectl(cmd).find(CURRENT_VERSION) != -1


def node_is_ready(platform, kubectl, role, nr):
    node_name = platform.get_nodes_names(role)[nr]
    cmd = ("get nodes {} -o jsonpath='{{range @.status.conditions[*]}}"
           "{{@.type}}={{@.status}};{{end}}'").format(node_name)

    return kubectl.run_kubectl(cmd).find("Ready=True") != -1



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


@pytest.mark.disruptive
def test_upgrade_apply_all_fine(setup, platform, skuba, kubectl):
    """
    Starting from a up-to-date cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl)

    # node upgrade apply
    masters = platform.get_num_nodes("master")
    master_names = platform.get_nodes_names("master")
    for n in range(0, masters):
        master = skuba.node_upgrade("plan", "master", n)
        assert master.find(
            f'Node {master_names[n]} is up to date'
        ) != -1

    workers = platform.get_num_nodes("worker")
    workers_names = platform.get_nodes_names("worker")
    for n in range(0, workers):
        worker = skuba.node_upgrade("plan", "worker", n)
        assert worker.find(
            f'Node {workers_names[n]} is up to date'
        ) != -1


@pytest.mark.disruptive
def test_upgrade_apply_from_previous(setup, platform, skuba, kubectl):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

    for role in ("master", "worker"):
        num_nodes = platform.get_num_nodes(role)
        for n in range(0, num_nodes):
            assert node_is_ready(platform, kubectl, role, n)
            result = skuba.node_upgrade("apply", role, n)
            assert result.find("successfully upgraded") != -1
            assert node_is_upgraded(kubectl, platform, role, n)


@pytest.mark.disruptive
def test_upgrade_apply_user_lock(setup, platform, kubectl, skuba):
    """
    Starting from an outdated cluster, check what node upgrade apply reports.
    """

    setup_kubernetes_version(platform, skuba, kubectl, PREVIOUS_VERSION)

    # lock kured
    kubectl_cmd = (
        "-n kube-system annotate ds kured weave.works/kured-node-lock="
        "'{\"nodeID\":\"manual\"}'")
    kubectl.run_kubectl(kubectl_cmd)

    for role in ("master", "worker"):
        num_nodes = platform.get_num_nodes(role)
        for n in range(0, num_nodes):
            # disable skuba-update.timer
            platform.ssh_run(role, n, "sudo systemctl disable --now skuba-update.timer")
            assert node_is_ready(platform, kubectl, role, n)
            result = skuba.node_upgrade("apply", role, n)
            assert result.find("successfully upgraded") != -1
            assert node_is_upgraded(kubectl, platform, role, n)
            ssh_cmd = "sudo systemctl is-enabled skuba-update.timer || :"
            assert platform.ssh_run(role, n, ssh_cmd).find("disabled") != -1

    kubectl_cmd = (r"-n kube-system get ds/kured -o jsonpath="
                   r"'{.metadata.annotations.weave\.works/kured-node-lock}'")
    result = wait(kubectl.run_kubectl,
                  kubectl_cmd,
                  wait_backoff=30,
                  wait_retries=3,
                  wait_allow=(RuntimeError))
    assert result.find("manual") != -1
