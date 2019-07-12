import pytest
from skuba import Skuba
from kubectl import Kubectl
from platforms import Platform
from utils import BaseConfig

def pytest_addoption(parser):
    """
    Adds the option pytest option list.
    This options can be used to initilize fixtures.
    """
    parser.addoption("--vars", action="store",help="vars yaml" )
    parser.addoption("--platform", action="store",help="target platform" )

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
    platform = Platform.get_platform(conf, target)
    return platform
