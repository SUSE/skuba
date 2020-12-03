import os
import pytest

from tests.utils import (check_node_is_ready, check_node_version, CURRENT_VERSION, wait)

# Migrates a node to the upgrate option speficied in the option
def migrate_node(platform, checker, kubectl, role, node, regcode, option=1):
    platform.ssh_run(role, node, f'sudo SUSEConnect -r {regcode}')
    platform.ssh_run(role, node, "sudo SUSEConnect -p sle-module-containers/15.1/x86_64")
    platform.ssh_run(role, node, f'sudo SUSEConnect -p caasp/4.0/x86_64 -r {regcode}')
    platform.ssh_run(role, node, "sudo sudo zypper in -y --no-recommends zypper-migration-plugin")
    platform.ssh_run(role, node, (f'sudo zypper migration --migration {option}'
                                  ' --no-recommends --non-interactive'
                                  ' --auto-agree-with-licenses --allow-vendor-change'))

    # update to latest kubernetes package version, if any
    k8s_major, k8s_minor, _ =  CURRENT_VERSION.split('.')
    platform.ssh_run(role, node, 'sudo zypper ar http://download.suse.de/ibs/Devel:/CaaSP:/4.5/SLE_15_SP2/ CaasP_Devel')
    platform.ssh_run(role, node, ('sudo zypper update --allow-vendor-change -y'
                                  f' kubernetes-{k8s_major}.{k8s_minor}-kubeadm'))
    #:FIXME use kured for controlled reboot.
    platform.ssh_run(role, node, "nohup sudo reboot &>/dev/null & exit")

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

    checker.check_node(role, node, stage="joined", timeout=60)

def test_upgrade_from_4_2(deployment, platform, checker, skuba, kubectl):

    skuba.cluster_upgrade(action="localconfig")

    # TODO: find a more elegant way to pick the REG_CODE
    reg_code = os.environ['REG_CODE']
    assert reg_code is not None

    for role in ("master", "worker"):
        num_nodes = platform.get_num_nodes(role)
        for node in range(0, num_nodes):
            migrate_node(platform, checker, kubectl, role, node, reg_code)
            result = skuba.node_upgrade("apply", role, node, timeout=300)
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
