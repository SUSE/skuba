import pytest
from tests.utils import wait


@pytest.mark.disruptive
def test_remove_worker(deployment, conf, platform, skuba, kubectl):
    workers = kubectl.get_node_names_by_role("worker")
    workers_count = len(workers)

    # Remove the worker
    skuba.node_remove(role="worker", nr=workers_count - 1)
    assert len(kubectl.get_node_names_by_role("worker")) == workers_count - 1

    wait(kubectl.run_kubectl, 'wait --timeout=5m --for=condition=Ready pods --all --namespace=kube-system', wait_delay=60, wait_timeout=300, wait_backoff=30, wait_retries=5, wait_allow=(RuntimeError))
