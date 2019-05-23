"""
    Caaspctl will run caaspctl's functions such as bootstrapping,
                         adding and removing node(s), cluster status.
"""
import os
import json
import subprocess
from timeout_decorator import timeout
from utils import step
from utils import Utils
from utils import Constant

class Caaspctl:
    """ Caaspctl class"""
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.state = self._load_tfstate()
        self._num_master, self._num_worker = self._load_num_of_nodes()

    @timeout(600)
    @step
    def create_caaspctl(self):
        """Configure Environment"""

        cmd = "rm -fr go"
        self.utils.runshellcommand(cmd)
        cmd = "mkdir -p go/src/github.com/SUSE"
        self.utils.runshellcommand(cmd)
        cmd = "cp -a caaspctl go/src/github.com/SUSE/"
        self.utils.runshellcommand(cmd)

        self.utils.gorun("go version")
        print("Building caaspctl")
        self.utils.gorun("make")

    @timeout(600)
    @step
    def bootstrap_environment(self):
        """Bootstrap Environment"""
        self.utils.setup_ssh()
        self.caaspctl_cluster_init()
        self.caaspctl_node_bootstrap()
        self.add_worker_in_cluster()
        self.caaspctl_cluster_status()


    @step
    def cleanup(self):
        """Cleanup caaspctl working environment"""
        # TODO: check why (and if) the following two commands are needed
        cmd = 'mkdir -p {}/logs'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        # This is pretty aggressive but modules are also present
        # in workspace and they lack the 'w' bit so just set
        # everything so we can do whatever we want during cleanup
        cmd = 'chmod -R 777 {}'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        _dirs = [os.path.join(self.conf.workspace, "test-cluster"),
                 os.path.join(self.conf.workspace, "go"),
                 os.path.join(self.conf.workspace, "logs"),
                 os.path.join(self.conf.workspace, "ssh-agent-sock"),
                 os.path.join(self.conf.workspace, "test-cluster")]

        self.utils.cleanup(_dirs)

    @step
    def caaspctl_cluster_init(self):
        print("Cleaning up any previous test-cluster dir")
        cmd = "rm {}/test-cluster -rf".format(self.conf.workspace)
        self.utils.runshellcommand(cmd)
        cmd = "cluster init --control-plane {} test-cluster".format(
            self._get_lb_ipaddr())
        self.utils.run_caaspctl(cmd, init=True)

    @step
    def caaspctl_node_bootstrap(self):
        cmd = "node bootstrap --user {username} --sudo --target"\
              "  {ip} my-master-0".format(
                  ip=self._get_masters_ipaddrs()[0],
                  username=self.conf.nodeuser)
        self.utils.run_caaspctl(cmd)
        self._num_master += 1

    @step
    def _caaspctl_node_join(self, role="worker", num=0):
        try:
            if role == "master":
                ip_addr = self._get_masters_ipaddrs()[num]
            else:
                ip_addr = self._get_workers_ipaddrs()[num]
        except:
            raise("{}Error: there is not enough node to add {} node in"
                  " cluster{}".format(
                      Constant.RED, role, Constant.RED_EXIT))

        cmd = "node join --role {role} --user {username} --sudo"\
              " --target {ip} my-{role}-{num}".format(
                  role=role, ip=ip_addr, num=num,
                  username=self.conf.nodeuser)
        try:
            self.utils.run_caaspctl(cmd)
        except:
            raise Exception("{}Error: {}{}".format(Constant.RED,
                                                   cmd,
                                                   Constant.RED_EXIT))

    @step
    def _caaspctl_node_remove(self, role="worker", num=0):
        if num <= 0:
            raise Exception("{}Error: there is not enough node to"
                            " remove {} node in cluster{}"\
                .format(Constant.RED, role, Constant.RED_EXIT))

        cmd = "node remove my-{role}-{num}".format(role=role, num=num)
        try:
            self.utils.run_caaspctl(cmd)
        except:
            raise Exception("{}Error: {}{}".format(Constant.RED, cmd,
                                                   Constant.RED_EXIT))

    @timeout(600)
    @step
    def add_worker_in_cluster(self):
        self._caaspctl_node_join(role="worker", num=self._num_worker)
        self._num_worker += 1


    @timeout(600)
    @step
    def add_master_in_cluster(self):

        self._caaspctl_node_join(role="master", num=self._num_master)
        self._num_master += 1

    @timeout(600)
    @step
    def remove_worker_in_cluster(self):

        self._num_worker -= 1
        self._caaspctl_node_remove(role="worker", num=self._num_worker)



    @timeout(600)
    @step
    def remove_master_in_cluster(self):
        self._num_master -= 1
        self._caaspctl_node_remove(role="master", num=self._num_master)


    @step
    def add_nodes_in_cluster(self, num_master=1, num_worker=1):
        cluster = Caaspctl(self.conf)

        for _ in range(num_worker):
            cluster.add_worker_in_cluster()
        for _ in range(num_master):
            cluster.add_master_in_cluster()

    @step
    def remove_nodes_in_cluster(self, num_master=0, num_worker=1):
        cluster = Caaspctl(self.conf)

        for _ in range(num_worker):
            cluster.remove_worker_in_cluster()
        for _ in range(num_master):
            cluster.remove_master_in_cluster()

    def caaspctl_cluster_status(self):
        self.utils.run_caaspctl("cluster status")

    def _load_num_of_nodes(self):
        try:
            test_cluster = os.path.join(self.conf.workspace,
                                        "test-cluster")
            binpath = os.path.join(self.conf.workspace,
                                   'go/bin/caaspctl')
            cmd = "cd " + test_cluster + "; " + binpath\
                  + " cluster status"
            output = Utils.runshellcommand_withoutput(cmd)
        except subprocess.CalledProcessError:
            return 0, 0
        return output.count("master"), output.count("worker")

    def _load_tfstate(self):
        terraform_tfstate_path = os.path.join(self.conf.terraform_dir,
                                              "terraform.tfstate")
        print("Reading {}".format(terraform_tfstate_path))
        with open(terraform_tfstate_path) as terraform_tfstate:
            return json.load(terraform_tfstate)

    def _get_lb_ipaddr(self):
        return self.state["modules"][0]["outputs"]\
            ["ip_ext_load_balancer"]["value"]

    def _get_masters_ipaddrs(self):
        return self.state["modules"][0]["outputs"]["ip_masters"]\
            ["value"]

    def _get_workers_ipaddrs(self):
        return self.state["modules"][0]["outputs"]["ip_workers"]\
            ["value"]


    @timeout(600)
    @step
    def gather_logs(self):
        logging_error = False

        try:
            ipaddrs = self._get_masters_ipaddrs() + \
                      self._get_workers_ipaddrs()
            for ipa in ipaddrs:
                print("------------------------------------------------"
                      "--------------")
                print("Gathering logs from {}".format(ipa))
                self.utils.ssh_run(ipa, "cat /var/run/cloud-init/"
                                        "status.json")
                print("------------------------------------------------"
                      "--------------")
                self.utils.ssh_run(ipa, "cat /var/log/"
                                        "cloud-init-output.log")
        except subprocess.CalledProcessError as ex:
            logging_error = True
            print("Error while collecting logs from cluster \n {}"\
                  .format(ex))

        if logging_error:
            raise Exception("Failure(s) while collecting logs")
