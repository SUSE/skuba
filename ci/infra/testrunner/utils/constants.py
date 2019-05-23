import yaml, os
from format import Format

class Constant:
    TERRAFORM_EXAMPLE="terraform.tfvars.ci.example"
    TERRAFORM_JSON_OUT = "tfout.json"
    SSH_OPTS = "-oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null " + \
           "-oConnectTimeout=60 -oBatchMode=yes "


class BaseConfig:

    def __new__(cls, yaml_path, *args, **kwargs):
        obj = super().__new__(cls, *args, **kwargs)
        obj.platform = None  #"openstack, vmware, bare-metal
        obj.workspace = None
        obj.skuba_dir = None
        obj.terraform_dir = None
        obj.terraform_json_path = None
        obj.ssh_key_option = None
        obj.username = None
        obj.nodeuser = None

        obj.openstack = BaseConfig.Openstack()
        obj.jenkins = BaseConfig.Jenkins()
        obj.git = BaseConfig.Git()

        obj.lb = BaseConfig.NodeConfig()
        obj.master = BaseConfig.NodeConfig()
        obj.worker = BaseConfig.NodeConfig()
        obj.test = BaseConfig.Test()

        config_classes = (
            BaseConfig.NodeConfig,
            BaseConfig.Jenkins,
            BaseConfig.Test,
            BaseConfig.Git,
            BaseConfig.Openstack,
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

    class Jenkins:
        def __init__(self):
            super().__init__()
            self.job_name = None
            self.build_number = None
            self.run_name = None

    class Git:
        def __init__(self):
            super().__init__()
            self.change_author = None
            self.change_author_email = None
            self.github_token = None
            self.branch_name = "master"

    class Openstack:
        def __init__(self):
            super().__init__()
            self.openrc = None

    class Test:
        def __init__(self):
            super().__init__()
            self.replica_count = 5
            self.replicas_creation_interval_seconds = 5
            self.podname = "default"
            self.no_destroy = False


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

            if key in vars and isinstance(value, config_classes):
                BaseConfig.inject_attrs_from_yaml(value, vars[key], config_classes)
                continue

            new_key = key.upper()
            new_value = None

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
        conf.skuba_dir = os.path.realpath(os.path.join(conf.workspace, "skuba"))
        conf.terraform_dir = os.path.join(conf.skuba_dir, "ci/infra/{}".format(conf.platform))
        conf.terraform_json_path = os.path.join(conf.workspace, Constant.TERRAFORM_JSON_OUT)

        if not conf.jenkins.job_name:
            conf.jenkins.job_name = conf.username
        conf.jenkins.run_name = "{}-{}".format(conf.jenkins.job_name, str(conf.jenkins.build_number))

        if conf.ssh_key_option == "id_shared":
            conf.ssh_key_option = os.path.join(conf.skuba_dir, "ci/infra/id_shared")
        elif conf.ssh_key_option == "id_rsa":
            conf.ssh_key_option = os.path.join(os.path.expanduser("~"), ".ssh/id_rsa")

        conf.git.change_author = os.getenv('GIT_COMMITTER_NAME', 'CaaSP Jenkins')
        conf.git.change_author_email = os.getenv('GIT_COMMITTER_EMAIL', 'containers-bugowner@suse.de')

        return conf

    @staticmethod
    def verify(conf):
        if not conf.workspace and conf.workspace == "":
            raise ValueError(Format.alert("You should setup workspace value in a configured yaml file "
                                           "before using testrunner (skuba/ci/infra/testrunner/vars)"))
        if os.path.normpath(conf.workspace) == os.path.normpath((os.getenv("HOME"))):
            raise ValueError(Format.alert("workspace should not be your home directory"))
        if not os.path.exists(os.path.join(conf.workspace, "skuba")):
            raise ValueError(Format.alert("Your working directory, {} does not include \"skuba\" directory.\n\t    "
                                "Check your working directory in a configured yaml file".format(conf.workspace)))
        if conf.platform == "openstack" and not os.path.isfile(conf.openstack.openrc):
            raise ValueError(Format.alert("Your openrc file path \"{}\" does not exist.\n\t    "
                                 "Check your openrc file path in a configured yaml file".format(conf.openstack.openrc)))
        return conf
#if __name__ == '__main__':
#    _conf = BaseConfig()
