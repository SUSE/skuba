#!/usr/bin/env python3 -Wd -b

"""
    Runs end-to-end product tests for v4+.
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

import sys
from argparse import ArgumentParser

from platforms import (Openstack, VMware)
from tests import Tests
from utils import (BaseConfig, Format, Skuba, Utils)

__version__ = "0.0.3"


def main():
    help = """
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.
    """
    parser = ArgumentParser(help)

    parser.add_argument("-z", "--git-rebase", dest="git_rebase", action="store_true",
                        help="git rebase to master")
    parser.add_argument("-i", "--info", dest="ip_info", action='store_true', help='ip info')
    parser.add_argument("-x", "--cleanup", dest="cleanup", action='store_true',
                          help="cleanup created skuba environment")
    parser.add_argument("-t", "--terraform-apply", dest="apply_terraform", action="store_true",
                        help="deploy nodes for cluster in your configured platform \
                              e.g) openstack, vmware, ..")
    parser.add_argument("-c", "--create-skuba", dest="create_skuba", action="store_true",
                        help="create skuba environment {workspace}/go/src/github.com/SUSE/skuba\
                              and build skuba in that directory")
    parser.add_argument("-b", "--bootstrap", dest="boostrap", action="store_true",
                        help="bootstrap k8s cluster with deployed nodes in your platform")
    parser.add_argument("-k", "--status", dest="cluster_status", action="store_true",
                        help="check K8s cluster status")
    parser.add_argument("-a", "--add-nodes", dest="add_nodes", action="store_true",
                        help="add nodes in k8s cluster. Default values are -m=1, -w=1")
    parser.add_argument("-r", "--remove-nodes", dest="remove_nodes", action="store_true",
                        help="remove nodes in k8s cluster. default values are -m=1, -w=1")
    parser.add_argument("-l", "--log", dest="log", action="store_true", help="gather logs from nodes")
    parser.add_argument("-v", "--vars", dest="yaml_path", default="vars/openstack.yaml",
                        help='path for platform yaml file. \
                              Default is vars/openstack.yaml in {workspace}/ci/infra/testrunner. \
                              eg) -v vars/myconfig.yaml')
    parser.add_argument("-m", "--master", dest="num_master", type=int, default=1,
                        help='number of masters to add or delete. It is dependening on \
                              number of deployed master nodes in your yaml file. Default value is 1. \
                              eg) -m 2')
    parser.add_argument("-w", "--worker", dest="num_worker", type=int, default=1,
                        help='number of workers to add or delete. It is dependening on \
                              number of deployed worker nodes in your yaml file. Default value is 1  \
                              eg) -w 2')

    options = parser.parse_args()
    conf = BaseConfig(options.yaml_path)

    if options.ip_info:
        Utils(conf).info()
    if options.git_rebase:
        Utils(conf).git_rebase()
    elif options.cleanup:
        get_platform(conf).cleanup()
        Skuba.cleanup(conf)
    elif options.apply_terraform:
        get_platform(conf).apply_terraform()
    elif options.create_skuba:
        Skuba.build(conf)
    elif options.boostrap:
        Tests(conf).bootstrap_environment()
    elif options.cluster_status:
        Skuba(conf).cluster_status()
    elif options.add_nodes:
        Tests(conf).add_nodes_in_cluster(num_master=options.num_master, num_worker=options.num_worker)
    elif options.remove_nodes:
        Tests(conf).remove_nodes_in_cluster(num_master=options.num_master, num_worker=options.num_worker)
    elif options.log:
        Skuba(conf).gather_logs()

    sys.exit(0)

def get_platform(conf):
    if conf.platform == "openstack":
        platform = Openstack(conf)
    elif conf.platform == "vmware":
        platform = VMware(conf)
    elif conf.platform == "bare-metal":
        # TODO platform = Bare_metal(conf, utils)
        print("Todo: bare-metal is not ready yet")
        sys.exit(0)
    elif conf.platform == "libvirt":
        # TODO platform = Livbirt(conf, utils)
        print("Todo: libvirt is not ready yet")
        sys.exit(0)
    else:
        raise Exception(Format.alert("Platform Error: {} is not applicable".format(conf.platform)))

    return platform

if __name__ == '__main__':
    main()
