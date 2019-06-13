import os
import yaml

from utils.format import Format


class Constant:
    TERRAFORM_EXAMPLE = "terraform.tfvars.json.ci.example"
    TERRAFORM_JSON_OUT = "tfout.json"
    SSH_OPTS = "-oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null " + \
           "-oConnectTimeout=60 -oBatchMode=yes "


class BaseConfig:

    def __new__(cls, yaml_path, *args, **kwargs):
        obj = super().__new__(cls, *args, **kwargs)
        obj.platform = None  #"openstack, vmware, bare-metal
        obj.workspace = None
        obj.terraform_json_path = None
        obj.ssh_key_option = None
        obj.username = None
        obj.nodeuser = None

        obj.terraform = BaseConfig.Terraform()
        obj.openstack = BaseConfig.Openstack()
        obj.vmware = BaseConfig.VMware()
        obj.skuba = BaseConfig.Skuba()

        obj.lb = BaseConfig.NodeConfig()
        obj.master = BaseConfig.NodeConfig()
        obj.worker = BaseConfig.NodeConfig()
        obj.test = BaseConfig.Test()

        config_classes = (
            BaseConfig.NodeConfig,
            BaseConfig.Test,
            BaseConfig.Openstack,
            BaseConfig.Terraform,
            BaseConfig.Skuba,
            BaseConfig.VMware
        )

        #vars get the values from yaml file
        vars = BaseConfig.get_var_dict(yaml_path)
        #conf.objects will be overriden by the values from vars and matching ENV variables
        conf = BaseConfig.inject_attrs_from_yaml(obj, vars, config_classes)
        # Final mofification for conf variables
        conf = BaseConfig.finalize(conf)
        conf = BaseConfig.verify(conf)
        return conf

    class NodeConfig:
        def __init__(self, count=1, memory=4096, cpu=4):
            super().__init__()
            self.count = count
            self.memory = memory  # MB
            self.cpu = cpu
            self.ips = []
            self.external_ips = []

    class Openstack:
        def __init__(self):
            super().__init__()
            self.openrc = None

    class Terraform:
        def __init__(self):
            super().__init__()
            self.internal_net = None
            self.stack_name = None
            self.tfdir = None
            self.tfvars = Constant.TERRAFORM_EXAMPLE 
            self.plugin_dir = None

    class Skuba:
        def __init__(self):
            super().__init__()
            self.binpath = None
            self.srcpath = None

    class Test:
        def __init__(self):
            super().__init__()
            self.replica_count = 5
            self.replicas_creation_interval_seconds = 5
            self.podname = "default"
            self.no_destroy = False

    class VMware:
        def __init__(self):
            self.env_file = None
            self.template_name = None


    @staticmethod
    def get_yaml_path(yaml_path):
        utils_dir = os.path.dirname(os.path.realpath(__file__))
        testrunner_dir = os.path.join(utils_dir, "..")
        config_yaml_file_path = os.path.join(testrunner_dir, yaml_path)
        return os.path.abspath(config_yaml_file_path)

    @staticmethod
    def get_var_dict(yaml_path):
        config_yaml_file_path = BaseConfig.get_yaml_path(yaml_path)
        if not os.path.isfile(config_yaml_file_path):
            print(Format.alert("You have incorrect -v path for xml file: {}".format(config_yaml_file_path)))
            raise FileNotFoundError

        with open(config_yaml_file_path, 'r') as stream:
            _conf = yaml.safe_load(stream)
        return _conf

    @staticmethod
    def inject_attrs_from_yaml(obj, vars, config_classes):
        for key, value in obj.__dict__.items():

            # FIXME: If key is the yaml file and is a config class, but has no subelements 
            # the logic fails. This happens if we comment all values under the key, for example:
            # key:
            # # subkey: value  #commented!
            #  
            if key in vars and isinstance(value, config_classes):
                BaseConfig.inject_attrs_from_yaml(value, vars[key], config_classes)
                continue

            new_key = key.upper()
            new_value = None

            # FIXME: the env variables must be looked as CLASS_KEY to prevent name collitions
            # with well known variables (e.g. PATH) or between classes
            if os.getenv(new_key):
                new_value = os.getenv(new_key)

            # if env variable does not exist, store
            if not new_value and key in vars:
                obj.__dict__[key] = vars[key]

            # if username env variable exist but do not update username
            if new_value:
                obj.__dict__[key] = new_value

            # username should get from xml config file
            if key == "username":
                obj.__dict__[key] = vars[key]

        return obj

    @staticmethod
    def finalize(conf):
        conf.workspace = os.path.expanduser(conf.workspace)

        if not conf.skuba.binpath:
            conf.skuba.binpath = os.path.join(conf.workspace, 'go/bin/skuba')

        if not conf.skuba.srcpath:
            conf.skuba.srcpath = os.path.realpath(os.path.join(conf.workspace, "skuba"))

        if not conf.terraform.tfdir:
           conf.terraform.tfdir= os.path.join(conf.skuba.srcpath, "ci/infra/")

        conf.terraform_json_path = os.path.join(conf.workspace, Constant.TERRAFORM_JSON_OUT)

        if not conf.terraform.stack_name:
            conf.terraform.stack_name = conf.username

        # TODO: This variable should be in openstack configuration but due to
        # the way variables are processed in Terraform class, must be here for know.
        if not conf.terraform.internal_net:
            conf.terraform.internal_net = conf.terraform.stack_name

        #TODO: add the path to shared ssh credentials as a configuration parameter
        if conf.ssh_key_option == "id_shared":
            conf.ssh_key_option = os.path.join(conf.skuba.srcpath, "ci/infra/id_shared")
        elif conf.ssh_key_option == "id_rsa":
            conf.ssh_key_option = os.path.join(os.path.expanduser("~"), ".ssh/id_rsa")

        return conf

    @staticmethod
    def verify(conf):
        if not conf.workspace and conf.workspace == "":
            raise ValueError(Format.alert("You should setup workspace value in a configured yaml file "
                                           "before using testrunner (skuba/ci/infra/testrunner/vars)"))
        if not conf.terraform.stack_name:
            raise ValueError(Format.alert("Either a terraform stack name or an user name must be specified"))

        if os.path.normpath(conf.workspace) == os.path.normpath((os.getenv("HOME"))):
            raise ValueError(Format.alert("workspace should not be your home directory"))

        return conf
#if __name__ == '__main__':
#    _conf = BaseConfig()
