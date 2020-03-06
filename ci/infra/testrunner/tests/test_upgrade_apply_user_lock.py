import pytest

from tests.utils import PREVIOUS_VERSION, setup_kubernetes_version, node_is_ready, node_is_upgraded, wait


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