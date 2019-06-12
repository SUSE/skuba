#!/usr/bin/env python3 

import os
import subprocess
import sys

change_author = os.getenv('GIT_COMMITTER_NAME', 'CaaSP Jenkins')
author_email  = os.getenv('GIT_COMMITTER_EMAIL', 'containers-bugowner@suse.de')
branch_name   = os.getenv("BRANCH_NAME","master")
workspace     = os.getenv("WORKSPACE","")

if branch_name.lower() == "master":
    print("Rebase not required for master.")
    sys.exit(0)

try:
    cmd = 'git -c "user.name={}" -c "user.email={}" \
                   rebase origin/master'.format(change_author, author_email)
    cwd = workspace+"/skuba"
    subprocess.check_call(cmd, cwd=cwd, shell=True)
except subprocess.CalledProcessError as ex:
    print(ex)
    print(Format.alert("Rebase failed, manual rebase is required."))
    subprocess.check_call("git rebase --abort", cwd=cwd)
    sys.exit(1)
except Exception as ex:
    print(ex)
    print(Format.alert("Unknown error exiting."))
    sys.exit(2)

