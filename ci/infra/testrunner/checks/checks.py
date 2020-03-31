import time

import platforms
from utils.utils import Utils

_checks = {}

def check(name=None, roles=[], check_timeout=300, check_backoff=20):
    """Decorator for waiting a check to become true.
       Can receve the following arguments when invoking the check function
       name: name of the check (used for reporting, if not defined, check
             function is used
       roles: list of node roles this check applies
       check_timeout: the timeout for the check
       check_backoff: the backoff between retries

      The check_timout and check_backoff parameters can be overidden when
      calling the check_node function
    """
    def checker(check):
        def wait_condition(*args, **kwargs):
            _name  = name
            if not _name:
               _name = check.__name__

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
                    msg = (f'condition {_name} not satisfied after {timeout} seconds'
                           f'{". Last error:"+str(last_error) if last_error else ""}')
                    raise AssertionError(msg)

                time.sleep(backoff)

        for role in roles:
           role_checks = _checks.get(role, [])
           role_checks.append(wait_condition)
           _checks[role] = role_checks

        return wait_condition

    return checker


class Checker:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.utils.setup_ssh()
        self.platform = platform


    def check_node(self, role, node, timeout=180, backoff=20):
        start   = int(time.time())
        for check in _checks.get(role, []):
            remaining = timeout-(int(time.time())-start)
            check(self.conf, self.platform, role, node, check_timeout=remaining, check_backoff=backoff)


@check(name="apiserver healthz", roles=['master'])
def check_apiserver_healthz(conf, platform, role, node):
     platform = platforms.get_platform(conf, platform)
     cmd =   'curl -Ls --insecure https://localhost:6443/healthz'
     output = platform.ssh_run(role, node, cmd)
     return output.find("ok") > -1

@check(name="etcd health", roles=['master'])
def check_etcd_health(conf, platform, role, node):
    platform = platforms.get_platform(conf, platform)
    cmd = ('sudo curl -Ls --cacert /etc/kubernetes/pki/etcd/ca.crt '
           '--key /etc/kubernetes/pki/etcd/server.key '
           '--cert /etc/kubernetes/pki/etcd/server.crt '
           'https://localhost:2379/health')
    output = platform.ssh_run(role, node, cmd)
    return output.find("true") > -1
