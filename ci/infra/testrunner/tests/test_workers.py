from platforms import Platform
from skuba import Skuba
from kubectl import Kubectl
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

#def test_add_worker(setup, skuba):
#    skuba.node_join(role="worker", nr=0)
#    masters = skuba.num_of_nodes("master")
#    workers = skuba.num_of_nodes("worker")
#    assert masters == 1
#    assert workers == 1

def test_nginx_deployment(setup, skuba, kubectl):
    skuba.node_join(role="worker", nr=0)
    skuba.node_join(role="worker", nr=1)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 2
    kubectl.get_pods()
    kubectl.create_deployment("nginx", "nginx:stable-alpine")
    kubectl.scale_deployment("nginx", workers)
    kubectl.expose_deployment("nginx", 80)
    kubectl.wait_deployment("nginx", 3)
    kubectl.get_pods()
    assert kubectl.count_available_replicas("nginx") == workers
    result = kubectl.test_service("nginx")
    assert "Welcome to nginx" in result

