#!/usr/bin/env python3 

# This script is called by makefile only.

import subprocess

# todo: check ginkgo flags. 
subprocess.check_call("cd test && ginkgo -r --randomizeAllSpecs --randomizeSuites --cover --trace --race --progress -v", shell=True)
