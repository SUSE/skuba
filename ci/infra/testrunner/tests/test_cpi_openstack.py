import json
import pytest
import time
import os
import platforms
from skuba import Skuba
from utils import BaseConfig
from tests.utils import wait

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
    try:
        skuba.cluster_init(kubernetes_version, cloud_provider="openstack")
    except:
        pytest.fail("Failure on initializing cpi optionstack init  ...")

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=2)
def test_bootstrap_cpi_openstack_cluster(skuba):
    """
    Bootstrap the cluster with the openstack cpi
    """
    try:
        skuba.node_bootstrap(cloud_provider="openstack")
    except:
        pytest.fail("Failure on bootstrapping a cluster with cpi optionstack  ...")

def check_system_pods_ready(kubectl):
    pods = json.loads(kubectl.run_kubectl('get pods --namespace=kube-system -o json'))['items']
    for pod in pods:
        pod_status = pod['status']['phase']
        pod_name   = pod['metadata']['name']
        assert pod_status in ['Running', 'Completed'], f'Pod {pod_name} status {pod_status} != Running or Completed'

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=3)
def test_node_join_cpi_openstack_cluster(skuba, kubectl):
    """
    Join nodes to the cluster with the openstack cpi enabled
    """
    try:
        skuba.join_nodes(masters=1, workers=3)
    except:
        pytest.fail("Failure on joinning nodes to the cluster with cpi optionstack  ...")

    wait(check_system_pods_ready,
         kubectl, 
         wait_delay=60,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=300,
         wait_allow=(AssertionError))


@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=4)
def test_create_cinder_storage(kubectl, skuba):
    create_yamlfiles(skuba)
    try:
        file_path = os.path.join(skuba.cwd, "cinder.yaml")
        cmd = f" apply -f {file_path}"
        kubectl.run_kubectl(cmd)
    except:
        pytest.fail(f"Failure on 'kubectl {cmd}' ...")

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=5)
def test_wrtite_cinder_storage(kubectl, skuba):
    try:
        file_path = os.path.join(skuba.cwd, "test_write.yaml")
        cmd = f" apply -f {file_path}"
        kubectl.run_kubectl(cmd)
    except:
        pytest.fail(f"Failure on 'kubectl {cmd}' ...")
    time.sleep(60)
    try:
        cmd = " get pod | grep write-pod "
        out = kubectl.run_kubectl(cmd)
        assert "Completed" in out
    except:
        pytest.fail(f"Failure on 'kubectl {cmd}' ...")

@pytest.mark.disruptive
@pytest.mark.openstack
@pytest.mark.run(order=6)
def test_read_cinder_storage(kubectl, skuba):
    try:
        file_path = os.path.join(skuba.cwd, "test_read.yaml")
        cmd = f" apply -f {file_path}"
        kubectl.run_kubectl(cmd)
    except:
        pytest.fail(f"Failure on 'kubectl {cmd}' ...")
    time.sleep(60)
    try:
        cmd = " get pod | grep read-pod"
        out = kubectl.run_kubectl(cmd)
        assert "Completed" in out
    except:
        pytest.fail(f"Failure on 'kubectl {cmd}' ...")

def create_yamlfiles(skuba):
    filenames = {"cinder.yaml":get_cinder_yaml,
                 "test_write.yaml":get_test_write_pod_yaml,
                 "test_read.yaml":get_test_read_pod_yaml}
    for filename, _function in filenames.items():
        contents = _function()
        with open(os.path.join(skuba.cwd, filename), mode='wt') as yaml_file:
            for line in contents:
                yaml_file.write(line)

def get_cinder_yaml():
    yaml = '''
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
    return yaml

def get_test_read_pod_yaml():
    test_read_pod = '''
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
    return test_read_pod

def get_test_write_pod_yaml():
    test_write_pod = '''
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
    return test_write_pod
