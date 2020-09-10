import os
from unittest import mock
from utils.config import BaseConfig

user = "USERID"
home = "/path/to/home"
workspace = "/path/to/workspace"
env_workspace = "/path/to/workspace/from/env"
skuba_binpath = "go/bin/skuba"
terraform_tfdir = "ci/infra"
ssh_key = ".ssh/id_rsa"
tfvars = "terraform.tfvars.json.ci.example"


empty_yaml = """
""".format(workspace=workspace)


def test_defaults():
    """Test defaults according to the [documentation](../README.md)
    """
    with mock.patch("builtins.open", mock.mock_open(read_data=empty_yaml)), \
         mock.patch.dict("os.environ", clear=True,
                         values={"WORKSPACE": workspace,
                                 "HOME": home,
                                 "USER": user}):
        config = BaseConfig("vars.yaml")
        assert config.skuba.workdir == workspace
        assert config.skuba.binpath == os.path.join(workspace, skuba_binpath)
        assert config.skuba.cluster == "test-cluster"
        assert config.terraform.tfdir == os.path.join(workspace, terraform_tfdir)
        assert config.terraform.stack_name == user
        assert config.terraform.workdir == workspace
        assert config.terraform.tfvars == tfvars
        assert config.terraform.plugin_dir is None
        assert config.utils.ssh_key == os.path.join(home, ssh_key)


subs_yaml = """
packages:
  additional_pkgs:
  - $MYPACKAGE
  additional_repos:
    my_repo: $MYREPO
"""
my_package = "my-repo"
my_repo = "http://url/to/my/repo"


def test_substitutions():
    """Test substitution of environment variables in lists and  maps
    """
    with mock.patch("builtins.open", mock.mock_open(read_data=subs_yaml)), \
         mock.patch.dict("os.environ", clear=True,
                         values={"MYPACKAGE": my_package,
                                 "MYREPO": my_repo}):
        config = BaseConfig("vars.yaml")
        assert len(config.packages.additional_pkgs) == 1
        assert config.packages.additional_pkgs[0] == my_package
        assert config.packages.additional_repos["my_repo"] == my_repo
