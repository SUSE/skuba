import pytest
from skuba import Skuba
from platforms import Platform
from utils import BaseConfig
from tests import TestDriver

def pytest_addoption(parser):
    """
    Adds the option pytest option list.
    This options can be used to initilize fixtures.
    """
    parser.addoption(
        "--vars",
        action="store",
        help="vars yaml",
    )

@pytest.fixture
def conf(request):
    """Builds a conf object from a yaml file"""
    path = request.config.getoption("vars")
    return BaseConfig(path)

@pytest.fixture
def skuba(conf):
    return Skuba(conf)

@pytest.fixture
def platform(conf):
    platform = Platform.get_platform(request.param)
    platform.provision()
    def cleanup():
        platform.cleanup()

    request.addfinalizer(cleanup)
    return platform
