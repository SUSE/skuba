from mock import patch, call, Mock, ANY

from skuba_update.skuba_update import (
    main,
    run_command,
    run_zypper_command
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

    mock_process.returncode = 1
    result = run_command(['/bin/dummycmd', 'arg1'])
    assert result.output == "stdout"
    assert result.returncode == 1

    mock_process.communicate.return_value = (b'', b'stderr')
    result = run_command(['/bin/dummycmd', 'arg1'])
    assert result.output == "(no output on stdout)"
    assert result.returncode == 1


@patch('subprocess.Popen')
def test_main_wrong_version(mock_subprocess):
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


@patch('subprocess.Popen')
def test_main_bad_format_version(mock_subprocess):
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


@patch('subprocess.Popen')
def test_main_no_root(mock_subprocess):
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


@patch('os.environ.get', new={}.get, spec_set=True)
@patch('os.geteuid')
@patch('subprocess.Popen')
def test_main(mock_subprocess, mock_geteuid):
    mock_geteuid.return_value = 0
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper 1.14.15', b'')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    main()
    assert mock_subprocess.call_args_list == [
        call(['zypper', '--version'], stdout=ANY, stderr=ANY, env=ANY),
        call(['zypper', 'ref', '-s'], env=ANY),
        call([
            'zypper', '--non-interactive-include-reboot-patches', 'patch'
        ], env=ANY),
        call([
            'zypper', '--non-interactive-include-reboot-patches', 'patch-check'
        ], env=ANY)
    ]


@patch('os.environ.get', new={}.get, spec_set=True)
@patch('os.geteuid')
@patch('subprocess.Popen')
def test_main_zypper_returns_100(mock_subprocess, mock_geteuid):
    mock_geteuid.return_value = 0
    mock_process = Mock()
    mock_process.communicate.return_value = (b'zypper 1.14.15', b'')
    mock_process.returncode = 100
    mock_subprocess.return_value = mock_process
    main()
    assert mock_subprocess.call_args_list == [
        call(['zypper', '--version'], stdout=ANY, stderr=ANY, env=ANY),
        call(['zypper', 'ref', '-s'], env=ANY),
        call([
            'zypper', '--non-interactive-include-reboot-patches', 'patch'
        ], env=ANY),
        call([
            'zypper', '--non-interactive-include-reboot-patches', 'patch-check'
        ], env=ANY),
        call([
            'zypper', '--non-interactive-include-reboot-patches', 'patch'
        ], env=ANY)
    ]


@patch('subprocess.Popen')
def test_run_zypper_command(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (b'stdout', b'stderr')
    mock_process.returncode = 0
    mock_subprocess.return_value = mock_process
    assert run_zypper_command(['zypper', 'patch']) == 0
    mock_process.returncode = 100
    mock_subprocess.return_value = mock_process
    assert run_zypper_command(['zypper', 'patch']) == 100


@patch('subprocess.Popen')
def test_run_zypper_command_failure(mock_subprocess):
    mock_process = Mock()
    mock_process.communicate.return_value = (None, None)
    mock_process.returncode = 1
    mock_subprocess.return_value = mock_process
    exception = False
    try:
        run_zypper_command(['zypper', 'patch']) == 'stdout'
    except Exception as e:
        exception = True
        assert '"zypper patch" failed' in str(e)
    assert exception
