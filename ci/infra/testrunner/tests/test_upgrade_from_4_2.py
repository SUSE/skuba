import logging
import os
import pytest

from tests.utils import (check_node_is_ready, check_node_version, CURRENT_VERSION, wait)
logger = logging.getLogger('testrunner')


# Migrates a node to the upgrate option speficied in the option
def migrate_node(platform, kubectl, role, node, regcode, option=1):
    platform.ssh_run(role, node, f'sudo SUSEConnect -r {regcode}')
    platform.ssh_run(role, node, "sudo SUSEConnect -p sle-module-containers/15.1/x86_64")
    platform.ssh_run(role, node, f'sudo SUSEConnect -p caasp/4.0/x86_64 -r {regcode}')
    platform.ssh_run(role, node, "sudo sudo zypper in -y --no-recommends zypper-migration-plugin")
    platform.ssh_run(role, node, (f'sudo zypper migration --migration {option}'
                                  ' --no-recommends --non-interactive'
                                  ' --auto-agree-with-licenses --allow-vendor-change'))
    #:FIXME use kured for controlled reboot.
    platform.ssh_run(role, node, "sudo reboot &")

    # wait the node become live
    wait(platform.ssh_run,
        role,
        node,
        "true",
        wait_delay=60,
        wait_timeout=10,
        wait_backoff=30,
        wait_elapsed=180,
        wait_allow=(RuntimeError))

def test_upgrade_from_4_2(deployment, platform, skuba, kubectl):

    skuba.cluster_upgrade(action="localconfig")

    # TODO: find a more elegant way to pick the REG_CODE
    reg_code = os.environ['REG_CODE']
    assert reg_code is not None

    etcds = kubectl.run_kubectl('get pods -nkube-system --no-headers -o custom-columns=":metadata.name" | grep etcd')
    result = platform.ssh_run("master", 0, f'kubectl exec -ti -n kube-system {etcds[0]} -- etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key member list')
    logger.warning(result)

    for role in ("master", "worker"):
        num_nodes = platform.get_num_nodes(role)
        for node in range(0, num_nodes):
            migrate_node(platform, kubectl, role, node, reg_code)
            result = skuba.node_upgrade("apply", role, node)
            assert result.find("successfully upgraded") != -1

            # check node version is update
            wait(check_node_version,
                platform,
                kubectl,
                role,
                node,
                CURRENT_VERSION,
                wait_delay=60,
                wait_backoff=30,
                wait_elapsed=60 * 10,
                wait_allow=(AssertionError))
