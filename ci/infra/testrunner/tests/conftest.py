import pytest

import platforms
from kubectl import Kubectl
from skuba import Skuba
from utils import BaseConfig
from tests.utils import (check_pods_ready, wait)


def pytest_addoption(parser):
    """
    Adds the option pytest option list.
    This options can be used to initilize fixtures.
    """
    parser.addoption("--vars", action="store", help="vars yaml")
    parser.addoption("--platform", action="store", help="target platform")
    parser.addoption("--skip-setup",
                     choices=['provisioned', 'bootstrapped', 'deployed'],
                     help="Skip the given setup step.\n"
                          "'provisioned' For when you have already provisioned the nodes.\n"
                          "'bootstrapped' For when you have already bootstrapped the cluster.\n"
                          "'deployed' For when you already have a fully deployed cluster.")


@pytest.fixture
def provision(request, platform):
    if request.config.getoption("skip_setup") in ['provisioned', 'bootstrapped', 'deployed']:
        return

    def cleanup():
        platform.gather_logs()
        platform.cleanup()

    request.addfinalizer(cleanup)

    platform.provision()


@pytest.fixture
def bootstrap(request, provision, skuba):

    if request.config.getoption("skip_setup") in ['bootstrapped', 'deployed']:
        return

    skuba.cluster_init()
    skuba.node_bootstrap()


@pytest.fixture
def deployment(request, bootstrap, skuba, kubectl):
    if request.config.getoption("skip_setup") == 'deployed':
        return

    skuba.join_nodes()
    wait(check_pods_ready,
         kubectl,
         namespace="kube-system",
         wait_delay=60,
         wait_timeout=10,
         wait_backoff=60,
         wait_elapsed=60 * 30,
         wait_allow=(AssertionError))


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
def kubectl(conf):
    return Kubectl(conf)


@pytest.fixture
def platform(conf, target):
    platform = platforms.get_platform(conf, target)
    return platform


@pytest.fixture
def setup(request, platform, skuba):
    def cleanup():
        # if platform was not allocated, gather_logs may fail. Ignore.
        try:
            platform.gather_logs()
        finally:
            platform.cleanup()

    request.addfinalizer(cleanup)

    platform.provision(num_master=3, num_worker=3)
