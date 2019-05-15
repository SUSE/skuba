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


# we will have here a set of core feature, actually only 1, where we setup the cluster with caaspctl init etc. this need to run first the rest of feature 
# will be random and idempotent.

# TODO-01: @dmaiocchi: setup first serial features

subprocess.check_call("cd test && ginkgo --race --progress core-features", shell=True, env=dict(os.environ))


## TODO-02: this are parallel feature
## SECONDARY FEATURES:

# idempotent features run in random order
# this can be configured later. you can check the upstream doc. we have everything random here

# debug mode
#if os.environ.get('DEBUG')=='True':
#    subprocess.check_call("cd test && ginkgo -r --randomizeAllSpecs --randomizeSuites --trace --race --progress -v", shell=True, env=dict(os.environ))

#else:
#    subprocess.check_call("cd test && ginkgo -r --randomizeAllSpecs --randomizeSuites --race --progress", shell=True, env=dict(os.environ))
