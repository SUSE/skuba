import contextlib
import copy
import logging
import os
import pytest
import platforms
import tempfile
import shutil
from skuba import Skuba
from kubectl import Kubectl
from utils import BaseConfig


LOGGER = logging.getLogger('testrunner')
LOGGER.setLevel(logging.DEBUG)


def pytest_addoption(parser):
    """
    Adds the option pytest option list.
    This options can be used to initilize fixtures.
    """
    parser.addoption("--vars", action="store",help="vars yaml" )
    parser.addoption("--platform", action="store",help="target platform" )


class Workspace:
    def __init__(self):
        self.temporary_directory = tempfile.mkdtemp(prefix="skuba-tests")

    def cleanup(self):
        shutil.rmtree(self.temporary_directory)


class ClusterFixture(contextlib.AbstractContextManager):
    """Wrap cluster creation and bootstrapping into a context manager with state.

    This allows creating a cluster using this fixture that will keep track of it's
    own workspace:

    >>> with ClusterFixture(masters=1, workers=1, platform="openstack") as c:
    >>>     c.run("kubectl ...")

    Once 'c' goes out of scope the cluster directory from skuba will be removed and
    the terraform deployment will be destroyed.
    """
    def __init__(self, conf: BaseConfig, masters=1, workers=1, platform='openstack'):
        self.masters = masters
        self.workers = workers
        self.workspace = Workspace()
        self.config = BaseConfig(conf.yaml_path)
        self.__adjust_directories()
        self.platform = platforms.get_platform(self.config, platform)
        # Get the original platform configuration to copy over the configuration
        platforms.get_platform(conf, platform).copy_configuration(self.workspace.temporary_directory)
        self.skuba = Skuba(self.config, platform)

    def __enter__(self):
        LOGGER.info("Provisioning cluster")
        self.platform.provision(num_master=self.masters,
                                num_worker=self.workers)
        LOGGER.info("Initializing cluster")
        self.skuba.cluster_init()
        LOGGER.info("Bootstrap cluster")
        self.skuba.node_bootstrap()
        for i in range(self.masters):
            LOGGER.info(f"Join master {i}")
            self.skuba.node_join(role="master", nr=i)
        for i in range(self.workers):
            LOGGER.info(f"Join worker {i}")
            self.skuba.node_join(role="worker", nr=i)
        return super().__enter__()

    def __exit__(self, exc_type, exc_value, traceback):
        LOGGER.info("Cleanup cluster")
        self.skuba.cleanup(self.config)
        LOGGER.info("Cleanup terraform")
        # Right now there is no additional check if the terraform cleanup
        # worked, as it should throw an exception in case it fails breaking
        # here anyhow
        self.platform.cleanup()
        LOGGER.info("Cleanup workspace")
        self.workspace.cleanup()

    def __adjust_directories(self):
        """Reset the config workspace, to not create all cluster resources in the same
        directory.

        """
        LOGGER.debug(f"Adjusting workspace to {self.workspace.temporary_directory}")
        self.config.workspace = self.workspace.temporary_directory
        LOGGER.debug(f"Adjusting terraform directory to {self.workspace.temporary_directory}")
        self.config.terraform.tfdir = self.workspace.temporary_directory


@pytest.fixture
def cluster(conf: BaseConfig, target):
    """Wrap the Cluster creation class in a function, so the tests don't have to
    get the `conf` file.

    """
    def _(masters=1, workers=1, platform=target):
        return ClusterFixture(conf, masters, workers, platform)
    return _


# TODO: Deprecated. Remove from tests
@pytest.fixture
def setup(request, platform, skuba):
    platform.provision()
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)

    skuba.cluster_init()
    skuba.node_bootstrap()

@pytest.fixture
def provision(request, platform):
    platform.provision()
    def cleanup():
        platform.gather_logs()
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
