import pytest
from tests.utils import wait


@pytest.mark.disruptive
def test_remove_master(deployment, conf, platform, skuba, kubectl):
    masters = kubectl.get_node_names_by_role("master")
    masters_count = len(masters)

    # Remove the master
    skuba.node_remove(role="master", nr=masters_count - 1)
    assert len(kubectl.get_node_names_by_role("master")) == masters_count - 1

    wait(kubectl.run_kubectl, 'wait --timeout=5m --for=condition=Ready pods --all --namespace=kube-system  --field-selector=status.phase!=Succeeded', wait_delay=60, wait_timeout=300, wait_backoff=30, wait_retries=5, wait_allow=(RuntimeError))
