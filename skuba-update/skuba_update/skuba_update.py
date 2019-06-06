#!/usr/bin/env python
# -*- encoding: utf-8 -*-

# Copyright (c) 2019 SUSE LLC. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import subprocess
import re
from collections import namedtuple
from datetime import datetime
from pathlib import Path

# Since zypper 1.14.0, it will automatically create a /var/run/reboot-required
# text file whenever one of the applied patches requires the system to be
# rebooted. This will be used later on by kured.
REQUIRED_ZYPPER_VERSION = (1, 14, 0)

# The path to the reboot-needed file. This is the file that kured will be
# looking at.
REBOOT_REQUIRED_PATH = '/var/run/reboot-required'

# Exit codes as defined by zypper.

ZYPPER_EXIT_INF_UPDATE_NEEDED = 100
ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED = 101
ZYPPER_EXIT_INF_REBOOT_NEEDED = 102


def main():
    """
    main-entry point for program.
    """

    # First of all, check that we have the proper zypper version.
    if not check_version('zypper', REQUIRED_ZYPPER_VERSION):
        raise Exception('zypper version {0} or higher is required'.format(
            '.'.join([str(x) for x in REQUIRED_ZYPPER_VERSION])
        ))

    if os.geteuid() != 0:
        raise Exception('root privileges are required to run this tool')

    run_zypper_command(['zypper', 'ref', '-s'])
    run_zypper_command([
        'zypper', '--non-interactive',
        '--non-interactive-include-reboot-patches', 'patch'
    ])
    code = run_zypper_command(
        ['zypper', '--non-interactive-include-reboot-patches', 'patch-check']
    )
    if are_patches_available(code):
        run_zypper_command([
            'zypper', '--non-interactive',
            '--non-interactive-include-reboot-patches', 'patch'
        ])
    result = run_command(['zypper', 'ps', '-sss'])
    for service in result.output.splitlines():
        run_command(['systemctl', 'restart', service])


def is_zypper_error(code):
    """
    Returns true if the given code belongs to an error code according
    to zypper.
    """

    return code != 0 and code < 100


def are_patches_available(code):
    """
    Returns true if the given code is defined by zypper to mean that there are
    patches available.
    """

    return code == ZYPPER_EXIT_INF_UPDATE_NEEDED or \
        code == ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED


def is_reboot_needed(code):
    """
    Returns true of the given code is defined by zypper to mean that reboot is
    needed.
    """

    return code == ZYPPER_EXIT_INF_REBOOT_NEEDED


def log(message):
    """
    Prints the given message by prefixing a timestamp and the name of the
    program.
    """

    print('{0} [skuba-update] {1}'.format(
        datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        message)
    )


def run_zypper_command(command):
    """
    Run the given zypper command. The command is expected to be a tuple which
    also contains the 'zypper' string. It returns the exit code from zypper.
    """

    log('running \'{0}\''.format(' '.join(command)))
    process = subprocess.Popen(
        command,
        env=os.environ
    )
    process.communicate()
    if is_zypper_error(process.returncode):
        raise Exception('"{0}" failed'.format(' '.join(command)))

    # We trust the exit code of zypper over the presence of certain files.
    # Thus, we will make sure that the `REBOOT_REQUIRED_PATH` is present
    # depending on some exit codes.
    if is_reboot_needed(process.returncode):
        Path(REBOOT_REQUIRED_PATH).touch()

    return process.returncode


def run_command(command):
    """
    Runs the given command and it returns a named tuple containing: 'output',
    'error' and 'returncode'.
    """

    command_type = namedtuple(
        'command', ['output', 'error', 'returncode']
    )
    log('running \'{0}\''.format(' '.join(command)))
    process = subprocess.Popen(
        command,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=os.environ
    )
    output, error = process.communicate()
    if not error:
        error = bytes(b'(no output on stderr)')
    return command_type(
        output=output.decode(),
        error=error.decode(),
        returncode=process.returncode
    )


def check_version(call, version_waterline):
    """
    Checks if the given command version is equal or higher than
    the given version tuple. It returns true if the current command version is
    equal or higher to the given version waterline.

    It raises an exception if the execution fails or version can't be parsed.
    """

    arguments = [call] + ['--version']
    version_info = None
    try:
        output = run_command(arguments).output
        for line in output.splitlines():
            match = re.search('[0-9]+(.[0-9]+)*', line)
            if match:
                version_info = tuple(
                    int(elt) for elt in match.group(0).split('.')
                )
                break
        if version_info is None:
            raise Exception
    except Exception:
        message = 'Could not parse {0} version'.format(call)
        raise Exception(message)
    return version_info >= version_waterline


if __name__ == "__main__":  # pragma: no cover
    main()
