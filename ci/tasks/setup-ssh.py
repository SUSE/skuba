#!/usr/bin/env python3 

# !!!  This script is called by makefile only. !!!

import subprocess
import os

subprocess.check_call("chmod 400 ci/infra/id_shared", shell=True)
print("Starting ssh-agent ")
# use a dedicated agent to minimize stateful components
sock_fn = "/tmp/ssh-agent-sock"
try:
    subprocess.check_call("rm " + sock_fn, shell=True)
    subprocess.check_call("pkill -f 'ssh-agent -a {}'".format(sock_fn), shell=True)
    print("Killed previous instance of ssh-agent")
except:
    pass
subprocess.check_call("ssh-agent -a {}".format(sock_fn), shell=True)
print("adding id_shared ssh key")
subprocess.check_call("ssh-add " + "ci/infra/id_shared", env={"SSH_AUTH_SOCK": sock_fn}, shell=True)
