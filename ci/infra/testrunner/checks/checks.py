import time

import platforms
from utils.utils import Utils

_checks_by_role = {}
_checks_by_name = {}

def check(description=None, roles=[], check_timeout=300, check_backoff=20):
    """Decorator for waiting a check to become true.
       Can receve the following arguments when invoking the check function
       description: used for reporting. if not defined, check
                    function name is used
       roles: list of node roles this check applies
       check_timeout: the timeout for the check
       check_backoff: the backoff between retries

      The check_timout and check_backoff parameters can be overidden when
      calling the check_node function
    """
    def checker(check):
        def wait_condition(*args, **kwargs):
            _name = check.__name__
            _description = description
            if not _description:
                _description = _name

            timeout = kwargs.pop('check_timeout', check_timeout)
            backoff = kwargs.pop('check_backoff', check_backoff)
            deadline = int(time.time()) + timeout
            while True:
                last_error = None
                try:
                    if check(*args, **kwargs):
                        return True
                except Exception as ex:
                    last_error = ex

                if int(time.time()) >= deadline:
                    msg = (f'condition "{_description}" not satisfied after {timeout} seconds'
                           f'{". Last error:"+str(last_error) if last_error else ""}')
                    raise AssertionError(msg)

                time.sleep(backoff)

        _checks_by_name[check.__name__] = wait_condition
        for role in roles:
           role_checks = _checks_by_role.get(role, [])
           role_checks.append(wait_condition)
           _checks_by_role[role] = role_checks

        return wait_condition

    return checker


class Checker:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.utils.setup_ssh()
        self.platform = platform


    def check_node(self, role, node, checks=None, timeout=180, backoff=20):
        if checks:
            check_names = checks
            checks = []
            for name in check_names:
                checks.append(_checks_by_name[name])
        else:
            checks = _checks_by_role.get(role, [])

        start   = int(time.time())
        for check in checks:
            remaining = timeout-(int(time.time())-start)
            check(self.conf, self.platform, role, node, check_timeout=remaining, check_backoff=backoff)


@check(description="apiserver healthz check", roles=['master'])
def check_apiserver_healthz(conf, platform, role, node):
     platform = platforms.get_platform(conf, platform)
     cmd =   'curl -Ls --insecure https://localhost:6443/healthz'
     output = platform.ssh_run(role, node, cmd)
     return output.find("ok") > -1

@check(description="etcd health check", roles=['master'])
def check_etcd_health(conf, platform, role, node):
    platform = platforms.get_platform(conf, platform)
    cmd = ('sudo curl -Ls --cacert /etc/kubernetes/pki/etcd/ca.crt '
           '--key /etc/kubernetes/pki/etcd/server.key '
           '--cert /etc/kubernetes/pki/etcd/server.crt '
           'https://localhost:2379/health')
    output = platform.ssh_run(role, node, cmd)
    return output.find("true") > -1
