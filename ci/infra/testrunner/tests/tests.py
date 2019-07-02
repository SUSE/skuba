from timeout_decorator import timeout

from skuba import Skuba
from utils import ( step, Utils)


class Tests:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.skuba = Skuba(conf)
        # TODO: this is inefficient, but this logic would probably
        # change soon, as it makes no sense to duplicate the node count
        # here.
        self._num_master = self.skuba.num_of_nodes("master")
        self._num_worker = self.skuba.num_of_nodes("worker")

    @timeout(600)
    @step
    def bootstrap_environment(self):
        """Bootstrap Environment"""
        self.skuba.cluster_init()
        self.skuba.node_bootstrap()
        self._num_master = 1
        self.skuba.cluster_status()

    @timeout(600)
    @step
    def add_worker_in_cluster(self):
        self.skuba.node_join(role="worker", nr=self._num_worker)
        self._num_worker += 1

    @timeout(600)
    @step
    def add_master_in_cluster(self):
        self.skuba.node_join(role="master", nr=self._num_master)
        self._num_master += 1

    @timeout(600)
    @step
    def remove_worker_in_cluster(self):
        if self._num_worker == 0:
            raise Exception("No worker to remove from cluster")

        self.skuba.node_remove(role="worker", nr=self._num_worker-1)
        self._num_worker -= 1

    @timeout(600)
    @step
    def remove_master_in_cluster(self):
        if self._num_master == 0:
            raise Exception("No master nodes to remove from cluster")

        self.skuba.node_remove(role="master", nr=self._num_master-1)
        self._num_master -= 1

