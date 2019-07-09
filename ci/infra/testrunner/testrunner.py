#!/usr/bin/env python3 -Wd -b

"""
    Runs end-to-end product tests for v4+.
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

import sys
from argparse import ArgumentParser

from skuba import Skuba
from platforms import Platform
from tests import TestDriver
from utils import (BaseConfig, Format, Utils)

__version__ = "0.0.3"

def info(options):
    Utils(options.conf).info()

def cleanup(options):
    Platform.get_platform(options.conf).cleanup()
    Skuba.cleanup(options.conf)

def provision(options):
    Platform.get_platform(options.conf).provision(
                 num_master=options.master_count,
                 num_worker=options.worker_count)

def build_skuba(options):
    Skuba.build(options.conf)

def bootstrap(options):
    Skuba(options.conf).cluster_init()
    Skuba(options.conf).node_bootstrap()

def cluster_status(options):
    Skuba(options.conf).cluster_status()

def log(options):
    Skuba(options.conf).gather_logs()

def join_node(options):
        Skuba(options.conf).node_join(role=options.role, nr=options.node)

def remove_node(options):
        Skuba(options.conf).node_remove(role=options.role, nr=options.node)

def reset_node(options):
        Skuba(options.conf).node_reset(role=options.role, nr=options.node)

def test(options):
    TestDriver(options.conf).run(test_suite=options.test_suite, test=options.test,
            verbose=options.verbose, collect=options.collect)

def main():
    help = """
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.
    """
    parser = ArgumentParser(help)

    # Common parameters
    parser.add_argument("-v", "--vars", dest="yaml_path", default="vars/openstack.yaml",
                        help='path for platform yaml file. Default is vars/openstack.yaml \
                        in {workspace}/ci/infra/testrunner. eg: -v vars/myconfig.yaml')

    # Sub commands
    commands = parser.add_subparsers()

    cmd_info = commands.add_parser("info", help='ip info')
    cmd_info.set_defaults(func=info)

    cmd_log = commands.add_parser("log",  help="gather logs from nodes")
    cmd_log.set_defaults(func=log)

    cmd_cleanup = commands.add_parser("cleanup", help="cleanup created skuba environment")
    cmd_cleanup.set_defaults(func=cleanup)

    cmd_provision = commands.add_parser( "provision", help="provision nodes for cluster in \
                    your configured platform e.g: openstack, vmware.") 
    cmd_provision.set_defaults(func=provision)
    cmd_provision.add_argument("-m", "-master-count", dest="master_count", type=int, default=-1,
                    help='number of masters nodes to be deployed. eg: -m 2')
    cmd_provision.add_argument("-w", "--worker-count", dest="worker_count", type=int, default=-1,
                    help='number of workers nodes to be deployed. eg: -w 2')


    cmd_build = commands.add_parser("build-skuba", help="build skuba environment \
                        {workspace}/go/src/github.com/SUSE/skuba and build skuba \
                        in that directory")
    cmd_build.set_defaults(func=build_skuba)

    cmd_bootstrap = commands.add_parser("bootstrap", help="bootstrap k8s cluster with \
                        deployed nodes in your platform")
    cmd_bootstrap.set_defaults(func=bootstrap)

    cmd_status = commands.add_parser("status", help="check K8s cluster status")
    cmd_status.set_defaults(func=cluster_status)

    # common parameters for node commands
    node_args = ArgumentParser(add_help=False)
    node_args.add_argument("-r", "--role", dest="role",choices=["master","worker"], 
                   help='role of the node to be added or deleted. eg: --role master')
    node_args.add_argument("-n", "--node", dest="node", type=int,
                   help='node to be added or deleted.  eg: -n 0')

    cmd_join_node = commands.add_parser("join-node", parents=[node_args], 
                       help="add node in k8s cluster with the given role.")
    cmd_join_node.set_defaults(func=join_node)

    cmd_rem_node = commands.add_parser("remove-node", parents=[node_args],
                        help="remove node from k8s cluster.")
    cmd_rem_node.set_defaults(func=remove_node)

    cmd_reset_node = commands.add_parser("reset-node", parents=[node_args],
                        help="reset node reverting state previous to bootstap/join.")
    cmd_reset_node.set_defaults(func=reset_node)

    test_args = ArgumentParser(add_help=False)
    test_args.add_argument("-s", "--suit", dest="test_suite", help="test file name")
    test_args.add_argument("-t", "--test", dest="test", help="test to execute")
    test_args.add_argument("-l", "--list", dest="collect", action="store_true", default=False,
            help="only list tests to be executed")
    test_args.add_argument("-v", "--verbose", dest="verbose", action="store_true", default=False,
            help="show all output")
    cmd_test = commands.add_parser("test", parents=[test_args], help="execute tests")
    cmd_test.set_defaults(func=test)

    options = parser.parse_args()
    conf = BaseConfig(options.yaml_path)
    options.conf = conf
    options.func(options)

    sys.exit(0)

if __name__ == '__main__':
    main()
