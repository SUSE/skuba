#!/usr/bin/env python3 -Wd -b

"""
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

import sys
from argparse import ArgumentParser

from platforms import Openstack
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
        Skuba(conf).cleanup()
    elif options.apply_terraform:
        get_platform(conf).apply_terraform()
    elif options.create_skuba:
        Skuba(conf).create_skuba()
    elif options.log:
        Skuba(conf).gather_logs()

    sys.exit(0)

def get_platform(conf):
    if conf.platform == "openstack":
        platform = Openstack(conf)
    elif conf.platform == "vmware":
        # TODO platform = VMware(conf, utils)
        print("Todo: VMware is not ready yet")
        sys.exit(0)
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
