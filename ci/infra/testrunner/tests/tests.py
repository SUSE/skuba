from caaspctl import Caaspctl
from utils import step
from utils import Utils
from timeout_decorator import timeout

class Tests:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.caaspctl = Caaspctl(conf)
        self._num_master, self._num_worker = self.caaspctl.num_of_nodes()

    @timeout(600)
    @step
    def bootstrap_environment(self):
        """Bootstrap Environment"""
        self.utils.setup_ssh()
        self.caaspctl.cluster_init()
        self.caaspctl.node_bootstrap()
        self._num_master = 1
        self.add_worker_in_cluster()
        try:
            self.caaspctl.cluster_status()
        except:
            pass

    @timeout(600)
    @step
    def add_worker_in_cluster(self):
        try:
            self.caaspctl.node_join(role="worker", nr=self._num_worker)
            self._num_worker += 1
        except:
            self._num_worker -= 1

    @timeout(600)
    @step
    def add_master_in_cluster(self):
        try:
            self.caaspctl.node_join(role="master", nr=self._num_master)
            self._num_master += 1
        except:
            self._num_master -= 1

    @timeout(600)
    @step
    def remove_worker_in_cluster(self):
        try:
            self._num_worker -= 1
            self.caaspctl.node_remove(role="worker", nr=self._num_worker)
        except:
            self._num_worker += 1


    @timeout(600)
    @step
    def remove_master_in_cluster(self):
        try:
            self._num_master -= 1
            self.caaspctl.node_remove(role="master", nr=self._num_master)
        except:
            self._num_master += 1

    @step
    def add_nodes_in_cluster(self, num_master=1, num_worker=1):

        for _ in range(num_worker):
            self.add_worker_in_cluster()
        for _ in range(num_master):
            self.add_master_in_cluster()

        try:
            self.caaspctl.cluster_status()
        except:
            pass

    @step
    def remove_nodes_in_cluster(self, num_master=0, num_worker=1):

        for _ in range(num_worker):
            self.remove_worker_in_cluster()
        for _ in range(num_master):
            self.remove_master_in_cluster()

        try:
            self.caaspctl.cluster_status()
        except:
            pass
