#!/usr/bin/env python
# -*- encoding: utf-8 -*-

# Copyright (c) 2019 SUSE LLC.
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
import json
from collections import namedtuple
from datetime import datetime
from pathlib import Path
from xml.etree import ElementTree

# Since zypper 1.14.0, it will automatically create a /var/run/reboot-required
# text file whenever one of the applied patches requires the system to be
# rebooted. This will be used later on by kured.
# Since zypper 1.14.15, there is a subcommand to check if a reboot is needed
# `zypper needs-rebooting`
REQUIRED_ZYPPER_VERSION = (1, 14, 15)

# The path where zypper might write if it has detected that a patch/update that
# has been installed requires the machine to reboot in order to work properly.
ZYPPER_REBOOT_NEEDED_PATH = '/var/run/reboot-needed'

# The path to the reboot-needed file. This is the file that kured will be
# looking at.
REBOOT_REQUIRED_PATH = '/var/run/reboot-required'

# Exit codes as defined by zypper.

ZYPPER_EXIT_INF_REBOOT_NEEDED = 102
ZYPPER_EXIT_INF_RESTART_NEEDED = 103

# The path to the kubelet config used for running kubectl commands
KUBECONFIG_PATH = '/etc/kubernetes/kubelet.conf'

# The key for the annotation on the Kubernetes node for disruptive updates.
KUBE_DISRUPTIVE_UPDATES_KEY = 'caasp.suse.com/has-disruptive-updates'


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

    update()
    restart_services()
    annotate_resources()


def update():
    """
    Performs an update operation.
    """

    run_zypper_command(['zypper', 'ref', '-s'])
    code = run_zypper_patch()
    if is_restart_needed(code):
        run_zypper_patch()


def restart_services():
    """
    Restart services which are reported to have been updated and need a
    restart.
    """

    result = run_command(['zypper', 'ps', '-sss'])
    for service in result.output.splitlines():
        run_command(['systemctl', 'restart', service])


def annotate_resources():
    """
    Annotate all the needed Kubernetes resources for the current conditions.
    """

    if interruptive_updates_available():
        annotate(
            'nodes', node_name_from_machine_id(),
            KUBE_DISRUPTIVE_UPDATES_KEY, 'yes'
        )


def is_zypper_error(code):
    """
    Returns true if the given code belongs to an error code according
    to zypper.
    """

    return code != 0 and code < 100


def is_restart_needed(code):
    """
    Returns true of the given code is defined by zypper to mean that restart is
    needed (zypper itself has been updated).
    """

    return code == ZYPPER_EXIT_INF_RESTART_NEEDED


def is_reboot_needed():
    """
    Returns true if reboot is needed.
    """

    run_zypper_command(
        ['zypper', 'needs-rebooting']
    ) == ZYPPER_EXIT_INF_REBOOT_NEEDED


def interruptive_updates_available():
    """
    Returns True if there are interruptive updates available. Otherwise it
    returns False.
    """

    res = run_zypper_command(
        ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
        True
    )

    try:
        tree = ElementTree.fromstring(res.output)
    except ElementTree.ParseError:
        return False

    us = tree.find('update-status')
    if us is None:
        return False
    for update in us.find('update-list'):
        attr = update.attrib.get('interactive', '')
        if is_not_false_str(attr) and attr != 'reboot':
            return True
    return False


def is_not_false_str(string):
    """
    Returns true if the given string contains a non-falsey value.
    """

    return string is not None and string != '' and string != 'false'


def log(message):
    """
    Prints the given message by prefixing a timestamp and the name of the
    program.
    """

    print('{0} [skuba-update] {1}'.format(
        datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        message)
    )


def run_zypper_command(command, needsOutput=False):
    """
    Run the given zypper command. The command is expected to be a tuple which
    also contains the 'zypper' string. It returns the exit code from zypper.
    """

    process = run_command(command)
    if is_zypper_error(process.returncode):
        raise Exception('"{0}" failed'.format(' '.join(command)))
    if needsOutput:
        return process
    return process.returncode


def run_zypper_patch():
    code = run_zypper_command([
        'zypper', '--non-interactive',
        '--non-interactive-include-reboot-patches', 'patch'
    ])

    # There are two instances in which we should create the
    # REBOOT_REQUIRED_PATH file:
    #
    # 1. Zypper returned an exit code telling us to restart the system.
    # 2. `zypper needs-rebooting` returns ZYPPER_EXIT_INF_REBOOT_NEEDED.
    if code == ZYPPER_EXIT_INF_REBOOT_NEEDED or is_reboot_needed():
        Path(REBOOT_REQUIRED_PATH).touch()
    return code


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


def node_name_from_machine_id():
    """
    Reads the kubernetes node name from the machine-id
    """

    with open('/etc/machine-id') as machine_id_file:
        machine_id = machine_id_file.read().strip()

    nodes_raw_json = run_command([
        'KUBECONFIG={}'.format(KUBECONFIG_PATH),
        'kubectl', 'get', 'nodes', '-o', 'json'
    ])

    formatted = json.loads(nodes_raw_json.output)

    try:
        for node in formatted['items']:
            if node['status']['nodeInfo']['machineID'] == machine_id:
                return node['metadata']['name']
    except KeyError as e:
        raise Exception('Unexpected format for node name: {}'.format(e))

    raise Exception('Node name could not be determined via machine-id')


def annotate(resource, resource_name, key, value):
    """
    Annotates any kubernetes resource
    """

    ret = run_command([
        'KUBECONFIG={}'.format(KUBECONFIG_PATH),
        'kubectl', 'annotate', resource, resource_name,
        '{}={}'.format(key, value)
    ])

    return ret.output


if __name__ == "__main__":  # pragma: no cover
    main()
