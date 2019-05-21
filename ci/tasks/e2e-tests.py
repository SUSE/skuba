#!/usr/bin/env python3 

# !!!  This script is called by makefile only. !!!
# in root dir: make test-e2e

import subprocess
import os
import json
import sys

try:
  subprocess.check_call("ginkgo")
except Exception  as ex:
  print("Please install - ginkgo - before!")
  sys.exit(1)
  
# 1) set all IPS variable individually ( with the env. variables)
# 2) read this IPS from a tfstate which will set the env. variables.

if os.environ.get('IP_FROM_TF_STATE') == 'True' or 'TRUE':
  # we need to know which provider is beeing used to read the tfstates.
  if not "PLATFORM" in os.environ:
    raise("you need to set PLATFORM ENV. variable with LOAD_IP_FROM_TF_STATE env var")

  tf = os.path.join("ci/infra/{0}".format(os.environ['PLATFORM'].lower()), "terraform.tfstate")
  with open(tf) as f:
    tf_state = json.load(f)
  os.environ['CONTROLPLANE'] = tf_state["modules"][0]["outputs"]["ip_ext_load_balancer"]["value"]
  os.environ['MASTER00'] = tf_state["modules"][0]["outputs"]["ip_masters"]["value"][0]
  os.environ['WORKER00'] = tf_state["modules"][0]["outputs"]["ip_workers"]["value"][0]

subprocess.check_call("cd test && ginkgo -v --race --trace --progress core-features", shell=True, env=dict(os.environ))
