import pytest
import time

@pytest.fixture
def setup(request, platform, skuba):
    platform.provision()
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)

    skuba.cluster_init()
    skuba.node_bootstrap()

def test_add_worker(setup, skuba):
    skuba.node_join(role="worker", nr=0)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 1
