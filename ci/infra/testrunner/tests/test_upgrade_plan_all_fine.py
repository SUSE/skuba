import pytest


@pytest.mark.disruptive
def test_upgrade_plan_all_fine(provision, skuba, kubectl, platform):
    """
    Starting from a up-to-date cluster, check what cluster/node plan report.
    """

    out = skuba.cluster_upgrade(action="plan")

    assert out.find(
        "All nodes match the current cluster version"
    ) != -1
