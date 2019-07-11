from platforms import Platform
from skuba import Skuba
from kubectl import Kubectl
import pytest
import time

@pytest.fixture
def setup(request, conf):
    platform = Platform.get_platform(conf) 
    platform.provision()
    # def cleanup():
    #     platform.cleanup()
    # request.addfinalizer(cleanup)

    skuba = Skuba(conf)
    skuba.cluster_init()
    time.sleep(120)
    skuba.node_bootstrap()


def test_worker(setup, conf):
    skuba = Skuba(conf)
    kubectl = Kubectl(conf)
    skuba.node_join(role="worker", nr=0)
    skuba.node_join(role="worker", nr=1)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 2
    # skuba.node_remove(role="worker", nr=0)
    # masters = skuba.num_of_nodes("master")
    # workers = skuba.num_of_nodes("worker")
    # assert masters == 1
    # assert workers == 0
    kubectl.get_nodes(conf)
    kubectl.get_pods(conf)
    kubectl.create_deployment(conf, "nginx", "nginx:stable-alpine")
    kubectl.scale_deployment(conf, "nginx", workers)
    kubectl.expose_deployment(conf, "nginx", 80)
    kubectl.wait_deployment(conf, "nginx", 3)
    kubectl.get_pods(conf)
    assert kubectl.count_available_replicas(conf, "nginx") == workers

