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

import json
from mock import patch, call, mock_open, Mock, ANY

from skuba_update.skuba_update import (
    main,
    update,
    run_command,
    run_zypper_command,
    run_zypper_patch,
    node_name_from_machine_id,
    annotate,
    is_reboot_needed,
    annotate_updates_available,
    get_update_list,
    REBOOT_REQUIRED_PATH,
    ZYPPER_EXIT_INF_UPDATE_NEEDED,
    ZYPPER_EXIT_INF_RESTART_NEEDED,
    ZYPPER_EXIT_INF_REBOOT_NEEDED,
    KUBE_UPDATES_KEY,
    KUBE_SECURITY_UPDATES_KEY,
    KUBE_DISRUPTIVE_UPDATES_KEY
)


@patch('subprocess.Popen')
def test_run_command(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'stdout', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    result = run_command(['/bin/dummycmd', 'arg1'])
    assert result.output == "stdout"
    assert result.returncode == 0
    assert result.error == 'stderr'

    mock_process.returncode = 1
    result = run_command(['/bin/dummycmd', 'arg1'])
    assert result.output == "stdout"
    assert result.returncode == 1

    mock_process.communicate.return_value = (b'', b'stderr')
    result = run_command(['/bin/dummycmd', 'arg1'])
    assert result.output == ""
    assert result.returncode == 1


@patch('argparse.ArgumentParser.parse_args')
@patch('subprocess.Popen')
def test_main_wrong_version(mock_subprocess, mock_args):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper 1.13.0', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        main()
    except Exception as e:
        exception = True
        assert 'higher is required' in str(e)
    assert exception


@patch('argparse.ArgumentParser.parse_args')
@patch('subprocess.Popen')
def test_main_bad_format_version(mock_subprocess, mock_args):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        main()
    except Exception as e:
        exception = True
        assert 'Could not parse' in str(e)
    assert exception


@patch('argparse.ArgumentParser.parse_args')
@patch('subprocess.Popen')
def test_main_no_root(mock_subprocess, mock_args):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper 1.14.15', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        main()
    except Exception as e:
        exception = True
        assert 'root privileges' in str(e)
    assert exception


@patch('skuba_update.skuba_update.annotate_updates_available')
@patch('argparse.ArgumentParser.parse_args')
@patch('os.environ.get', new={}.get, spec_set=True)
@patch('os.geteuid')
@patch('subprocess.Popen')
def test_main(mock_subprocess, mock_geteuid, mock_args, mock_annotate):
    return_values = [
        (b'some_service1\nsome_service2', b''),
        (b'zypper 1.14.15', b'')
    ]

    def mock_communicate():
        if len(return_values) > 1:
            return return_values.pop()
        else:
            return return_values[0]

    args = Mock()
    args.annotate_only = False
    mock_args.return_value = args
    mock_geteuid.return_value = 0
    mock_process = Mock()
    mock_process.communicate.side_effect = mock_communicate
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    main()
    assert mock_subprocess.call_args_list == [
        call(['zypper', '--version'], stdout=-1, stderr=-1, env=ANY),
        call(['zypper', 'ref', '-s'], stdout=None, stderr=None, env=ANY),
        call([
            'zypper', '--non-interactive',
            '--non-interactive-include-reboot-patches', 'patch'
        ], stdout=None, stderr=None, env=ANY),
        call(['zypper', 'needs-rebooting'], stdout=None, stderr=None, env=ANY),
        call(
            ['zypper', 'ps', '-sss'],
            stdout=-1, stderr=-1, env=ANY
        ),
        call(
            ['systemctl', 'restart', 'some_service1'],
            stdout=None, stderr=None, env=ANY
        ),
        call(
            ['systemctl', 'restart', 'some_service2'],
            stdout=None, stderr=None, env=ANY
        ),
    ]


@patch('skuba_update.skuba_update.annotate_updates_available')
@patch('argparse.ArgumentParser.parse_args')
@patch('os.environ.get', new={}.get, spec_set=True)
@patch('os.geteuid')
@patch('subprocess.Popen')
def test_main_annotate_only(
    mock_subprocess, mock_geteuid, mock_args, mock_annotate
):
    args = Mock()
    args.annotate_only = True
    mock_args.return_value = args
    mock_geteuid.return_value = 0
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper 1.14.15', b'stderr')
    mock_process.returncode = ZYPPER_EXIT_INF_UPDATE_NEEDED
    mock_subprocess.return_value = mock_process
    main()
    assert mock_subprocess.call_args_list == [
        call(['zypper', '--version'], stdout=-1, stderr=-1, env=ANY),
        call(['zypper', 'ref', '-s'], stdout=None, stderr=None, env=ANY),
    ]


@patch('skuba_update.skuba_update.annotate_updates_available')
@patch('argparse.ArgumentParser.parse_args')
@patch('os.environ.get', new={}.get, spec_set=True)
@patch('os.geteuid')
@patch('subprocess.Popen')
def test_main_zypper_returns_100(
    mock_subprocess, mock_geteuid, mock_args, mock_annotate
):
    return_values = [(b'', b''), (b'zypper 1.14.15', b'')]

    def mock_communicate():
        if len(return_values) > 1:
            return return_values.pop()
        else:
            return return_values[0]

    args = Mock()
    args.annotate_only = False
    mock_args.return_value = args
    mock_geteuid.return_value = 0
    mock_process = Mock()
    mock_process.communicate.side_effect = mock_communicate
    mock_process.returncode = ZYPPER_EXIT_INF_RESTART_NEEDED
    mock_subprocess.return_value = mock_process
    main()
    assert mock_subprocess.call_args_list == [
        call(['zypper', '--version'], stdout=-1, stderr=-1, env=ANY),
        call(['zypper', 'ref', '-s'], stdout=None, stderr=None, env=ANY),
        call([
            'zypper', '--non-interactive',
            '--non-interactive-include-reboot-patches', 'patch'
        ], stdout=None, stderr=None, env=ANY),
        call([
            'zypper', 'needs-rebooting'
        ], stdout=None, stderr=None, env=ANY),
        call([
            'zypper', '--non-interactive',
            '--non-interactive-include-reboot-patches', 'patch'
        ], stdout=None, stderr=None, env=ANY),
        call([
            'zypper', 'needs-rebooting'
        ], stdout=None, stderr=None, env=ANY),
        call(
            ['zypper', 'ps', '-sss'],
            stdout=-1, stderr=-1, env=ANY
        ),
    ]


@patch('pathlib.Path.is_file')
@patch('subprocess.Popen')
def test_update_zypper_is_fine_but_created_needreboot(
        mock_subprocess, mock_is_file
):

    mock_process = Mock()
    mock_process.communicate.return_value = (b'stdout', b'stderr')

    mock_process.returncode = ZYPPER_EXIT_INF_REBOOT_NEEDED
    mock_subprocess.return_value = mock_process
    mock_is_file.return_value = True

    exception = False
    try:
        update()
    except PermissionError as e:
        exception = True
        msg = 'Permission denied: \'{0}\''.format(REBOOT_REQUIRED_PATH)
        assert msg in str(e)
    assert exception


@patch('subprocess.Popen')
def test_run_zypper_command(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'stdout', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    assert run_zypper_command(['zypper', 'patch']) == 0
    mock_process.returncode = ZYPPER_EXIT_INF_RESTART_NEEDED
    mock_subprocess.return_value = mock_process
    assert run_zypper_command(
        ['zypper', 'patch']) == ZYPPER_EXIT_INF_RESTART_NEEDED


@patch('subprocess.Popen')
def test_run_zypper_command_failure(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'', b'')
    mock_process.returncode = 1
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        run_zypper_command(['zypper', 'patch']) == 'stdout'
    except Exception as e:
        exception = True
        assert '"zypper patch" failed' in str(e)
    assert exception


@patch('subprocess.Popen')
def test_run_zypper_command_creates_file_on_102(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'', b'')
    mock_process.returncode = ZYPPER_EXIT_INF_REBOOT_NEEDED
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        run_zypper_patch() == 'stdout'
    except PermissionError as e:
        exception = True
        msg = 'Permission denied: \'{0}\''.format(REBOOT_REQUIRED_PATH)
        assert msg in str(e)
    assert exception


@patch('builtins.open',
       mock_open(read_data='9ea12911449eb7b5f8f228294bf9209a'))
@patch('subprocess.Popen')
@patch('json.loads')
def test_node_name_from_machine_id(mock_loads, mock_subprocess):
    json_node_object = {
        'items': [
            {
                'metadata': {
                    'name': 'my-node-1'
                },
                'status': {
                    'nodeInfo': {
                        'machineID': '49f8e2911a1449b7b5ef2bf92282909a'
                    }
                }
            },
            {
                'metadata': {
                    'name': 'my-node-2'
                },
                'status': {
                    'nodeInfo': {
                        'machineID': '9ea12911449eb7b5f8f228294bf9209a'
                    }
                }
            }
        ]
    }
    breaking_json_node_object = {'Items': []}

    mock_process = Mock()
    mock_process.communicate.return_value = (json.dumps(json_node_object)
                                             .encode(), b'')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    mock_loads.return_value = json_node_object
    assert node_name_from_machine_id() == 'my-node-2'

    json_node_object2 = json_node_object
    json_node_object2['items'][1]['status']['nodeInfo']['machineID'] = \
        'another-id-that-doesnt-reflect-a-node'
    mock_loads.return_value = json_node_object2
    exception = False
    try:
        node_name_from_machine_id() == 'my-node-2'
    except Exception as e:
        exception = True
        assert 'Node name could not be determined' in str(e)
    assert exception

    mock_loads.return_value = breaking_json_node_object
    exception = False
    try:
        node_name_from_machine_id() == 'my-node-2'
    except Exception as e:
        exception = True
        assert 'Unexpected format' in str(e)
    assert exception


@patch('subprocess.Popen')
def test_annotate(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'node/my-node-1 annotated',
                                             b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    assert annotate(
        'node', 'my-node-1',
        KUBE_DISRUPTIVE_UPDATES_KEY, 'yes'
    ) == 'node/my-node-1 annotated'


@patch('skuba_update.skuba_update.node_name_from_machine_id')
@patch('skuba_update.skuba_update.annotate')
@patch('subprocess.Popen')
def test_annotate_updates_empty(mock_subprocess, mock_annotate, mock_name):
    mock_name.return_value = 'mynode'
    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<stream><update-status><update-list>'
        b'</update-list></update-status></stream>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    annotate_updates_available()
    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        )
    ]
    assert mock_annotate.call_args_list == [
        call('node', 'mynode', KUBE_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_SECURITY_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_DISRUPTIVE_UPDATES_KEY, 'no')
    ]


@patch('skuba_update.skuba_update.node_name_from_machine_id')
@patch('skuba_update.skuba_update.annotate')
@patch('subprocess.Popen')
def test_annotate_updates(mock_subprocess, mock_annotate, mock_name):
    mock_name.return_value = 'mynode'
    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<stream><update-status><update-list><update interactive="message">'
        b'</update></update-list></update-status></stream>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    annotate_updates_available()
    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        )
    ]
    assert mock_annotate.call_args_list == [
        call('node', 'mynode', KUBE_UPDATES_KEY, 'yes'),
        call('node', 'mynode', KUBE_SECURITY_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_DISRUPTIVE_UPDATES_KEY, 'yes')
    ]


@patch("skuba_update.skuba_update.node_name_from_machine_id")
@patch("builtins.open", read_data="aa59dc0c5fe84247a77c26780dd0b3fd")
@patch('subprocess.Popen')
def test_annotate_updates_available(mock_subprocess, mock_open, mock_name):
    mock_name.return_value = 'mynode'

    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<stream><update-status><update-list><update interactive="message">'
        b'</update></update-list></update-status></stream>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process

    annotate_updates_available()

    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        ),
        call(
            ["kubectl", "annotate", "--overwrite", "node",
                "mynode", "caasp.suse.com/has-updates=yes"],
            stdout=-1, stderr=-1, env=ANY
        ),
        call(
            ["kubectl", "annotate", "--overwrite", "node",
                "mynode", "caasp.suse.com/has-security-updates=no"],
            stdout=-1, stderr=-1, env=ANY
        ),
        call(
            ["kubectl", "annotate", "--overwrite", "node",
                "mynode", "caasp.suse.com/has-disruptive-updates=yes"],
            stdout=-1, stderr=-1, env=ANY
        )
    ]


@patch('skuba_update.skuba_update.node_name_from_machine_id')
@patch('skuba_update.skuba_update.annotate')
@patch('subprocess.Popen')
def test_annotate_updates_bad_xml(mock_subprocess, mock_annotate, mock_name):
    mock_name.return_value = 'mynode'
    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<update-status><update-list><update interactive="message">'
        b'</update></update-list></update-status>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process

    annotate_updates_available()
    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        )
    ]
    assert mock_annotate.call_args_list == [
        call('node', 'mynode', KUBE_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_SECURITY_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_DISRUPTIVE_UPDATES_KEY, 'no')
    ]


@patch('skuba_update.skuba_update.node_name_from_machine_id')
@patch('skuba_update.skuba_update.annotate')
@patch('subprocess.Popen')
def test_annotate_updates_security(
    mock_subprocess, mock_annotate, mock_name
):
    mock_name.return_value = 'mynode'
    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<stream><update-status><update-list>'
        b'<update interactive="false" category="security">'
        b'</update></update-list></update-status></stream>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process

    annotate_updates_available()
    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        )
    ]
    assert mock_annotate.call_args_list == [
        call('node', 'mynode', KUBE_UPDATES_KEY, 'yes'),
        call('node', 'mynode', KUBE_SECURITY_UPDATES_KEY, 'yes'),
        call('node', 'mynode', KUBE_DISRUPTIVE_UPDATES_KEY, 'no')
    ]


@patch('skuba_update.skuba_update.node_name_from_machine_id')
@patch('skuba_update.skuba_update.annotate')
@patch('subprocess.Popen')
def test_annotate_updates_available_is_reboot(
    mock_subprocess, mock_annotate, mock_name
):
    mock_name.return_value = 'mynode'

    mock_process = Mock()
    mock_process.communicate.return_value = (
        b'<stream><update-status><update-list><update interactive="reboot">'
        b'</update></update-list></update-status></stream>', b''
    )
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process

    annotate_updates_available()
    assert mock_subprocess.call_args_list == [
        call(
            ['zypper', '--non-interactive', '--xmlout', 'list-patches'],
            stdout=-1, stderr=-1, env=ANY
        )
    ]
    assert mock_annotate.call_args_list == [
        call('node', 'mynode', KUBE_UPDATES_KEY, 'yes'),
        call('node', 'mynode', KUBE_SECURITY_UPDATES_KEY, 'no'),
        call('node', 'mynode', KUBE_DISRUPTIVE_UPDATES_KEY, 'yes')
    ]


@patch('subprocess.Popen')
def test_is_reboot_needed_truthy(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'', b'')
    mock_process.returncode = ZYPPER_EXIT_INF_REBOOT_NEEDED
    mock_subprocess.return_value = mock_process

    assert is_reboot_needed()


@patch('subprocess.Popen')
def test_is_reboot_needed_falsey(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'', b'')
    mock_process.returncode = ZYPPER_EXIT_INF_RESTART_NEEDED
    mock_subprocess.return_value = mock_process

    assert not is_reboot_needed()


def test_get_update_list_bad_xml():
    assert get_update_list('<xml') is None
