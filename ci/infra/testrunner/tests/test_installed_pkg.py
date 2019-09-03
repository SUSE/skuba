import logging

logger = logging.getLogger("testrunner")


def test_installed_pkg(deployment, platform, skuba, kubectl):
    assert skuba.num_of_nodes('master') > 0

    logger.info('checking for installed packages on control plane')
    output = platform.ssh_run('master', 0, 'rpm -q nfs-client')
    assert 'not installed' not in output
    output = platform.ssh_run('master', 0, 'rpm -q xfsprogs')
    assert 'not installed' not in output
    output = platform.ssh_run('master', 0, 'rpm -q ceph-common')
    assert 'not installed' not in output

    logger.info('checking for installed packages on worker')
    output = platform.ssh_run('worker', 0, 'rpm -q nfs-client')
    assert 'not installed' not in output
    output = platform.ssh_run('worker', 0, 'rpm -q xfsprogs')
    assert 'not installed' not in output
    output = platform.ssh_run('worker', 0, 'rpm -q ceph-common')
    assert 'not installed' not in output
