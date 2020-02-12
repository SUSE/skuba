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
        obj.yaml_path = yaml_path
        obj.workspace = None
        obj.username = None
        obj.log_dir = None

        obj.terraform = BaseConfig.Terraform()
        obj.openstack = BaseConfig.Openstack()
        obj.vmware = BaseConfig.VMware()
        obj.skuba = BaseConfig.Skuba()
        obj.test = BaseConfig.Test()
        obj.log = BaseConfig.Log()
        obj.packages = BaseConfig.Packages()

        config_classes = (
            BaseConfig.Packages,
            BaseConfig.NodeConfig,
            BaseConfig.Test,
            BaseConfig.Log,
            BaseConfig.Openstack,
            BaseConfig.Terraform,
            BaseConfig.Skuba,
            BaseConfig.VMware
        )

        # vars get the values from yaml file
        vars = BaseConfig.get_var_dict(yaml_path)
        # conf.objects will be overriden by the values from vars and matching ENV variables
        conf = BaseConfig.inject_attrs_from_yaml(obj, vars, config_classes)
        # Final modification for conf variables
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
            self.retries = 5
            self.internal_net = None
            self.stack_name = None
            self.tfdir = None
            self.tfvars = Constant.TERRAFORM_EXAMPLE
            self.plugin_dir = None
            self.ssh_key = None
            self.nodeuser = None
            self.lb = BaseConfig.NodeConfig()
            self.master = BaseConfig.NodeConfig()
            self.worker = BaseConfig.NodeConfig()

    class Skuba:
        def __init__(self):
            super().__init__()
            self.binpath = None
            self.srcpath = None
            self.verbosity = 5

    class Test:
        def __init__(self):
            super().__init__()
            self.no_destroy = False

    class Log:
        def __init__(self):
            super().__init__()
            self.level = "INFO"
            self.quiet = False
            self.file = "testrunner.log"
            self.overwrite = False

    class VMware:
        def __init__(self):
            self.env_file = None
            self.template_name = None


    class Packages:
        def __init__(self):
            self.mirror = None
            self.registry_code = None
            self.additional_repos = None
            self.additional_pkgs = None

    @staticmethod
    def get_yaml_path(yaml_path):
        utils_dir = os.path.dirname(os.path.realpath(__file__))
        testrunner_dir = os.path.join(utils_dir, "..")
        config_yaml_file_path = os.path.join(testrunner_dir, yaml_path)
        return os.path.abspath(config_yaml_file_path)

    @staticmethod
    def get_var_dict(yaml_path):
        config_yaml_file_path = BaseConfig.get_yaml_path(yaml_path)
        with open(config_yaml_file_path, 'r') as stream:
            _conf = yaml.safe_load(stream)
        return _conf

    @staticmethod
    def inject_attrs_from_yaml(obj, vars, config_classes):
        for key, value in obj.__dict__.items():
            if isinstance(value, config_classes):
                config_class = obj.__dict__[key]
                BaseConfig._set_config_class_attrs(config_class, key, vars, config_classes)
                continue

            env_key = key.upper()
            env_value = os.getenv(env_key)

            # if env variable exist but is not 'username' use env_value
            # username must get from config file
            if env_value and env_key != "USERNAME":
                obj.__dict__[key] = env_value
            elif key in vars:
                obj.__dict__[key] = vars[key]

        return obj

    @staticmethod
    def finalize(conf):
        conf.workspace = os.path.expanduser(conf.workspace)

        if not conf.log_dir:
            conf.log_dir = os.path.join(conf.workspace, 'platform_logs')
        elif not os.path.isabs(conf.log_dir):
            conf.log_dir = os.path.join(conf.workspace, conf.log_dir)

        if not conf.skuba.binpath:
            conf.skuba.binpath = os.path.join(conf.workspace, 'go/bin/skuba')

        if not conf.skuba.srcpath:
            conf.skuba.srcpath = os.path.realpath(os.path.join(conf.workspace, "skuba"))

        if not conf.terraform.tfdir:
            conf.terraform.tfdir = os.path.join(conf.skuba.srcpath, "ci/infra/")

        if not conf.terraform.stack_name:
            conf.terraform.stack_name = conf.username

        # TODO: This variable should be in openstack configuration but due to
        # the way variables are processed in Terraform class, must be here for know.
        if not conf.terraform.internal_net:
            conf.terraform.internal_net = conf.terraform.stack_name

        if not conf.terraform.ssh_key:
            conf.terraform.ssh_key = os.path.join(os.path.expanduser("~"), ".ssh/id_rsa")
        else:
            conf.terraform.ssh_key = os.path.expandvars(conf.terraform.ssh_key)

        return conf

    @staticmethod
    def verify(conf):
        if not conf.workspace and conf.workspace == "":
            raise ValueError(Format.alert("You should set the workspace value in a configured yaml file e.g. vars.yaml"
                                          " or set env var WORKSPACE before using testrunner)"))
        if not conf.terraform.stack_name:
            raise ValueError(Format.alert("Either a terraform stack name or an user name must be specified"))

        if os.path.normpath(conf.workspace) == os.path.normpath((os.getenv("HOME"))):
            raise ValueError(Format.alert("workspace should not be your home directory"))

        return conf

    @staticmethod
    def _set_config_class_attrs(config_class, class_name, variables, config_classes):
        config_obj = variables.get(class_name)

        if config_obj is not None:
            for k, v in config_obj.items():
                env_var = os.getenv(f"{class_name.upper()}_{k.upper()}")
                if env_var is not None:
                    config_class.__dict__[k] = env_var
                elif v:
                    if isinstance(config_class.__dict__[k], config_classes):
                        BaseConfig._set_config_class_attrs(config_class.__dict__[k], k, config_obj, config_classes)
                    else:
                        config_class.__dict__[k] = v
