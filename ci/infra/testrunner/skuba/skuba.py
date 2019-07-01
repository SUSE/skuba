import os

from timeout_decorator import timeout

from platforms.platform import Platform
from utils.format import Format
from utils.utils import (step, Utils)


class Skuba:

    def __init__(self, conf):
        self.conf = conf
        self.binpath = self.conf.skuba.binpath
        self.utils = Utils(self.conf)
        self.platform = Platform.get_platform(conf)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)

    def _verify_skuba_bin_dependency(self):
        if not os.path.isfile(self.binpath):
            raise FileNotFoundError(Format.alert("skuba not found at {}".format(self.binpath)))

    def _verify_bootstrap_dependency(self):
        if not os.path.exists(os.path.join(self.conf.workspace, "test-cluster")):
            raise Exception(Format.alert("test-cluster not found. Please run bootstrap and try again"))

    @staticmethod
    def build(conf):
        """Buids skuba from source"""
        utils = Utils(conf)
        utils.runshellcommand("rm -fr go")
        utils.runshellcommand("mkdir -p go/src/github.com/SUSE")
        utils.runshellcommand("cp -a {} go/src/github.com/SUSE/".format(conf.skuba.srcpath))
        utils.gorun("go version")
        print("Building skuba")
        utils.gorun("make")

    @staticmethod
    def cleanup(conf):
        """Cleanup skuba working environment"""
        utils = Utils(conf)
        # TODO: check why (and if) the following two commands are needed
        cmd = 'mkdir -p {}/logs'.format(conf.workspace)
        utils.runshellcommand(cmd)

        # This is pretty aggressive but modules are also present
        # in workspace and they lack the 'w' bit so just set
        # everything so we can do whatever we want during cleanup
        cmd = 'chmod -R 777 {}'.format(conf.workspace)
        utils.runshellcommand(cmd)

        # TODO: appending workspace is not necessary as runshellcommand has it as workdirectory
        dirs = [os.path.join(conf.workspace, "test-cluster"),
                os.path.join(conf.workspace, "go"),
                os.path.join(conf.workspace, "logs"),
                #TODO: move this to utils as ssh_cleanup
                os.path.join(conf.workspace, "ssh-agent-sock")]

        cleanup_failure = False
        for dir in dirs:
            try: 
                utils.runshellcommand("rm -rf {}".format(dir))
            except Exception as ex:
                cleanup_failure = True
                print("Received the following error {}".format(ex))
                print("Attempting to finish cleaup")

        if cleanup_failure:
            raise Exception("Failure(s) during cleanup")

    @step
    def cluster_init(self):

        print("Cleaning up any previous test-cluster dir")
        self.utils.runshellcommand("rm -rf {}".format(self.cwd))
        cmd = "cluster init --control-plane {} test-cluster".format(self.platform.get_lb_ipaddr())
        # Override work directory, because init must run in the parent directory of the
        # cluster directory
        self._run_skuba(cmd, cwd=self.conf.workspace)

    @step
    def node_bootstrap(self):
        self._verify_bootstrap_dependency()

        master0_ip = self.platform.get_nodes_ipaddrs("master")[0]
        cmd = "node bootstrap --user {username} --sudo --target \
                 {ip} my-master-0".format(ip=master0_ip, username=self.conf.nodeuser)
        self._run_skuba(cmd)

    @step
    def node_join(self, role="worker", nr=0):
        self._verify_bootstrap_dependency()

        ip_addrs = self.platform.get_nodes_ipaddrs(role)

        if nr < 0:
            raise ValueError("Node number cannot be negative")

        if nr >= len(ip_addrs):
            raise Exception(Format.alert("Node {role}-{nr} no deployed in "
                      "infrastructure".format(role=role, nr=nr)))

        cmd = "node join --role {role} --user {username} --sudo --target {ip} \
               my-{role}-{nr}".format(role=role, ip=ip_addrs[nr], nr=nr, 
                                   username=self.conf.nodeuser)
        try: 
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}") from ex

    @step
    def node_remove(self, role="worker", nr=0):
        self._verify_bootstrap_dependency()

        if role not in ("master", "worker"):
            raise ValueError("Invalid role {}".format(role))

        n_nodes = self.num_of_nodes(role)

        if nr < 0:
            raise ValueError("Node number must be non negative")

        if nr >= n_nodes:
            raise ValueError("Error: there is no {role}-{rn} \
                              node to remove from cluster".format(role=role, nr=nr))

        cmd = "node remove my-{role}-{nr}".format(role=role, nr=nr)

        try: 
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(cmd)) from ex

    @step
    def node_reset(self, role="worker", nr=0):
        self._verify_bootstrap_dependency()

        ip_addrs = self.platform.get_nodes_ipaddrs(role)

        if nr < 0:
            raise ValueError("Node number cannot be negative")

        if nr >= len(ip_addrs):
            raise Exception(Format.alert("Node {role}-{nr} not deployed in "
                      "infrastructure".format(role=role, nr=nr)))

        cmd = "node reset --user {username} --sudo --target {ip}".format(
                ip=ip_addrs[nr], username=self.conf.nodeuser)
        try: 
            self._run_skuba(cmd)
        except Exception as ex:
            raise Exception("Error executing cmd {}".format(cmd)) from ex

    def cluster_status(self):
        self._verify_bootstrap_dependency()
        self._run_skuba("cluster status")

    def num_of_nodes(self, role):

        if role not in ("master", "worker"):
            raise ValueException("Invalid role '{}'".format(role))

        test_cluster = os.path.join(self.conf.workspace, "test-cluster")
        cmd = "cd " + test_cluster + "; " + self.binpath + " cluster status"
        output = self.utils.runshellcommand_withoutput(cmd)
        return output.count(role)

    @timeout(600)
    @step
    def gather_logs(self):
        logging_errors = []
        log_dir_path = os.path.join(self.conf.workspace, 'testrunner_logs')

        if not os.path.isdir(log_dir_path):
            os.mkdir(log_dir_path)
            print(f"Created log dir {log_dir_path}")

        for ipa in self.platform.get_nodes_ipaddrs("master"):
            logging_error = self._copy_cloud_init_logs(ipa, 'master', log_dir_path)
            if logging_error is not None:
                logging_errors.append(logging_error)

        for ipa in self.platform.get_nodes_ipaddrs("worker"):
            logging_error = self._copy_cloud_init_logs(ipa, 'worker', log_dir_path)
            if logging_error is not None:
                logging_errors.append(logging_error)

        if logging_errors:
            raise Exception(f"{len(logging_errors)} Failure(s) while collecting logs")

    def _copy_cloud_init_logs(self, ip_address, node_type, log_dir_path):
        node_log_dir_path = os.path.join(log_dir_path, f"{node_type}_{ip_address.replace('.', '_')}")

        if not os.path.isdir(node_log_dir_path):
            os.mkdir(node_log_dir_path)
            print(f"Created log dir {node_log_dir_path}")
        try:
            print(f"Gathering cloud-init logs from {node_type} at {ip_address}")
            self.utils.scp_file(ip_address, "/var/run/cloud-init/status.json", node_log_dir_path)
            self.utils.scp_file(ip_address, "/var/log/cloud-init-output.log", node_log_dir_path)
            self.utils.scp_file(ip_address, "/var/log/cloud-init.log", node_log_dir_path)
        except Exception as ex:
            print(f"Error while collecting cloud-init logs from {node_type} at {ip_address}\n {ex}")
            return ex

    def _run_skuba(self, cmd, cwd=None):
        """Running skuba command in cwd.
        The cwd defautls to {workspace}/test-cluster but can be overrided
        for example, for the init command that must run in {workspace}
        """
        self._verify_skuba_bin_dependency()

        if cwd is None:
           cwd=self.cwd

        env = {"SSH_AUTH_SOCK": os.path.join(self.conf.workspace, "ssh-agent-sock")}

        self.utils.runshellcommand(self.binpath + " "+ cmd, cwd=cwd, env=env)

