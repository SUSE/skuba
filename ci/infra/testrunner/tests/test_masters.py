import pytest
from tests.utils import wait


@pytest.mark.disruptive
def test_remove_master(deployment, conf, platform, skuba, kubectl):
    initial_masters = skuba.num_of_nodes("master")

    # Remove the master
    skuba.node_remove(role="master", nr=initial_masters - 1)
    assert skuba.num_of_nodes("master") == initial_masters - 1

    wait(kubectl.run_kubectl, 'wait --timeout=5m --for=condition=Ready pods --all --namespace=kube-system  --field-selector=status.phase!=Succeeded', wait_delay=60, wait_timeout=300, wait_backoff=30, wait_retries=5, wait_allow=(RuntimeError))
