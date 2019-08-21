from skuba import Skuba
import pytest
import time


@pytest.mark.disruptive
def test_add_worker(bootstrap, skuba):
   skuba.node_join(role="worker", nr=0)
   masters = skuba.num_of_nodes("master")
   workers = skuba.num_of_nodes("worker")
   assert masters == 1
   assert workers == 1
