import os
from unittest import mock
from utils.config import BaseConfig

user = "USERID"
home = "/path/to/home"
workspace = "/path/to/workspace"
env_workspace = "/path/to/workspace/from/env"
skuba_binpath = "go/bin/skuba"
terraform_tfdir = "skuba/ci/infra"
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
                                 "USER": user
                         }):
        config = BaseConfig("vars.yaml")
        assert config.skuba.workdir == workspace
        assert config.skuba.binpath == os.path.join(workspace,skuba_binpath)
        assert config.skuba.cluster == "test-cluster"
        assert config.terraform.tfdir == os.path.join(workspace,terraform_tfdir)
        assert config.terraform.stack_name == user
        assert config.terraform.workdir ==  workspace
        assert config.terraform.tfvars ==  tfvars
        assert config.terraform.plugin_dir == None
        assert config.utils.ssh_key == os.path.join(home, ssh_key)


