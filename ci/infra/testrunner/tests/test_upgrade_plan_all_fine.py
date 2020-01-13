import pytest

from tests.utils import setup_kubernetes_version


@pytest.mark.disruptive
def test_upgrade_plan_all_fine(setup, skuba, kubectl, platform):
    """
    Starting from a up-to-date cluster, check what cluster/node plan report.
    """

    setup_kubernetes_version(platform, skuba, kubectl)
    out = skuba.cluster_upgrade_plan()

    assert out.find(
        "All nodes match the current cluster version"
    ) != -1
