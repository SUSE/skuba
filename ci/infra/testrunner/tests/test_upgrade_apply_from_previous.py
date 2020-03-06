import pytest

from tests.utils import PREVIOUS_VERSION, setup_kubernetes_version, node_is_ready, node_is_upgraded


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