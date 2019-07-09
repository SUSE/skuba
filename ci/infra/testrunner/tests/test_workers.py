from platforms import Platform
from skuba import Skuba
import pytest
import time

@pytest.fixture
def setup(request, conf):
    platform = Platform.get_platform(conf) 
    platform.provision()
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)

    skuba = Skuba(conf)
    skuba.cluster_init()
    time.sleep(120)
    skuba.node_bootstrap()


def test_worker(setup, conf):
    skuba = Skuba(conf)
    skuba.node_join(role="worker", nr=0)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 1
    skuba.node_remove(role="worker", nr=0)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 0

