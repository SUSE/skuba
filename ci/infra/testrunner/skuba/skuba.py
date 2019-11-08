import logging
import os
import stat
import time

import platforms
from utils.format import Format
from utils.utils import (step, Utils)

logger = logging.getLogger('testrunner')


class Skuba:

    def __init__(self, conf, platform):
        self.conf = conf
        self.binpath = self.conf.skuba.binpath
        self.utils = Utils(self.conf)
        self.platform = platforms.get_platform(conf, platform)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)
        self.utils.setup_ssh()

    def _verify_skuba_bin_dependency(self):
        if not os.path.isfile(self.binpath):
            raise FileNotFoundError("skuba not found at {}".format(self.binpath))

    def _verify_bootstrap_dependency(self):
        if not os.path.exists(os.path.join(self.conf.workspace, "test-cluster")):
            raise ValueError("test-cluster not found. Please run bootstrap and try again")

    @staticmethod
    @step
    def cleanup(conf):
        """Cleanup skuba working environment"""

        # TODO: check why (and if) the following command is needed
        # This is pretty aggressive but modules are also present
        # in workspace and they lack the 'w' bit so just set
        # everything so we can do whatever we want during cleanup
        Utils.chmod_recursive(conf.workspace, 0o777)

        dirs = [
            os.path.join(conf.workspace, "test-cluster"),
            os.path.join(conf.workspace, "go"),
            os.path.join(conf.workspace, "logs"),
        ]

        Utils.cleanup_files(dirs)

    @step
    def cluster_init(self, kubernetes_version=None, cloud_provider=None):
        logger.debug("Cleaning up any previous test-cluster dir")
        self.utils.cleanup_file(self.cwd)

        k8s_version_option, cloud_provider_option = "", ""
        if kubernetes_version:
            k8s_version_option = "--kubernetes-version {}".format(kubernetes_version)
        if cloud_provider:
            cloud_provider_option = "--cloud-provider {}".format(type(self.platform).__name__.lower())

        cmd = "cluster init --control-plane {} {} {} test-cluster".format(self.platform.get_lb_ipaddr(),
                                                                          k8s_version_option, cloud_provider_option)
        # Override work directory, because init must run in the parent directory of the
        # cluster directory
        self._run_skuba(cmd, cwd=self.conf.workspace)

    @step
    def node_bootstrap(self, cloud_provider=None):
        self._verify_bootstrap_dependency()

        if cloud_provider:
            self.platform.setup_cloud_provider()

        master0_ip = self.platform.get_nodes_ipaddrs("master")[0]
        master0_name = self.platform.get_nodes_names("master")[0]
        cmd = (f'node bootstrap --user {self.conf.nodeuser} --sudo --target '
               f'{master0_ip} {master0_name}')
        self._run_skuba(cmd)


    @step
    def node_join(self, role="worker", nr=0):
        self._verify_bootstrap_dependency()

        ip_addrs = self.platform.get_nodes_ipaddrs(role)
        node_names = self.platform.get_nodes_names(role)

        if nr < 0:
            raise ValueError("Node number cannot be negative")

        if nr >= len(ip_addrs):
            raise Exception(Format.alert("Node {role}-{nr} is not deployed in "
                                         "infrastructure".format(role=role, nr=nr)))

        cmd = (f'node join --role {role} --user {self.conf.nodeuser} '
               f' --sudo --target {ip_addrs[nr]} {node_names[nr]}')
        try:
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex

    def join_nodes(self, masters=None, workers=None):
        if masters is None:
            masters = self.platform.get_num_nodes("master")
        if workers is None:
            workers = self.platform.get_num_nodes("worker")

        nodes = [("master", n) for n in range(1, masters)] + \
                [("worker", n) for n in range(0, workers)] 
        for role, node in nodes:
            self.node_join(role, node)
            self._wait_node_joined(role, node, timeout=180, backoff=20)


    def _wait_node_joined(self, role, node, timeout=60, backoff=10):
        
        node_name = self.platform.get_nodes_names(role)[node]
        deadline = int(time.time()) + timeout
        while True:
            last_error = None
            try:
                status = self.cluster_status()
                if status.find(node_name) > -1:
                    return
            except Exception as ex:
                last_error = ex

            if int(time.time()) >= deadline:
                raise Exception((f'Node {node_name} not shown ready after {timeout} seconds'
                                 f'{". Last error:"+str(last_error) if last_error else ""}'))
            time.sleep(backoff)
        
        
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
    def node_upgrade(self, action, role, nr, ignore_errors=False):
        self._verify_bootstrap_dependency()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role '{}'".format(role))

        if nr >= self.num_of_nodes(role):
            raise ValueError("Error: there is no {}-{} \
                              node in the cluster".format(role, nr))

        if action == "plan":
            node_names = self.platform.get_nodes_names(role)
            cmd = f'node upgrade plan {node_names[nr]}'
        elif action == "apply":
            ip_addrs = self.platform.get_nodes_ipaddrs(role)
            cmd = (f'node upgrade apply --user {self.conf.nodeuser} --sudo'
                   f' --target {ip_addrs[nr]}')
        else:
            raise ValueError("Invalid action '{}'".format(action))

        return self._run_skuba(cmd, ignore_errors=ignore_errors)

    @step
    def cluster_upgrade_plan(self):
        self._verify_bootstrap_dependency()
        return self._run_skuba("cluster upgrade plan")

    @step
    def cluster_status(self):
        self._verify_bootstrap_dependency()
        return self._run_skuba("cluster status")

    @step
    def addon_upgrade(self, action):
        self._verify_bootstrap_dependency()
        if action not in ['plan', 'apply']:
            raise ValureEror("Invalid action '{}'".format(action))
        return self._run_skuba("addon upgrade {0}".format(action))

    @step
    def num_of_nodes(self, role):

        if role not in ("master", "worker"):
            raise ValueError("Invalid role '{}'".format(role))

        test_cluster = os.path.join(self.conf.workspace, "test-cluster")
        cmd = "cd " + test_cluster + "; " + self.binpath + " cluster status"
        output = self.utils.runshellcommand(cmd)
        return output.count(role)

    @step
    def get_kubeconfig(self):
        path = "{cwd}/test-cluster/admin.conf".format(cwd=self.conf.workspace)
        return path

    def _run_skuba(self, cmd, cwd=None, verbosity=None, ignore_errors=False):
        """Running skuba command in cwd.
        The cwd defautls to {workspace}/test-cluster but can be overrided
        for example, for the init command that must run in {workspace}
        """
        self._verify_skuba_bin_dependency()

        if cwd is None:
            cwd = self.cwd

        if verbosity is None:
            verbosity = self.conf.skuba.verbosity
        try:
            v = int(verbosity)
        except ValueError:
            raise ValueError(f"verbosity '{verbosity}' is not an int")
        verbosity = v

        env = {
            "SSH_AUTH_SOCK": self.utils.ssh_sock_fn(),
            "PATH": os.environ['PATH']
        }

        return self.utils.runshellcommand(
            f"{self.binpath} -v {verbosity} {cmd}",
            cwd=cwd,
            env=env,
            ignore_errors=ignore_errors,
        )
