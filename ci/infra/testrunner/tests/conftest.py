import pytest
import platforms
from skuba import Skuba
from kubectl import Kubectl
from utils import BaseConfig

def pytest_addoption(parser):
    """
    Adds the option pytest option list.
    This options can be used to initilize fixtures.
    """
    parser.addoption("--vars", action="store",help="vars yaml" )
    parser.addoption("--platform", action="store",help="target platform" )

@pytest.fixture
def provision(request, platform):
    platform.provision()
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)


@pytest.fixture
def bootstrap(provision, skuba):
    skuba.cluster_init()
    skuba.node_bootstrap()


@pytest.fixture
def deployment(bootstrap, platform, skuba):
    masters = platform.get_num_nodes("master")
    for n in range (1, masters):
        skuba.node_join("master", n)

    workers = platform.get_num_nodes("worker")
    for n in range (0, workers):
        skuba.node_join("worker", n)


@pytest.fixture
def conf(request):
    """Builds a conf object from a yaml file"""
    path = request.config.getoption("vars")
    return BaseConfig(path)


@pytest.fixture
def target(request):
    """Returns the target platform"""
    platform = request.config.getoption("platform")
    return platform


@pytest.fixture
def skuba(conf, target):
    return Skuba(conf, target)


@pytest.fixture
def kubectl(conf, target):
    return Kubectl(conf, target)


@pytest.fixture
def platform(conf, target):
    platform = platforms.get_platform(conf, target)
    return platform
