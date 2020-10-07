import logging
import os
import toml

import platforms
from checks import Checker
from utils.utils import (step, Utils)

logger = logging.getLogger('testrunner')


class Skuba:

    def __init__(self, conf, platform):
        self.conf = conf
        self.binpath = self.conf.skuba.binpath
        self.utils = Utils(self.conf)
        self.platform = platforms.get_platform(conf, platform)
        self.workdir = self.conf.skuba.workdir
        self.cluster = self.conf.skuba.cluster
        self.cluster_dir = os.path.join(self.workdir, self.cluster)
        self.utils.setup_ssh()
        self.checker = Checker(conf, platform)

    def _verify_skuba_bin_dependency(self):
        if not os.path.isfile(self.binpath):
            raise FileNotFoundError("skuba not found at {}".format(self.binpath))

    def _verify_bootstrap_dependency(self):
        if not os.path.exists(self.cluster_dir):
            raise ValueError(f'cluster {self.cluster} not found at {self.workdir}.'
                             ' Please run bootstrap and try again')

    @staticmethod
    @step
    def cleanup(conf):
        """Cleanup skuba working environment"""
        cluster_dir = os.path.join(conf.skuba.workdir, conf.skuba.cluster)
        Utils.cleanup_file(cluster_dir)

    @step
    def cluster_deploy(self, kubernetes_version=None, cloud_provider=None,
                       timeout=None, registry_mirror=None):
        """Deploy a cluster joining all nodes"""
        self.cluster_bootstrap(kubernetes_version=kubernetes_version,
                               cloud_provider=cloud_provider,
                               timeout=timeout,
                               registry_mirror=registry_mirror)
        self.join_nodes(timeout=timeout)

    @step
    def cluster_init(self, kubernetes_version=None, cloud_provider=None):
        logger.debug("Cleaning up any previous cluster dir")
        self.utils.cleanup_file(self.cluster_dir)

        k8s_version_option, cloud_provider_option = "", ""
        if kubernetes_version:
            k8s_version_option = "--kubernetes-version {}".format(kubernetes_version)
        if cloud_provider:
            cloud_provider_option = "--cloud-provider {}".format(type(self.platform).__name__.lower())

        cmd = "cluster init --control-plane {} {} {} test-cluster".format(self.platform.get_lb_ipaddr(),
                                                                          k8s_version_option, cloud_provider_option)
        # Override work directory, because init must run in the parent directory of the
        # cluster directory
        self._run_skuba(cmd, cwd=self.workdir)

    @step
    def cluster_bootstrap(self, kubernetes_version=None, cloud_provider=None,
                          timeout=None, registry_mirror=None):
        self.cluster_init(kubernetes_version, cloud_provider)
        self.node_bootstrap(cloud_provider, timeout=timeout, registry_mirror=registry_mirror)

    @step
    def node_bootstrap(self, cloud_provider=None, timeout=None, registry_mirror=None):
        self._verify_bootstrap_dependency()

        if cloud_provider:
            self.platform.setup_cloud_provider(self.workdir)

        if registry_mirror:
            self._setup_container_registries(registry_mirror)

        master0_ip = self.platform.get_nodes_ipaddrs("master")[0]
        master0_name = self.platform.get_nodes_names("master")[0]
        cmd = (f'node bootstrap --user {self.utils.ssh_user()} --sudo --target '
               f'{master0_ip} {master0_name}')
        self._run_skuba(cmd)

        self.checker.check_node("master", 0, stage="joined", timeout=timeout)

    def _setup_container_registries(self, registry_mirror):
        mirrors = {}
        for l in registry_mirror:
            if l[0] not in mirrors:
                mirrors[l[0]] = []
            mirrors[l[0]].append(l[1])

        conf = {'unqualified-search-registries': ['docker.io'], 'registry': []}
        for location, mirror_list in mirrors.items():
            mirror_toml = []
            for m in mirror_list:
                mirror_toml.append({'location': m, 'insecure': True})
            conf['registry'].append(
                    {'prefix': location, 'location': location,
                        'mirror': mirror_toml})
        conf_string = toml.dumps(conf)
        c_dir = os.path.join(self.cluster_dir, 'addons/containers')
        os.mkdir(c_dir)
        with open(os.path.join(c_dir, 'registries.conf'), 'w') as f:
            f.write(conf_string)
    @step
    def node_join(self, role="worker", nr=0, timeout=None):
        self._verify_bootstrap_dependency()

        ip_addrs = self.platform.get_nodes_ipaddrs(role)
        node_names = self.platform.get_nodes_names(role)

        if nr < 0:
            raise ValueError("Node number cannot be negative")

        if nr >= len(ip_addrs):
            raise Exception(f'Node {role}-{nr} is not deployed in '
                            'infrastructure')

        cmd = (f'node join --role {role} --user {self.utils.ssh_user()} '
               f' --sudo --target {ip_addrs[nr]} {node_names[nr]}')
        try:
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception(f'Error joining node {role}-{nr}') from ex

        self.checker.check_node(role, nr, stage="joined", timeout=timeout)

    def join_nodes(self, masters=None, workers=None, timeout=None):
        if masters is None:
            masters = self.platform.get_num_nodes("master")
        if workers is None:
            workers = self.platform.get_num_nodes("worker")

        for node in range(1, masters):
            self.node_join("master", node, timeout=timeout)

        for node in range(0, workers):
            self.node_join("worker", node, timeout=timeout)

    @step
    def node_remove(self, role="worker", nr=0):
        self._verify_bootstrap_dependency()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role {}".format(role))

        n_nodes = self.num_of_nodes(role)

        if nr < 0:
            raise ValueError("Node number must be non negative")

        if nr >= n_nodes:
            raise ValueError("Error: there is no {role}-{nr} \
                              node to remove from cluster".format(role=role, nr=nr))

        node_names = self.platform.get_nodes_names(role)
        cmd = f'node remove {node_names[nr]}'

        try:
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(cmd)) from ex


    @step
    def node_upgrade(self, action, role, nr, ignore_errors=False, timeout=300):
        self._verify_bootstrap_dependency()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role '{}'".format(role))

        if nr >= self.num_of_nodes(role):
            raise ValueError("Error: there is no {}-{} \
                              node in the cluster".format(role, nr))

        node_names = self.platform.get_nodes_names(role)
        if action == "plan":
            cmd = f'node upgrade plan {node_names[nr]}'
        elif action == "apply":
            ip_addrs = self.platform.get_nodes_ipaddrs(role)
            cmd = (f'node upgrade apply --user {self.utils.ssh_user()} --sudo'
                   f' --target {ip_addrs[nr]}')
        else:
            raise ValueError("Invalid action '{}'".format(action))

        result=None
        try:
            result = self._run_skuba(cmd, ignore_errors=ignore_errors)
            self.checker.check_node(role, nr, stage="joined", timeout=timeout)
            return result
        except Exception as ex:
            raise Exception(f'Error upgrading node {node_names[nr]}') from ex


    @step
    def cluster_upgrade(self, action):
        self._verify_bootstrap_dependency()

        if action not in ["plan", "localconfig"]:
            raise ValueError(f'Invalid cluster upgrade action: {action}')

        return self._run_skuba(f'cluster upgrade {action}')

    @step
    def cluster_status(self):
        self._verify_bootstrap_dependency()
        return self._run_skuba("cluster status")

    @step
    def addon_refresh(self, action):
        self._verify_bootstrap_dependency()
        if action not in ['localconfig']:
            raise ValueError("Invalid action '{}'".format(action))
        return self._run_skuba("addon refresh {0}".format(action))

    @step
    def addon_upgrade(self, action):
        self._verify_bootstrap_dependency()
        if action not in ['plan', 'apply']:
            raise ValueError("Invalid action '{}'".format(action))
        return self._run_skuba("addon upgrade {0}".format(action))

    @step
    def num_of_nodes(self, role):

        if role not in ("master", "worker"):
            raise ValueError("Invalid role '{}'".format(role))

        cmd = "cluster status"
        output = self._run_skuba(cmd)
        return output.count(role)

    def _run_skuba(self, cmd, cwd=None, verbosity=None, ignore_errors=False):
        """Running skuba command.
        The cwd defautls to cluster_dir but can be overrided
        for example, for the init command that must run in {workspace}
        """
        self._verify_skuba_bin_dependency()

        if cwd is None:
            cwd = self.cluster_dir

        if verbosity is None:
            verbosity = self.conf.skuba.verbosity
        try:
            v = int(verbosity)
        except ValueError:
            raise ValueError(f"verbosity '{verbosity}' is not an int")
        verbosity = v

        return self.utils.runshellcommand(
            f"{self.binpath} -v {verbosity} {cmd}",
            cwd=cwd,
            ignore_errors=ignore_errors,
        )
