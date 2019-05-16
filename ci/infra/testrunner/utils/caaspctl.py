import os, json
from timeout_decorator import timeout
from utils import step
from utils import Utils
from utils import Constant

class Caaspctl:

    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self._num_master, self._num_worker = 0, 0

    def verify_tf_dependency(self):
        if not os.path.exists(os.path.join(os.path.join(self.conf.workspace, "tfout.json"))):
            raise RuntimeError("{}You need to run \"testrunner --terraform first"
                               " before running any caaspctl commands\"{}".format(Constant.RED, Constant.COLOR_EXIT))

    def verify_caaspctl_bin_dependency(self):
        caaspctl_bin_path = os.path.join(self.conf.workspace, 'go/bin/caaspctl')
        if not os.path.isfile(caaspctl_bin_path):
            raise FileNotFoundError("{} You need to run 'testrunner --create-caaspctl' first ".format(
                                                                                  Constant.RED, Constant.COLOR_EXIT))
    def verify_boostrap_dependency(self):
        if not os.path.exists(os.path.join(self.conf.workspace, "test-cluster")):
            raise RuntimeError("{}Dir test-cluster does not exists. You need to run \"testrunner --bootstrap\""
                               " first {} ".format(Constant.RED, Constant.COLOR_EXIT))

    @timeout(600)
    @step
    def create_caaspctl(self):
        """Configure Environment"""
        self.utils.runshellcommand("rm -fr go")
        self.utils.runshellcommand("mkdir -p go/src/github.com/SUSE")
        self.utils.runshellcommand("cp -a caaspctl go/src/github.com/SUSE/")
        self.utils.gorun("go version")
        print("{}Building caaspctl from {}{}".format(Constant.BLUE,
                            os.path.join(self.conf.workspace, "go/src/github.com/SUSE/"), Constant.COLOR_EXIT))
        self.utils.gorun("make")

    @timeout(600)
    @step
    def bootstrap_environment(self):
        """Bootstrap Environment"""
        self.verify_caaspctl_bin_dependency()
        self.verify_tf_dependency()
        self.utils.setup_ssh()
        self._caaspctl_cluster_init()
        self._caaspctl_node_bootstrap()
        self.caaspctl_cluster_status()


    @step
    def _caaspctl_cluster_init(self):
        print("Cleaning up any previous test-cluster dir")
        self.utils.runshellcommand("rm {}/test-cluster -rf".format(self.conf.workspace))
        cmd = "cluster init --control-plane {} test-cluster".format(self._get_lb_ipaddr())
        self.utils.run_caaspctl(cmd, init=True)

    @step
    def _caaspctl_node_bootstrap(self):
        cmd = "node bootstrap --user {username} --sudo --target \
                 {ip} my-master-0".format(ip=self._get_masters_ipaddrs()[0], username=self.conf.nodeuser)
        self.utils.run_caaspctl(cmd)
        self._num_master += 1

    def _caaspctl_node_join(self, role="worker", nr=0):
        if role == "master":
            ip_addr = self._get_masters_ipaddrs()[nr]
        else:
            ip_addr = self._get_workers_ipaddrs()[nr]

        cmd = "node join --role {role} --user {username} --sudo --target {ip} my-{role}-{nr}".format(
            role=role, ip=ip_addr, nr=nr, username=self.conf.nodeuser)
        try:
            self.utils.run_caaspctl(cmd)
        except:
            raise RuntimeError("{}Error: {}{}".format(Constant.RED, cmd, Constant.COLOR_EXIT))

    def _caaspctl_node_remove(self, role="worker", nr=0):
        if role == "master":
            ip_addr = self._get_masters_ipaddrs()[nr]
        else:
            ip_addr = self._get_workers_ipaddrs()[nr]

        cmd = "node remove my-{role}-{nr}".format(role=role, ip=ip_addr, nr=nr, username=self.conf.nodeuser)
        try:
            self.utils.run_caaspctl(cmd)
        except:
            raise RuntimeError("{}Error: {}{}".format(Constant.RED, cmd, Constant.COLOR_EXIT))

    def _add_worker_in_cluster(self):
        avialable_workers = len(self._get_workers_ipaddrs())
        if self._num_worker >= avialable_workers:
            raise ValueError("{}Error: there is no available worker node left in cluster{}".format(
                Constant.RED, Constant.COLOR_EXIT))
        try:
            self._caaspctl_node_join(role="worker", nr=self._num_worker)
            self._num_worker += 1
        except:
            self._num_worker -= 1
            raise

    def _add_master_in_cluster(self):
        avialable_masters = len(self._get_masters_ipaddrs())
        if self._num_master >= avialable_masters:
            raise ValueError("{}Error: there is no available master node left in cluster{}".format(
                Constant.RED, Constant.COLOR_EXIT))
        try:
            self._caaspctl_node_join(role="master", nr=self._num_master)
            self._num_master += 1
        except:
            self._num_master -= 1
            raise

    def _remove_worker_in_cluster(self):
        self._num_worker -= 1
        if self._num_worker < 0:
            raise ValueError("{}Error: there is not enough worker node to remove in cluster{}".format(
                Constant.RED, Constant.COLOR_EXIT))
        try:
            self._caaspctl_node_remove(role="worker", nr=self._num_worker)
        except:
            self._num_worker += 1
            raise

    def _remove_master_in_cluster(self):
        self._num_master -= 1
        if self._num_master <= 0:
            raise ValueError("{}Error: there is only one master node left in cluster{}".format(
                Constant.RED, Constant.COLOR_EXIT))
        try:
            self._caaspctl_node_remove(role="master", nr=self._num_master)
        except:
            self._num_master += 1
            raise

    @timeout(600)
    @step
    def add_nodes_in_cluster(self, num_master=0, num_worker=0):
        self.verify_caaspctl_bin_dependency()
        self.verify_boostrap_dependency()
        self._load_num_of_nodes()

        for _ in range(num_worker):
            self._add_worker_in_cluster()
        for _ in range(num_master):
            self._add_master_in_cluster()

    @timeout(600)
    @step
    def remove_nodes_in_cluster(self, num_master=0, num_worker=0):
        self.verify_caaspctl_bin_dependency()
        self.verify_boostrap_dependency()
        self._load_num_of_nodes()

        for _ in range(num_worker):
            self._remove_worker_in_cluster()
        for _ in range(num_master):
            self._remove_master_in_cluster()


    def caaspctl_cluster_status(self):
        self.verify_caaspctl_bin_dependency()
        self.verify_boostrap_dependency()
        self.utils.run_caaspctl("cluster status")

    def _load_num_of_nodes(self):
        try:
            test_cluster = os.path.join(self.conf.workspace, "test-cluster")
            binpath = os.path.join(self.conf.workspace, 'go/bin/caaspctl')
            cmd = "cd " + test_cluster + "; " +binpath + " cluster status"
            output = self.utils.runshellcommand_withoutput(cmd)
        except:
             self._num_master, self._num_worker = 0, 0
        self._num_master, self._num_worker = output.count("master"), output.count("worker")

    def _load_tfstate(self):
        fn = os.path.join(self.conf.terraform_dir, "terraform.tfstate")
        print("Reading {}".format(fn))
        with open(fn) as f:
            return json.load(f)

    def _get_lb_ipaddr(self):
        state = self._load_tfstate()
        return state["modules"][0]["outputs"]["ip_ext_load_balancer"]["value"]

    def _get_masters_ipaddrs(self):
        state = self._load_tfstate()
        return state["modules"][0]["outputs"]["ip_masters"]["value"]

    def _get_workers_ipaddrs(self):
        state = self._load_tfstate()
        return state["modules"][0]["outputs"]["ip_workers"]["value"]

    @timeout(600)
    @step
    def gather_logs(self):
        self.verify_tf_dependency()
        ipaddrs = self._get_masters_ipaddrs() + self._get_workers_ipaddrs()
        for ipa in ipaddrs:
            print("--------------------------------------------------------------")
            print("Gathering logs from {}".format(ipa))
            self.utils.ssh_run(ipa, "cat /var/run/cloud-init/status.json")
            print("--------------------------------------------------------------")
            self.utils.ssh_run(ipa, "cat /var/log/cloud-init-output.log")
