import pytest
from tests.utils import wait


@pytest.mark.disruptive
def test_add_worker(bootstrap, skuba):
    skuba.node_join(role="worker", nr=0)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 1


@pytest.mark.disruptive
def test_remove_worker(deployment, conf, platform, skuba, kubectl):
    initial_workers = skuba.num_of_nodes("worker")

    # Remove the worker
    skuba.node_remove(role="worker", nr=initial_workers - 1)
    assert skuba.num_of_nodes("worker") == initial_workers - 1

    wait(kubectl.run_kubectl, 'wait --timeout=5m --for=condition=Ready pods --all --namespace=kube-system', wait_delay=60, wait_timeout=300, wait_backoff=30, wait_retries=5, wait_allow=(RuntimeError))
