import pytest
import os
import platforms
from skuba import Skuba
from utils import BaseConfig
from tests.utils import (check_pods_ready, wait)

CINDER_YAML = '''
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gold
provisioner: kubernetes.io/cinder
parameters:
  availability: nova
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: cinder-claim
  annotations:
    volume.beta.kubernetes.io/storage-class: gold
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: gold
'''

TEST_READ_POD = '''
apiVersion: v1
kind: Pod
metadata:
  name: read-pod
spec:
  containers:
  - name: read-pod
    image: gcr.io/google_containers/busybox:1.24
    command:
      - "/bin/sh"
    args:
      - "-c"
      - "test -f /mnt/SUCCESS && exit 0 || exit 1"
    volumeMounts:
      - name: cinder
        mountPath: "/mnt"
  restartPolicy: "Never"
  volumes:
    - name: cinder
      persistentVolumeClaim:
        claimName: cinder-claim
'''

TEST_WRITE_POD = '''
apiVersion: v1
kind: Pod
metadata:
  name: write-pod
spec:
  containers:
  - name: write-pod
    image: gcr.io/google_containers/busybox:1.24
    command:
      - "/bin/sh"
    args:
      - "-c"
      - "touch /mnt/SUCCESS && exit 0 || exit 1"
    volumeMounts:
      - name: cinder
        mountPath: "/mnt"
  restartPolicy: "Never"
  volumes:
    - name: cinder
      persistentVolumeClaim:
        claimName: cinder-claim
'''


@pytest.fixture(autouse=True, scope='module')
def conf(request):
    """Builds a conf object from a yaml file"""
    path = request.config.getoption("vars")
    return BaseConfig(path)


@pytest.fixture(autouse=True, scope='module')
def target(request):
    """Returns the target platform"""
    platform = request.config.getoption("platform")
    return platform


@pytest.fixture(autouse=True, scope='module')
def platform(conf, target):
    platform = platforms.get_platform(conf, target)
    return platform


@pytest.fixture(autouse=True, scope='module')
def skuba(conf, target):
    return Skuba(conf, target)


@pytest.fixture(autouse=True, scope='module')
def setup(request, platform, skuba, conf):
    def cleanup():
        platform.cleanup()
        skuba.cleanup(conf)

    request.addfinalizer(cleanup)
    platform.provision(num_master=1, num_worker=3)


@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=1)
def test_init_cpi_openstack_cluster(skuba, kubernetes_version=None):
    """
    Initialize the cluster with the openstack cpi
    """
    skuba.cluster_init(kubernetes_version, cloud_provider="openstack")

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=2)
def test_bootstrap_cpi_openstack_cluster(skuba):
    """
    Bootstrap the cluster with the openstack cpi
    """
    skuba.node_bootstrap(cloud_provider="openstack")


@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=3)
def test_node_join_cpi_openstack_cluster(skuba, kubectl):
    """
    Join nodes to the cluster with the openstack cpi enabled
    """
    skuba.join_nodes(masters=1, workers=3)

    wait(check_pods_ready,
         kubectl,
         namespace="kube-system",
         wait_delay=60,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=300,
         wait_allow=(AssertionError))


@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=4)
def test_create_cinder_storage(kubectl, skuba):
        # TODO: result of action is successful  
        kubectl.run_kubectl(" apply -f -", stdin=CINDER_YAML.encode())


@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=5)
def test_wrtite_cinder_storage(kubectl, skuba):
    kubectl.run_kubectl("apply -f -", stdin=TEST_WRITE_POD.encode())

    wait(
        check_pods_ready,
        kubectl,
        pods=["write-pod"],
        statuses=["Succeeded"],
        wait_delay=60,
        wait_retries=1,
        wait_allow=(AssertionError))

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=6)
def test_read_cinder_storage(kubectl, skuba):
    kubectl.run_kubectl(" apply -f -", stdin=TEST_READ_POD.encode())

    wait(
        check_pods_ready,
        kubectl,
        pods=["read-pod"],
        statuses=["Succeeded"],
        wait_delay=60,
        wait_retries=1,
        wait_allow=(AssertionError))
