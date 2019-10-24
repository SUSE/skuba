import logging
import pathlib
import textwrap

import pytest

from tests.utils import wait

logger = logging.getLogger("testrunner")

CONTENT = """---
apiVersion: v1
kind: Pod
metadata:
    name: leap
spec:
    containers:
    - name: app
      image: opensuse/leap:latest
      command: ['/bin/sh', '-c', 'sleep 3600']
---
apiVersion: v1
kind: Pod
metadata:
    name: sle12sp4
spec:
    containers:
    - name: app
      image: registry.suse.com/suse/sles12sp4:latest
      command: ['/bin/sh', '-c', 'sleep 3600']
---
apiVersion: v1
kind: Pod
metadata:
    name: sle15
spec:
    containers:
    - name: app
      image: registry.suse.com/suse/sle15:latest
      command: ['/bin/sh', '-c', 'sleep 3600']
---
apiVersion: v1
kind: Pod
metadata:
    name: sle15sp1
spec:
    containers:
    - name: app
      image: registry.suse.de/suse/containers/sle-server/15/containers/suse/sle15:15.1
      command: ['/bin/sh', '-c', 'sleep 3600']"""


@pytest.fixture
def setup_manifest():
    cwd = pathlib.Path.cwd()
    p = cwd / "pods.yaml"
    p.touch()
    p.write_text(textwrap.dedent(CONTENT))
    yield p
    p.unlink()


@pytest.mark.flaky
def test_dockercaps(deployment, kubectl, setup_manifest):
    logger.info("Deploy testcases")
    kubectl.run_kubectl(
        "apply -f {}".format(setup_manifest))

    wait(kubectl.run_kubectl,
         "wait --for=condition=ready pods --all --timeout=0",
         wait_delay=30,
         wait_timeout=10,
         wait_backoff=30,
         wait_elapsed=180,
         wait_allow=(RuntimeError))

    logger.info("Test: Run 'su root -c id' on the containers")
    pods = ["sle12sp4", "leap", "sle15", "sle15sp1"]
    for container in pods:
        output = kubectl.run_kubectl(
            "exec -it {} -- su root -c id".format(container))
        assert 'uid=0' in output

    logger.info("Test: Add a new user to the containers")
    for container in pods:
        output = kubectl.run_kubectl(
            "exec -it {} -- useradd panos".format(container))
        assert 'PAM' not in output

    # Remove the testing pods
    kubectl.run_kubectl(
        "delete -f {}".format(setup_manifest))
