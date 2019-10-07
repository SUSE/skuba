import pytest
from tests.utils import wait

def test_reboot(deployment, platform):

    try:
        platform.ssh_run("worker", 0, "sudo reboot &")
    except Exception:
        pass

    wait(platform.ssh_run, "worker", 0, "/bin/true", wait_delay=30,  wait_timeout=10, wait_backoff=30, wait_retries=5, wait_allow=(RuntimeError))
