
import os
import subprocess
import sys
from functools import wraps

import requests
from timeout_decorator import timeout

from utils.format import Format
from utils.constants import Constant

_stepdepth = 0


def step(f):
    @wraps(f)
    def wrapped(*args, **kwargs):
        global _stepdepth
        _stepdepth += 1
        print("{} entering {} {}".format(Format.DOT * _stepdepth, f.__name__,
                                  f.__doc__ or ""))
        r = f(*args, **kwargs)
        print("{}  exiting {}".format(Format.DOT_EXIT * _stepdepth, f.__name__))
        _stepdepth -= 1
        return r
    return wrapped


class Utils:

    def __init__(self, conf):
        self.conf = conf

    def runshellcommand(self, cmd, cwd=None, env=None):
        """Running shell command in {workspace} if cwd == None
           Eg) cwd is "skuba", cmd will run shell in {workspace}/skuba/
               cwd is None, cmd will run in {workspace}
               cwd is abs path, cmd will run in cwd
        Keyword arguments:
        cmd -- command to run
        cwd -- dir to run the cmd
        env -- environment variables
        """
        if not cwd:
            cwd = self.conf.workspace

        if not os.path.isabs(cwd):
            cwd = os.path.join(self.conf.workspace, cwd)

        if not os.path.exists(cwd):
            raise Exception(Format.alert("Directory {} does not exists".format(cwd)))

        print(Format.alert("$ {} > {}".format(cwd, cmd)))
        subprocess.check_call(cmd, cwd=cwd, shell=True, env=env)

    def authorized_keys(self):
        public_key_path = self.conf.ssh_key_option + ".pub"
        key_fn = self.conf.ssh_key_option
        self.runshellcommand("chmod 400 " + key_fn)
        with open(public_key_path ) as f:
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

    def runshellcommand_withoutput(self, cmd, ignore_errors=True):
        p = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        output, err = p.communicate()
        rc = p.returncode
        if not ignore_errors:
            if rc != 0:
                print(err)
                raise RuntimeError(Format.alert("Cannot run command {}{}\033[0m".format(cmd)))
        return output.decode()

    @timeout(60)
    @step
    def setup_ssh(self):

        self.runshellcommand("chmod 400 " + self.conf.ssh_key_option)
        print("Starting ssh-agent ")
        # use a dedicated agent to minimize stateful components
        sock_fn = os.path.join(self.conf.workspace, "ssh-agent-sock")
        try:
            self.runshellcommand("pkill -f 'ssh-agent -a {}'".format(sock_fn))
            print("Killed previous instance of ssh-agent")
        except:
            pass
        self.runshellcommand("ssh-agent -a {}".format(sock_fn))
        print("adding id_shared ssh key")
        self.runshellcommand("ssh-add " + self.conf.ssh_key_option, env={"SSH_AUTH_SOCK": sock_fn})

    @timeout(90)
    @step
    def git_rebase(self):
        if self.conf.git.branch_name.lower() == "master":
            print("Rebase not required for master.")
            return

        try:
            cmd = 'git -c "user.name={}" -c "user.email={}" \
                           rebase origin/master'.format(self.conf.git.change_author, self.conf.git.change_author_email)
            self.runshellcommand(cmd, cwd="skuba")
        except subprocess.CalledProcessError as ex:
            print(ex)
            print(Format.alert("Rebase failed, manual rebase is required."))
            self.runshellcommand("git rebase --abort", cwd="skuba")
            sys.exit(1)
        except Exception as ex:
            print(ex)
            print(Format.alert("Unknown error exiting."))
            sys.exit(2)

    @timeout(30)
    @step
    def info(self):
        """Node info"""
        print("Env vars: {}".format(sorted(os.environ)))

        self.runshellcommand('ip a')
        self.runshellcommand('ip r')
        self.runshellcommand('cat /etc/resolv.conf')

        try:
            r = requests.get('http://169.254.169.254/2009-04-04/meta-data/public-ipv4', timeout=2)
            r.raise_for_status()
        except (requests.HTTPError, requests.Timeout) as err:
            print(err)
            print(Format.alert('Meta Data service unavailable could not get external IP addr{}'))
        else:
            print('External IP addr: {}'.format(r.text))
