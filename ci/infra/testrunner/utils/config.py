import os
import string
import sys

import yaml

from utils.format import Format

class dict_with_default(dict):
    def __init__(self, values, default):
        self.default = default
        super().__init__(values)

    def __missing__(self, key):
        return self.default

class Constant:
    TERRAFORM_EXAMPLE = "terraform.tfvars.json.ci.example"
    TERRAFORM_JSON_OUT = "tfout.json"
    SSH_OPTS = "-oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null " + \
               "-oConnectTimeout=60 -oBatchMode=yes "


class BaseConfig:

    def __new__(cls, yaml_path, *args, **kwargs):
        obj = super().__new__(cls, *args, **kwargs)
        obj.yaml_path = yaml_path
        obj.platform  = BaseConfig.Platform()
        obj.terraform = BaseConfig.Terraform()
        obj.openstack = BaseConfig.Openstack()
        obj.vmware = BaseConfig.VMware()
        obj.libvirt = BaseConfig.Libvirt()
        obj.skuba = BaseConfig.Skuba()
        obj.test = BaseConfig.Test()
        obj.log = BaseConfig.Log()
        obj.packages = BaseConfig.Packages()
        obj.kubectl = BaseConfig.Kubectl()
        obj.utils = BaseConfig.Utils()

        # vars get the values from yaml file
        vars = BaseConfig.get_var_dict(yaml_path)
        # conf.objects will be overriden by the values from vars and matching ENV variables
        BaseConfig.inject_attrs_from_yaml(obj, vars, "")
        # Final modification for conf variables
        BaseConfig.finalize(obj)
        BaseConfig.verify(obj)
        return obj

    class NodeConfig:
        def __init__(self, count=1, memory=4096, cpu=4):
            super().__init__()
            self.count = count
            self.memory = memory  # MB
            self.cpu = cpu
            self.ips = []
            self.external_ips = []

    class Utils:
        def __init__(self):
            self.ssh_sock =  "/tmp/testrunner_ssh_sock"
            self.ssh_key = "$HOME/.ssh/id_rsa"
            self.ssh_user = "sles"

    class Platform:
        def __init__(self):
            self.log_dir = "$WORKSPACE/platform_logs"

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
            self.workdir = "$WORKSPACE"
            self.tfdir = "$WORKSPACE/skuba/ci/infra"
            self.tfvars = Constant.TERRAFORM_EXAMPLE
            self.plugin_dir = None
            self.lb = BaseConfig.NodeConfig()
            self.master = BaseConfig.NodeConfig()
            self.worker = BaseConfig.NodeConfig()

    class Skuba:
        def __init__(self):
            super().__init__()
            self.workdir = "$WORKSPACE"
            self.cluster = "test-cluster"
            self.binpath = "$WORKSPACE/go/bin/skuba"
            self.verbosity = 5

    class Kubectl:
        def __init__(self):
            super().__init__()
            self.workdir = "$WORKSPACE"
            self.cluster = "test-cluster"
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
            self.file = "$WORKSPACE/testrunner.log"
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
    def print(config, level=0, out=sys.stdout):
        """ Prints the configuration
        """
        print(f'{"  "*level}{config.__class__.__name__}:', file=out)
        for key, value in config.__dict__.items():
            if isinstance(value, BaseConfig.config_classes):
                BaseConfig.print(value, level=level+1, out=out)
                continue

            print(f'{"  "*(level+1)}{key}: {value}', file=out)

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
    def inject_attrs_from_yaml(config, vars, ctx):
        """ Set values for configuration attributes
        The order of precedence is:
        - An environment variable exists with the fully qualified name of the
          attribute (e.g SKUBA_BINPATH)
        - The attribute from vars
        - default value for configuration

        After the attribute's value is set, a environement variables in the
        value are expanded.
        """

        for key, value in config.__dict__.items():
            if isinstance(value, BaseConfig.config_classes):
                sub_ctx = "{}{}".format(ctx+"_" if ctx else "", key)
                sub_config = value
                sub_vars = vars.get(key) if vars else None
                BaseConfig.inject_attrs_from_yaml(sub_config, sub_vars, sub_ctx)
                continue

            env_key = "{}{}".format(ctx+"_" if ctx else "", key).upper()
            env_value = os.getenv(env_key)

            # If env variable exists, use it. If not, use value fom vars, if
            # it exists (otherwise, default value from BaseConfig class will be
            # used)
            if env_value:
                value = env_value
            elif vars and key in vars:
                value = vars[key]

            config.__dict__[key] = BaseConfig.substitute(value)

    @staticmethod
    def substitute(value):
        """subtitute environment variables in the value of the attribute
           recursively substitute values in list or maps
        """
        if value is not None:
            if type(value) == str:
                value = string.Template(value).safe_substitute(dict_with_default(os.environ, ''))
            elif type(value) == list:
                value = [BaseConfig.substitute(e) for e in value]
            elif type(value) == dict:
                value = { k: BaseConfig.substitute(v) for k,v in value.items()}
        return value

    @staticmethod
    def finalize(conf):
        """ Finalize configuration.
        Deprecated. Will be removed
        """
        return

    @staticmethod
    def verify(conf):
        """ Validates configuration.
        Deprecated. Will be removed
        """
        return


    config_classes = (
        Platform,
        Packages,
        NodeConfig,
        Test,
        Log,
        Openstack,
        Terraform,
        Skuba,
        Kubectl,
        VMware,
        Libvirt,
        Utils
    )


