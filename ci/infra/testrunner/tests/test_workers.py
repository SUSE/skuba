from skuba import Skuba
import pytest
import time

def test_add_worker(setup, skuba):
   skuba.node_join(role="worker", nr=0)
   masters = skuba.num_of_nodes("master")
   workers = skuba.num_of_nodes("worker")
   assert masters == 1
   assert workers == 1
