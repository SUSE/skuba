import os
import string

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
        obj.workspace = "$HOME/workspace"
        obj.log_dir = "$WORKSPACE/platform_logs"

        obj.terraform = BaseConfig.Terraform()
        obj.openstack = BaseConfig.Openstack()
        obj.vmware = BaseConfig.VMware()
        obj.libvirt = BaseConfig.Libvirt()
        obj.skuba = BaseConfig.Skuba()
        obj.test = BaseConfig.Test()
        obj.log = BaseConfig.Log()
        obj.packages = BaseConfig.Packages()
        obj.kubectl = BaseConfig.Kubectl()

        config_classes = (
            BaseConfig.Packages,
            BaseConfig.NodeConfig,
            BaseConfig.Test,
            BaseConfig.Log,
            BaseConfig.Openstack,
            BaseConfig.Terraform,
            BaseConfig.Skuba,
            BaseConfig.Kubectl,
            BaseConfig.VMware,
            BaseConfig.Libvirt
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
            self.stack_name = "$USER"
            self.tfdir = "$WORKSPACE/skuba/ci/infra"
            self.tfvars = Constant.TERRAFORM_EXAMPLE
            self.plugin_dir = None
            self.ssh_key = "$HOME/.ssh/id_rsa"
            self.nodeuser = None
            self.lb = BaseConfig.NodeConfig()
            self.master = BaseConfig.NodeConfig()
            self.worker = BaseConfig.NodeConfig()

    class Skuba:
        def __init__(self):
            super().__init__()
            self.binpath = "$WORKSPACE/go/bin/skuba"
            self.verbosity = 5

    class Kubectl:
        def __init__(self):
            super().__init__()
            self.binpath = "/usr/bin/kubectl"
            self.kubeconfig = "$WORKSPACE/test-cluster/admin.conf"

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

    class Libvirt:
        def __init__(self):
            super().__init__()
            self.uri = "qemu:///system"
            self.keyfile = None
            self.image_uri = None

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
        """ Set values for configuration attributes
        The order of precedence is:
        - An environment variable exists with the fully qualified name of the
          attribute
        - The attribute from vars
        - default value for configuration

        After the attribute's value is set, a environement variables in the
        value are expanded.
        """
        for key, value in obj.__dict__.items():
            if isinstance(value, config_classes):
                config_class = obj.__dict__[key]
                BaseConfig._set_config_class_attrs(config_class, key, vars, config_classes)
                continue

            env_key = key.upper()
            env_value = os.getenv(env_key)

            # If env variable exists, use it. If not, use value fom vars, if
            # it exists (otherwise, default value from BaseConfig class will be
            # used)
            if env_value:
                value = env_value
            elif key in vars:
                value = vars[key]

            # subtitute environment variables in the value of the attribute
            if value is not None:
                obj.__dict__[key] = string.Template(value).substitute(os.environ)

        return obj

    @staticmethod
    def finalize(conf):
        """ Finalize configuration.
            Deprecated. Will be removed
        """
        return conf

    @staticmethod
    def verify(conf):
        if not conf.workspace and conf.workspace == "":
            raise ValueError(Format.alert("You should set the workspace value in a configured yaml file e.g. vars.yaml"
                                          " or set env var WORKSPACE before using testrunner)"))

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
