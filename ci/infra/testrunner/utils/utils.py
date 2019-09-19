import glob
import hashlib
import logging
import os
import shutil
import subprocess
from functools import wraps
from tempfile import gettempdir
from threading import Thread

import requests
from timeout_decorator import timeout

from utils.constants import Constant
from utils.format import Format

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
        logger.debug("{}  exiting {}".format(
            Format.DOT_EXIT * _stepdepth, f.__name__))
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
                try:
                    # Attempt to remove the file first, because a socket (e.g.
                    # ssh-agent) is not a file but has to be removed like one.
                    os.remove(file)
                except IsADirectoryError:
                    shutil.rmtree(file)
            except Exception as ex:
                logger.exception(ex)
        else:
            logger.warning(f"Nothing to clean up for {file}")

    @staticmethod
    def cleanup_files(files):
        """Remove any files or dirs in a list if they exist"""
        for file in files:
            Utils.cleanup_file(file)

    def ssh_cleanup(self):
        """Remove ssh sock files"""
        # TODO: also kill ssh agent here? maybe move pkill to kill_ssh_agent()?
        sock_file = self.ssh_sock_fn()
        sock_dir = os.path.dirname(sock_file)
        try:
            Utils.cleanup_file(sock_file)
            # also remove tempdir if it's empty afterwards
            if 0 == len(os.listdir(sock_dir)):
                os.rmdir(sock_dir)
            else:
                logger.warning(f"Dir {sock_dir} not empty; leaving it")
        except FileNotFoundError:
            pass
        except OSError as ex:
            logger.exception(ex)

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
                logger.debug(
                    f"Error while collecting {log} from {ip_address}\n {ex}")
                logging_errors = True

        for log in logs.get("dirs", []):
            try:
                self.rsync(ip_address, log, store_path)
            except Exception as ex:
                logger.debug(
                    f"Error while collecting {log} from {ip_address}\n {ex}")
                logging_errors = True

        for service in logs.get("services", []):
            try:
                self.ssh_run(
                    ip_address, f"sudo journalctl -xeu {service} > {service}.log")
                self.scp_file(ip_address, f"{service}.log", store_path)
            except Exception as ex:
                logger.debug(
                    f"Error while collecting {service}.log from {ip_address}\n {ex}")
                logging_errors = True

        return logging_errors

    def authorized_keys(self):
        public_key_path = self.conf.ssh_key + ".pub"
        os.chmod(self.conf.ssh_key, 0o400)

        with open(public_key_path) as f:
            pubkey = f.read().strip()
        return pubkey

    def ssh_run(self, ipaddr, cmd):
        key_fn = self.conf.ssh_key
        cmd = "ssh " + Constant.SSH_OPTS + " -i {key_fn} {username}@{ip} -- '{cmd}'".format(
            key_fn=key_fn, ip=ipaddr, cmd=cmd, username=self.conf.nodeuser)
        return self.runshellcommand(cmd)

    def scp_file(self, ip_address, remote_file_path, local_file_path):
        """
        Copies a remote file from the given ip to the give path
        :param ip_address: (str) IP address of the node to copy from
        :param remote_file_path: (str) Path of the file to be copied
        :param local_file_path: (str) Path where to store the log
        :return:
        """
        cmd = (f"scp {Constant.SSH_OPTS} -i {self.conf.ssh_key}"
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
        cmd = (f'rsync -avz --no-owner --no-perms -e "ssh {Constant.SSH_OPTS} -i {self.conf.ssh_key}"  '
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
            raise FileNotFoundError(Format.alert("Directory {} does not exists".format(cwd)))

        if logging.DEBUG >= logger.level:
            logger.debug("Executing command\n"
                         "    cwd: {} \n"
                         "    env: {}\n"
                         "    cmd: {}".format(cwd, str(env) if env else "{}", cmd))
        else:
            logger.info("Executing command {}".format(cmd))

        stdout, stderr = [], []
        p = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=True, env=env, cwd=cwd)
        stdoutStreamer = Thread(target = self.read_fd, args = (p, p.stdout, logger.debug, stdout))
        stderrStreamer = Thread(target = self.read_fd, args = (p, p.stderr, logger.error, stderr))
        stdoutStreamer.start()
        stderrStreamer.start()
        stdoutStreamer.join()
        stderrStreamer.join()
        # this is redundant, at this point threads were joined and they waited for the subprocess
        # to exit, however it should not hurt to explicitly wait for it again (no-op).
        p.wait()
        stdout, stderr = "".join(stdout), "".join(stderr)

        if p.returncode != 0:
            if not ignore_errors:
                raise RuntimeError("Error executing command {}".format(cmd))
            else:
                return stderr
        return stdout

    def ssh_sock_fn(self):
        """generate path to ssh socket

        A socket path can't be over 107 chars on Linux, so generate a short
        hash of the workspace and use that in $TMPDIR (usually /tmp) so we have
        a predictable, test-unique, fixed-length path.
        """
        path = os.path.join(
            gettempdir(),
            hashlib.md5(self.conf.workspace.encode()).hexdigest(),
            "ssh-agent-sock"
        )
        maxl = 107
        if len(path) > maxl:
            raise Exception(f"Socket path '{path}' len {len(path)} > {maxl}")
        return path

    def read_fd(self, proc, fd, logger_func, output):
        """Read from fd, logging using logger_func

        Read from fd, until proc is finished. All contents will
        also be appended onto output."""
        while True:
            contents = fd.readline().decode()
            if contents == '' and proc.poll() is not None:
                return
            if contents:
                output.append(contents)
                logger_func(contents.strip())

    @timeout(60)
    @step
    def setup_ssh(self):
        os.chmod(self.conf.ssh_key, 0o400)

        # use a dedicated agent to minimize stateful components
        sock_fn = self.ssh_sock_fn()
        # be sure directory containing socket exists and socket doesn't exist
        if os.path.exists(sock_fn):
            try:
                if os.path.isdir(sock_fn):
                    os.path.rmdir(sock_fn)  # rmdir only removes an empty dir
                else:
                    os.remove(sock_fn)
            except FileNotFoundError:
                pass
        try:
            os.mkdir(os.path.dirname(sock_fn), mode=0o700)
        except FileExistsError:
            if os.path.isdir(os.path.dirname(sock_fn)):
                pass
            else:
                raise
        # clean up old ssh agent process(es)
        try:
            self.runshellcommand("pkill -f 'ssh-agent -a {}'".format(sock_fn))
            logger.warning("Killed previous instance of ssh-agent")
        except:
            pass
        self.runshellcommand("ssh-agent -a {}".format(sock_fn))
        self.runshellcommand(
            "ssh-add " + self.conf.ssh_key, env={"SSH_AUTH_SOCK": sock_fn})

    @timeout(30)
    @step
    def info(self):
        """Node info"""
        info_lines = "Env vars: {}\n".format(sorted(os.environ))
        info_lines += self.runshellcommand('ip a')
        info_lines += self.runshellcommand('ip r')
        info_lines += self.runshellcommand('cat /etc/resolv.conf')

        # TODO: the logic for retrieving external is platform depedant and should be
        # moved to the corresponding platform
        try:
            r = requests.get(
                'http://169.254.169.254/2009-04-04/meta-data/public-ipv4', timeout=2)
            r.raise_for_status()
        except (requests.HTTPError, requests.Timeout) as err:
            logger.warning(
                f'Meta Data service unavailable could not get external IP addr{err}')
        else:
            info_lines += 'External IP addr: {}'.format(r.text)

        return info_lines
