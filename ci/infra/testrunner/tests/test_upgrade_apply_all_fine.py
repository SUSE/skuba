import pytest

from tests.utils import setup_kubernetes_version


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
