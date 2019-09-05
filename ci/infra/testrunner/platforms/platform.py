import logging
import os

import requests
from timeout_decorator import timeout

from utils import (step, Utils)

logger = logging.getLogger('testrunner')


class Platform:
    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(conf)

        # Which logs will be collected from the nodes
        self.logs = {
            "files": [],
            "dirs": ["/var/log/pods"],
            "services": ["kubelet"]
        }

        # Files that will be deleted during the cleanup stage
        self.tmp_files = []

    @step
    def cleanup(self):
        """Clean up"""
        try:
            self._cleanup_platform()
        except Exception as ex:
            logger.warning("Received the following error '{}'\nAttempting to finish cleanup".format(ex))
            raise Exception("Failure(s) during cleanup")
        finally:
            self.utils.cleanup_files(self.tmp_files)
            self.utils.ssh_cleanup()

    @timeout(600)
    @step
    def gather_logs(self):
        logging_errors = False

        node_ips = {
            "master": self.get_nodes_ipaddrs("master"),
            "worker": self.get_nodes_ipaddrs("worker")
        }

        if not os.path.isdir(self.conf.log_dir):
            os.mkdir(self.conf.log_dir)
            logger.info(f"Created log dir {self.conf.log_dir}")

        for node_type in node_ips:
            for ip_address in node_ips[node_type]:
                node_log_dir = self._create_node_log_dir(ip_address, node_type, self.conf.log_dir)
                logging_error = self.utils.collect_remote_logs(ip_address, self.logs, node_log_dir)

                if logging_error:
                    logging_errors = logging_error

        platform_log_error = self._get_platform_logs()

        if platform_log_error:
            logging_errors = platform_log_error

        return logging_errors

    def get_lb_ipaddr(self):
        """
        Get the IP of the Load Balancer
        :return:
        """
        pass

    def get_nodes_ipaddrs(self, role):
        """
        Get the IP addresses of the given type of node
        :param role: the type of node
        :return:
        """
        return []

    def get_num_nodes(self, role):
        """
        Get the number of nodes of a  given type
        :param role: the type of node
        :return: num of nodes
        """
        pass

    @step
    def provision(self, num_master=-1, num_worker=-1, retries=4):
        """Provision a cluster"""
        if num_master > -1 or num_worker > -1:
            logger.warning("Overriding number of nodes")
            if num_master > -1:
                self.conf.master.count = num_master
                logger.warning("   Masters:{} ".format(num_master))

            if num_worker > -1:
                self.conf.worker.count = num_worker
                logger.warning("   Workers:{} ".format(num_worker))

        # TODO: define the number of retries as a configuration parameter
        for i in range(0, retries):
            retry = i + 1

            try:
                self._provision_platform()
                break
            except Exception as ex:
                logger.warning(f"Provision attempt {retry}/{retries} failed")
                if retry == retries:
                    raise Exception(f"Failed {self.__class__.__name__} deployment") from ex

    def ssh_run(self, role, nr, cmd):
        ip_addrs = self.get_nodes_ipaddrs(role)
        if nr >= len(ip_addrs):
            raise ValueError(f'Node {role}-{nr} not deployed in platform')

        return self.utils.ssh_run(ip_addrs[nr], cmd)

    @staticmethod
    def _create_node_log_dir(ip_address, node_type, log_dir_path):
        node_log_dir_path = os.path.join(log_dir_path, f"{node_type}_{ip_address.replace('.', '_')}")

        if not os.path.isdir(node_log_dir_path):
            os.mkdir(node_log_dir_path)
            logger.info(f"Created log dir {node_log_dir_path}")

        return node_log_dir_path

    def _cleanup_platform(self):
        """Platform specific cleanup. Expected to be overridden by platforms"""

    def _env_setup_cmd(self):
        """Returns the command for setting up the platform environment"""
        return ""

    def _provision_platform(self):
        """Platform specific provisioning"""

    def _get_platform_logs(self):
        """Platform specific logs to collect. Expected to be overridden by platforms"""
        return False

    def all_apiservers_responsive(self):
        """Check if all apiservers are responsive to make sure the load balancer can function correctly"""
        ip_addrs = self.get_nodes_ipaddrs("master")
        for ip in ip_addrs:
            try:
                response = requests.get("https://{}:6443/healthz".format(ip), verify=False)
            except requests.exceptions.RequestException:
                return False
            if response.status_code != 200:
                return False
        return True

    def setup_cloud_provider(self):
        raise ValueError("Cloud provider is not supported for this platform")