import glob
import logging
import os
import shutil
import subprocess
from functools import wraps

import requests
from timeout_decorator import timeout

from utils.format import Format
from utils.constants import Constant

logger = logging.getLogger('testrunner')

_stepdepth = 0


def step(f):
    @wraps(f)
    def wrapped(*args, **kwargs):
        global _stepdepth
        _stepdepth += 1
        logger.debug("{} entering {} {}".format(Format.DOT * _stepdepth, f.__name__,
                                  f.__doc__ or ""))
        r = f(*args, **kwargs)
        logger.debug("{}  exiting {}".format(Format.DOT_EXIT * _stepdepth, f.__name__))
        _stepdepth -= 1
        return r
    return wrapped


class Utils:

    def __init__(self, conf):
        self.conf = conf

    @staticmethod
    def chmod_recursive(directory, permissions):
        os.chmod(directory, permissions)

        for file in glob.glob(os.path.join(directory, "**/*"), recursive=True):
            try:
                os.chmod(file, permissions)
            except Exception as ex:
                logger.exception(ex)

    @staticmethod
    def cleanup_file(file):
        if os.path.exists(file):
            logger.info(f"Cleaning up {file}")
            try:
                if os.path.isfile(file):
                    os.remove(file)
                else:
                    shutil.rmtree(file)
            except Exception as ex:
                logger.exception(ex)
        else:
            logger.warning(f"Could not clean up {file}")

    @staticmethod
    def cleanup_files(files):
        """Remove any files or dirs in a list if they exist"""
        for file in files:
            Utils.cleanup_file(file)

    def collect_remote_logs(self, ip_address, logs, store_path):
        """
        Collect logs from a remote machine
        :param ip_address: (str) IP of the machine to collect the logs from
        :param logs: (dict: list) The different logs to collect {"files": [], "dirs": [], ""services": []}
        :param store_path: (str) Path to copy the logs to
        :return: (bool) True if there was an error while collecting the logs
        """
        logging_errors = False

        for log in logs.get("files", []):
            try:
                self.scp_file(ip_address, log, store_path)
            except Exception as ex:
                logger.debug(f"Error while collecting {log} from {ip_address}\n {ex}")
                logging_errors = True

        for log in logs.get("dirs", []):
            try:
                self.rsync(ip_address, log, store_path)
            except Exception as ex:
                logger.debug(f"Error while collecting {log} from {ip_address}\n {ex}")
                logging_errors = True

        for service in logs.get("services", []):
            try:
                self.ssh_run(ip_address, f"sudo journalctl -xeu {service} > {service}.log")
                self.scp_file(ip_address, f"{service}.log", store_path)
            except Exception as ex:
                logger.debug(f"Error while collecting {service}.log from {ip_address}\n {ex}")
                logging_errors = True

        return logging_errors

    def authorized_keys(self):
        public_key_path = self.conf.ssh_key_option + ".pub"
        os.chmod(self.conf.ssh_key_option, 0o400)

        with open(public_key_path) as f:
            pubkey = f.read().strip()
        return pubkey

    def gorun(self, cmd=None, extra_env=None):
        """Running go command in {workspace}/go/src/github.com/SUSE/skuba"""
        env = {
            'GOPATH': os.path.join(self.conf.workspace,'go'),
            'PATH': os.environ['PATH'],
            'HOME': os.environ['HOME']
        }

        if extra_env:
            env.update(extra_env)

        self.runshellcommand(cmd, cwd="go/src/github.com/SUSE/skuba", env=env)

    def ssh_run(self, ipaddr, cmd):
        key_fn = self.conf.ssh_key_option
        cmd = "ssh " + Constant.SSH_OPTS + " -i {key_fn} {username}@{ip} -- '{cmd}'".format(
            key_fn=key_fn, ip=ipaddr, cmd=cmd, username=self.conf.nodeuser)
        self.runshellcommand(cmd)

    def scp_file(self, ip_address, remote_file_path, local_file_path):
        """
        Copies a remote file from the given ip to the give path
        :param ip_address: (str) IP address of the node to copy from
        :param remote_file_path: (str) Path of the file to be copied
        :param local_file_path: (str) Path where to store the log
        :return:
        """
        cmd = (f"scp {Constant.SSH_OPTS} -i {self.conf.ssh_key_option}"
               f" {self.conf.nodeuser}@{ip_address}:{remote_file_path} {local_file_path}")
        self.runshellcommand(cmd)

    def rsync(self, ip_address, remote_dir_path, local_dir_path):
        """
        Copies a remote dir from the given ip to the give path
        :param ip_address: (str) IP address of the node to copy from
        :param remote_dir_path: (str) Path of the dir to be copied
        :param local_dir_path: (str) Path where to store the dir
        :return:
        """
        cmd = (f'rsync -avz --no-owner --no-perms -e "ssh {Constant.SSH_OPTS} -i {self.conf.ssh_key_option}"  '
               f'--rsync-path="sudo rsync" --ignore-missing-args {self.conf.nodeuser}@{ip_address}:{remote_dir_path} '
               f'{local_dir_path}')
        self.runshellcommand(cmd)

    def runshellcommand(self, cmd, cwd=None, env={}, ignore_errors=False):
        """Running shell command in {workspace} if cwd == None
           Eg) cwd is "skuba", cmd will run shell in {workspace}/skuba/
               cwd is None, cmd will run in {workspace}
               cwd is abs path, cmd will run in cwd
        Keyword arguments:
        cmd -- command to run
        cwd -- dir to run the cmd
        env -- environment variables
        ignore_errors -- don't raise exception if command fails
        """
        if not cwd:
            cwd = self.conf.workspace

        if not os.path.isabs(cwd):
            cwd = os.path.join(self.conf.workspace, cwd)

        if not os.path.exists(cwd):
            raise Exception(Format.alert("Directory {} does not exists".format(cwd)))

        if logging.DEBUG >= logger.level:
            logger.debug("Executing command\n"
                         "    cwd: {} \n"
                         "    env: {}\n"
                         "    cmd: {}".format(cwd, str(env) if env else "{}", cmd)) 
        else:
            logger.info("Executing command {}".format(cmd))
        
        p = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, 
                 env=env, cwd=cwd)
        logger.debug(p.stdout.decode())
        if p.returncode != 0:
           logger.error(p.stderr)
           if not ignore_errors:
                raise RuntimeError("Error executing command {}".format(cmd))
        return p.stdout.decode()

    def ssh_sock_fn(self):
        return os.path.join(self.conf.workspace, "ssh-agent-sock")

    @timeout(60)
    @step
    def setup_ssh(self):
        os.chmod(self.conf.ssh_key_option, 0o400)

        # use a dedicated agent to minimize stateful components
        sock_fn = self.ssh_sock_fn()
        try:
            self.runshellcommand("pkill -f 'ssh-agent -a {}'".format(sock_fn))
            logger.warning("Killed previous instance of ssh-agent")
        except:
            pass
        self.runshellcommand("ssh-agent -a {}".format(sock_fn))
        self.runshellcommand("ssh-add " + self.conf.ssh_key_option, env={"SSH_AUTH_SOCK": sock_fn})

    @timeout(30)
    @step
    def info(self):
        """Node info"""
        info_lines  = "Env vars: {}\n".format(sorted(os.environ))
        info_lines += self.runshellcommand('ip a')
        info_lines += self.runshellcommand('ip r')
        info_lines += self.runshellcommand('cat /etc/resolv.conf')

        # TODO: the logic for retrieving external is platform depedant and should be
        # moved to the corresponding platform
        try:
            r = requests.get('http://169.254.169.254/2009-04-04/meta-data/public-ipv4', timeout=2)
            r.raise_for_status()
        except (requests.HTTPError, requests.Timeout) as err:
            logger.warning(f'Meta Data service unavailable could not get external IP addr{err}')
        else:
            info_lines += 'External IP addr: {}'.format(r.text)

        return info_lines
