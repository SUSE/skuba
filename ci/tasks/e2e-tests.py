#!/usr/bin/env python3 

# !!!  This script is called by makefile only. !!!
# in root dir: make test-e2e

import subprocess
import os

# control-plane is --> caaspctl-init --controlplane var flag.
# if the ENV variable isn't set set it by default to "10.17.1.0", which is what a libvirt deployment set
if not "CONTROLPLANE" in os.environ:
  print("controlplane env var not defined, taking 10.17.1.0")
  os.environ['CONTROLPLANE'] = "10.17.1.0"

if not "MASTER00" in os.environ:
  print("master00 env var not defined, taking 10.17.2.0")
  os.environ['MASTER00'] = "10.17.2.0"

if not "WORKER00" in os.environ:
  print("worker00 env var not defined, taking 10.17.3.0")
  os.environ['WORKER00'] = "10.17.3.0"


subprocess.check_call("cd test && ginkgo -v --race --trace --progress core-features", shell=True, env=dict(os.environ))
