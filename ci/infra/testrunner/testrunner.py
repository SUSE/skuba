#!/usr/bin/env python3 -Wd -b

"""
    Runs end-to-end product tests for v4+.
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

import logging
import sys
from argparse import REMAINDER, ArgumentParser


import platforms
from skuba import Skuba
from tests import TestDriver
from utils import BaseConfig, Logger, Utils

__version__ = "0.0.3"

logger = logging.getLogger("testrunner")


def info(options):
    print(Utils(options.conf).info())


def cleanup(options):
    platforms.get_platform(options.conf, options.platform).cleanup()
    Skuba.cleanup(options.conf)


def provision(options):
    platforms.get_platform(options.conf, options.platform).provision(
        num_master=options.master_count,
        num_worker=options.worker_count)


def bootstrap(options):
    skuba = Skuba(options.conf, options.platform)
    skuba.cluster_init(kubernetes_version=options.kubernetes_version)
    skuba.node_bootstrap()


def cluster_status(options):
    print(Skuba(options.conf, options.platform).cluster_status())


def cluster_upgrade_plan(options):
    Skuba(options.conf, options.platform).cluster_upgrade_plan()


def get_logs(options):
    platform_logging_errors = platforms.get_platform(
        options.conf, options.platform).gather_logs()

    if platform_logging_errors:
        raise Exception("Failure(s) while collecting logs")


def join_node(options):
    Skuba(options.conf, options.platform).node_join(
        role=options.role, nr=options.node)


def remove_node(options):
    Skuba(options.conf, options.platform).node_remove(
        role=options.role, nr=options.node)


def node_upgrade(options):
    Skuba(options.conf, options.platform).node_upgrade(
        action=options.upgrade_action, role=options.role, nr=options.node)


def test(options):
    TestDriver(options.conf, options.platform).run(test_suite=options.test_suite, test=options.test,
                                                   verbose=options.verbose, collect=options.collect,
                                                   skip_setup=options.skip_setup)


def ssh(options):
    platforms.get_platform(options.conf, options.platform).ssh_run(
        role=options.role, nr=options.node, cmd=" ".join(options.cmd))


def main():
    help_str = """
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.
    """
    parser = ArgumentParser(help_str)

    # Common parameters
    parser.add_argument("-v", "--vars", dest="yaml_path", default="vars.yaml",
                        help='path for platform yaml file. Default is vars.yaml. eg: -v myconfig.yaml')
    parser.add_argument("-p", "--platform",
                        default="openstack",
                        choices=["openstack", "vmware",
                                 "bare-metal", "libvirt"],
                        help="The platform you're targeting. Defaults to openstack")
    parser.add_argument("-l", "--log-level", dest="log_level", default=None, help="log level",
                        choices=["DEBUG", "INFO", "WARNING", "ERROR"])

    # Sub commands
    commands = parser.add_subparsers(help="command", dest="command")

    cmd_info = commands.add_parser("info", help='ip info')
    cmd_info.set_defaults(func=info)

    cmd_log = commands.add_parser("get_logs",  help="gather logs from nodes")
    cmd_log.set_defaults(func=get_logs)

    cmd_cleanup = commands.add_parser(
        "cleanup", help="cleanup created skuba environment")
    cmd_cleanup.set_defaults(func=cleanup)

    cmd_provision = commands.add_parser("provision", help="provision nodes for cluster in "
                                                          "your configured platform e.g: openstack, vmware.")
    cmd_provision.set_defaults(func=provision)
    cmd_provision.add_argument("-m", "--master-count", dest="master_count", type=int, default=-1,
                               help='number of masters nodes to be deployed. eg: -m 2')
    cmd_provision.add_argument("-w", "--worker-count", dest="worker_count", type=int, default=-1,
                               help='number of workers nodes to be deployed. eg: -w 2')

    cmd_bootstrap = commands.add_parser("bootstrap", help="bootstrap k8s cluster with \
                        deployed nodes in your platform")
    cmd_bootstrap.add_argument("-k", "--kubernetes-version", help="kubernetes version",
                               dest="kubernetes_version", default=None)
    cmd_bootstrap.set_defaults(func=bootstrap)

    cmd_status = commands.add_parser("status", help="check K8s cluster status")
    cmd_status.set_defaults(func=cluster_status)

    cmd_cluster_upgrade_plan = commands.add_parser(
        "cluster-upgrade-plan", help="Cluster upgrade plan")
    cmd_cluster_upgrade_plan.set_defaults(func=cluster_upgrade_plan)

    # common parameters for node commands
    node_args = ArgumentParser(add_help=False)
    node_args.add_argument("-r", "--role", dest="role", choices=["master", "worker"],
                           help='role of the node to be added or deleted. eg: --role master')
    node_args.add_argument("-n", "--node", dest="node", type=int,
                           help='node to be added or deleted.  eg: -n 0')

    cmd_join_node = commands.add_parser("join-node", parents=[node_args],
                                        help="add node in k8s cluster with the given role.")
    cmd_join_node.set_defaults(func=join_node)

    cmd_rem_node = commands.add_parser("remove-node", parents=[node_args],
                                       help="remove node from k8s cluster.")
    cmd_rem_node.set_defaults(func=remove_node)

    cmd_node_upgrade = commands.add_parser("node-upgrade", parents=[node_args],
                                           help="upgrade kubernetes version in node")
    cmd_node_upgrade.add_argument("-a", "--action", dest="upgrade_action",
                                  help="action: plan or apply upgrade", choices=["plan", "apply"])
    cmd_node_upgrade.set_defaults(func=node_upgrade)

    ssh_args = ArgumentParser(add_help=False)
    ssh_args.add_argument("-c", "--cmd", dest="cmd", nargs=REMAINDER, help="remote command and its arguments. e.g ls -al. Must be last argument for ssh command")
    cmd_ssh = commands.add_parser("ssh", parents=[node_args, ssh_args], help="Execute command in node via ssh.")
    cmd_ssh.set_defaults(func=ssh)

    test_args = ArgumentParser(add_help=False)
    test_args.add_argument(
        "-s", "--suite", dest="test_suite", help="test file name")
    test_args.add_argument("-t", "--test", dest="test", help="test to execute")
    test_args.add_argument("-l", "--list", dest="collect", action="store_true", default=False,
                           help="only list tests to be executed")
    test_args.add_argument("-v", "--verbose", dest="verbose", action="store_true", default=False,
                           help="show all output")
    test_args.add_argument("--skip-setup", help="setup step to skip")
    cmd_test = commands.add_parser(
        "test", parents=[test_args], help="execute tests")
    cmd_test.set_defaults(func=test)

    options = parser.parse_args()
    try:
        conf = BaseConfig(options.yaml_path)
        Logger.config_logger(conf, level=options.log_level)
        options.conf = conf
        options.func(options)
    except Exception as ex:
        logger.error("Exception executing testrunner command '{}': {}".format(
            options.command, ex, exc_info=True))
        sys.exit(255)


if __name__ == '__main__':
    main()
