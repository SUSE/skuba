#!/usr/bin/env python3

import os
import subprocess
import sys

change_author = os.getenv('GIT_COMMITTER_NAME', 'CaaSP Jenkins')
author_email  = os.getenv('GIT_COMMITTER_EMAIL', 'containers-bugowner@suse.de')
change_id     = os.getenv("CHANGE_ID")
target        = os.getenv("CHANGE_TARGET")
workspace     = os.getenv("WORKSPACE","")

if not change_id:
    print("Not a PR. Rebase not required.")
    sys.exit(0)

try:
    cmd = (f'git -c "user.name={change_author}" -c "user.email={author_email}"'
           f' rebase origin/{target}')
    cwd = workspace+"/skuba"
    subprocess.check_call(cmd, cwd=cwd, shell=True)
except subprocess.CalledProcessError as ex:
    print(ex)
    print("Rebase failed, manual rebase is required.")
    subprocess.check_call("git rebase --abort", cwd=cwd)
    sys.exit(1)
except Exception as ex:
    print(ex)
    print("Unknown error exiting.")
    sys.exit(2)

