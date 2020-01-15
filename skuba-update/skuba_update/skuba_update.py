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

import argparse
import json
import os
import re
import subprocess
from collections import namedtuple
from datetime import datetime
from pathlib import Path
from xml.etree import ElementTree

import pkg_resources

# Since zypper 1.14.0, it will automatically create a `/var/run/reboot-needed`
# text file whenever one of the applied patches requires the system to be
# rebooted.
# Since zypper 1.14.15, there is a subcommand to check if a reboot is needed
# `zypper needs-rebooting`
REQUIRED_ZYPPER_VERSION = (1, 14, 15)

# The path where zypper might write if it has detected that a patch/update that
# has been installed requires the machine to reboot in order to work properly.
ZYPPER_REBOOT_NEEDED_PATH = '/var/run/reboot-needed'

# The path to the reboot-required file. This is the file that kured will be
# looking at.
REBOOT_REQUIRED_PATH = '/var/run/reboot-required'

# Exit codes as defined by zypper.

ZYPPER_EXIT_INF_UPDATE_NEEDED = 100
ZYPPER_EXIT_INF_SEC_UPDATE_NEEDED = 101
ZYPPER_EXIT_INF_REBOOT_NEEDED = 102
ZYPPER_EXIT_INF_RESTART_NEEDED = 103

# The path to the kubelet config used for running kubectl commands
KUBECONFIG_PATH = '/etc/kubernetes/kubelet.conf'

# Updates annotation keys on the Kubernetes node.
KUBE_UPDATES_KEY = 'caasp.suse.com/has-updates'
KUBE_SECURITY_UPDATES_KEY = 'caasp.suse.com/has-security-updates'
KUBE_DISRUPTIVE_UPDATES_KEY = 'caasp.suse.com/has-disruptive-updates'
KUBE_CAASP_RELEASE_VERSION_KEY = 'caasp.suse.com/caasp-release-version'


def main():
    """
    main-entry point for program.
    """

    args = parse_args()

    # Check that we have the proper zypper version.
    if not check_version('zypper', REQUIRED_ZYPPER_VERSION):
        raise Exception('zypper version {0} or higher is required'.format(
            '.'.join([str(x) for x in REQUIRED_ZYPPER_VERSION])
        ))

    if os.geteuid() != 0:
        raise Exception('root privileges are required to run this tool')

    run_zypper_command(['ref', '-s'])
    if not args.annotate_only:
        code = update()
        restart_services()
        annotate_node()
        reboot_sentinel_file(code)
    else:
        annotate_node()


def parse_args():
    """
    Returns the parsed arguments.
    """

    annotate_only_msg = \
        'Do not install any update, just annotate there are available updates'
    version_msg = '%(prog)s {0}'.format(version())

    parser = argparse.ArgumentParser(description='Updates a CaaSP node')
    parser.add_argument(
        '--annotate-only', action='store_true', help=annotate_only_msg
    )
    parser.add_argument(
        '--version', action='version', version=version_msg
    )

    return parser.parse_args()


def version():
    """
    Returns the version of the current skuba-update
    """

    return pkg_resources.require('skuba-update')[0].version


def update():
    """
    Performs an update operation.
    """

    code = run_zypper_patch()
    if is_restart_needed(code):
        code = run_zypper_patch()
    return code


def annotate_node():
    node_name = node_name_from_machine_id()
    annotate_updates_available(node_name)
    annotate_caasp_release_version(node_name)


def annotate_updates_available(node_name):
    """
    Performs a zypper list-patches and annotates the node like so:

      1. If there is at least one update of any kind `has_updates` flag is set.
      2. If there is at least one security update `has_security_updates` flag
         is set.
      3. If there is at least one disruptive update `has_disruptive_updates`
         flag is set.
    """

    patch_xml = run_zypper_command(
        ['--non-interactive', '--xmlout', 'list-patches'],
        needsOutput=True
    ).output
    updates = get_update_list(patch_xml)
    annotate(
        'node', node_name, KUBE_UPDATES_KEY,
        'yes' if has_updates(updates) else 'no'
    )
    annotate(
        'node', node_name, KUBE_SECURITY_UPDATES_KEY,
        'yes' if has_security_updates(updates) else 'no'
    )
    annotate(
        'node', node_name, KUBE_DISRUPTIVE_UPDATES_KEY,
        'yes' if has_disruptive_updates(updates) else 'no'
    )


def annotate_caasp_release_version(node_name):
    """
    Performs fetch caasp-release version and annotates to the node.
    """

    cmd = run_command(['rpm', '-q', 'caasp-release',
                       '--queryformat', '%{VERSION}'])
    if cmd.returncode != 0 or not cmd.output:
        log('Failed get caasp-release rpm package version')
        return

    annotate(
        'node', node_name, KUBE_CAASP_RELEASE_VERSION_KEY,
        cmd.output,
    )


def get_update_list(patch_xml):
    """
    Fetch the update list from the given XML output.
    """

    try:
        tree = ElementTree.fromstring(patch_xml)
    except ElementTree.ParseError:
        return None

    us = tree.find('update-status')
    if us is None:
        return None
    return us.find('update-list')


def has_updates(update_list):
    """
    Returns true if there are updates available.
    """
    if update_list is None or len(update_list) == 0:
        return False
    else:
        return True


def has_security_updates(update_list):
    """
    Returns true if there are security updates available.
    """

    return filter_updates(update_list, 'category', lambda x: x == 'security')


def has_disruptive_updates(update_list):
    """
    Returns true if there are disruptive updates available.
    """

    return filter_updates(update_list, 'interactive', is_not_false_str)


def filter_updates(update_list, attrib, attrib_check):
    """
    Returns True if there is at least one update having the given
    attribute (attrib) that is also passing the checker
    function (attrib_check)
    """

    if has_updates(update_list):
        for update in update_list:
            attr = update.attrib.get(attrib, '')
            if attrib_check(attr):
                return True

    return False


def restart_services():
    """
    Restart services which are reported to have been updated and need a
    restart.
    """

    result = run_zypper_command(['ps', '-sss'], needsOutput=True)
    for service in result.output.splitlines():
        cmd = run_command(['systemctl', 'restart', service], needsOutput=False)
        if cmd.returncode != 0:
            log('Warning! Service \'{0}\' restart returned non zero exit code')


def is_zypper_error(code):
    """
    Returns true if the given code belongs to an error code according
    to zypper.
    """

    return code != 0 and code < ZYPPER_EXIT_INF_UPDATE_NEEDED


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

    return run_zypper_command(
        ['needs-rebooting']
    ) == ZYPPER_EXIT_INF_REBOOT_NEEDED


def reboot_sentinel_file(code):
    # There are two instances in which we should create the
    # REBOOT_REQUIRED_PATH file:
    #
    # 1. Zypper returned an exit code telling us to restart the system.
    # 2. `zypper needs-rebooting` returns ZYPPER_EXIT_INF_REBOOT_NEEDED.
    if code == ZYPPER_EXIT_INF_REBOOT_NEEDED or is_reboot_needed():
        Path(REBOOT_REQUIRED_PATH).touch()
    return code


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
    zypperCommand = ['zypper', '--userdata', 'skuba-update', ] + command

    process = run_command(zypperCommand, needsOutput)
    if is_zypper_error(process.returncode):
        raise Exception('"{0}" failed'.format(' '.join(zypperCommand)))
    if needsOutput:
        return process
    return process.returncode


def run_zypper_patch():
    return run_zypper_command([
        '--non-interactive', '--non-interactive-include-reboot-patches',
        'patch'
    ])


def run_command(command, needsOutput=True, added_env={}):
    """
    Runs the given command and it returns a named tuple containing: 'output',
    'error' and 'returncode'. It also accepts a dictionary `added_env`, which
    will be added to the environment for the given command.
    """

    env = os.environ.copy()
    if len(added_env) > 0:
        env.update(added_env)

    command_type = namedtuple(
        'command', ['output', 'error', 'returncode']
    )
    log('running \'{0}\''.format(' '.join(command)))
    process = subprocess.Popen(
        command,
        stdout=subprocess.PIPE if needsOutput else None,
        stderr=subprocess.PIPE if needsOutput else None,
        env=env
    )
    output, error = process.communicate()
    return command_type(
        output=output.decode() if needsOutput else None,
        error=error.decode() if needsOutput else None,
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

    nodes_raw_json = run_command(
        ['kubectl', 'get', 'nodes', '-o', 'json'],
        added_env={'KUBECONFIG': KUBECONFIG_PATH}
    )

    if nodes_raw_json.returncode != 0 or not nodes_raw_json.output:
        raise Exception('Kubectl failed getting nodes list')

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
        'kubectl', 'annotate', '--overwrite', resource, resource_name,
        '{}={}'.format(key, value)],
        added_env={'KUBECONFIG': KUBECONFIG_PATH}
    )

    if ret.returncode != 0:
        log('Warning! kubectl returned non zero exit code')

    return ret.output


if __name__ == "__main__":  # pragma: no cover
    main()
