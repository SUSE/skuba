#!/usr/bin/env python3 

# !!!  This script is called by makefile only. !!!
# in root dir: make test-e2e

import subprocess
import os

# control-plane is --> caaspctl-init --controlplane var flag.
# if the ENV variable isn't set set it by default to "10.17.1.0", which is what a libvirt deployment set
if not "control-plane" in os.environ:
  os.environ['control-plane'] = "10.17.1.0"

# this can be configured later. you can check the upstream doc. we have everything random here
subprocess.check_call("cd test && ginkgo -r --randomizeAllSpecs --randomizeSuites --cover --trace --race --progress -v", shell=True, env=dict(os.environ))
