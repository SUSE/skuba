import time

import platforms
from kubectl import Kubectl
from utils.utils import Utils


class Check():
    def __init__(self, name, description, func, scope, roles=[], stages=[]):
        self.name = name
        self.description = description
        self.func = func
        self.scope = scope
        self.roles = roles
        self.stages = stages


_checks = []
_checks_by_name = {}


def check(description=None, scope=None, roles=[], stages=[], check_timeout=300, check_backoff=20):
    """Decorator for waiting a check to become true.
       Can receve the following arguments when invoking the check function
       description: used for reporting. if not defined, check
                    function name is used
       scope: either "cluster" or "node"
       roles: list of node roles this check applies
       stages: list of deployment stages this check applies to (e.g provisioned, joined)
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

        if scope is None:
            raise ValueError("scope must be defined: 'cluster' or 'node'")

        _check = Check(
            check.__name__,
            description,
            wait_condition,
            scope,
            roles=roles,
            stages=stages)
        _checks.append(_check)
        _checks_by_name[_check.name] = _check

        return wait_condition

    return checker


class Checker:

    def __init__(self, conf, platform):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.utils.setup_ssh()
        self.platform = platform

    def _filter_checks(self, checks, scope=None, stage=None, role=None):
        _filtered = checks
        if scope:
            _filtered = [c for c in _filtered if scope == c.scope]
        if stage:
            _filtered = [c for c in _filtered if stage in c.stages]
        if role:
            _filtered = [c for c in _filtered if role in c.roles]

        return _filtered

    def _filter_by_name(self, names):
        checks = []
        for name in names:
            _check = _checks_by_name.get(name, None)
            if _check is None:
                raise ValueError("Check {name} not found")
            checks.append(_check)

        return checks

    def check_node(self, role, node, checks=None, stage=None, timeout=180, backoff=20):

        # Prevent defaults to be accidentally overridden by callers with None
        if timeout is None:
            timeout = 180
        if backoff is None:
            backoff = 20

        if checks:
            checks = self._filter_by_name(checks)
            for check in checks:
                if check.scope != "node":
                    raise Exception(f'check {check.name} is not a node check')
        else:
            if not stage:
                raise ValueError("stage must be specified")
            checks = self._filter_checks(_checks, stage=stage, scope="node", role=role)

        start = int(time.time())
        for check in checks:
            remaining = timeout - (int(time.time()) - start)
            check.func(self.conf, self.platform, role, node, check_timeout=remaining, check_backoff=backoff)

    def check_cluster(self, checks=None, stage=None, timeout=180, backoff=20):
        if checks:
            checks = self._filter_by_name(checks)
            for check in checks:
                if check.scope != "cluster":
                    raise Exception(f'check {check.name} is not a cluster check')
        else:
            if not stage:
                raise ValueError("stage must be specified")
            checks = self._filter_checks(_checks, stage=stage, scope="cluster")

        start = int(time.time())
        for check in checks:
            remaining = timeout - (int(time.time()) - start)
            check.func(self.conf, self.platform, check_timeout=remaining, check_backoff=backoff)


@check(description="apiserver healthz check", scope="node", roles=['master'])
def check_apiserver_healthz(conf, platform, role, node):
    platform = platforms.get_platform(conf, platform)
    cmd = 'curl -Ls --insecure https://localhost:6443/healthz'
    output = platform.ssh_run(role, node, cmd)
    return output.find("ok") > -1


@check(description="etcd health check", scope="node", roles=['master'], stages=["joined"])
def check_etcd_health(conf, platform, role, node):
    platform = platforms.get_platform(conf, platform)
    cmd = ('sudo curl -Ls --cacert /etc/kubernetes/pki/etcd/ca.crt '
           '--key /etc/kubernetes/pki/etcd/server.key '
           '--cert /etc/kubernetes/pki/etcd/server.crt '
           'https://localhost:2379/health')
    output = platform.ssh_run(role, node, cmd)
    return output.find("true") > -1


@check(description="check node is ready", scope="node", roles=["master", "worker"], stages=["joined"])
def check_node_ready(conf, platform, role, node):
    platform = platforms.get_platform(conf, platform)
    node_name = platform.get_nodes_names(role)[node]
    cmd = ("get nodes {} -o jsonpath='{{range @.status.conditions[*]}}"
           "{{@.type}}={{@.status}};{{end}}'").format(node_name)
    kubectl = Kubectl(conf)
    return kubectl.run_kubectl(cmd).find("Ready=True") != -1


@check(description="check system pods ready", scope="cluster", stages=["joined"])
def check_system_pods_ready(conf, platform):
    kubectl = Kubectl(conf)
    return check_pods_ready(kubectl, namespace="kube-system")


def check_pods_ready(kubectl, namespace=None, pods=[], node=None, statuses=['Running', 'Succeeded']):
    ns = f'{"--namespace="+namespace if namespace else ""}'
    node_selector = f'{"--field-selector spec.nodeName="+node if node else ""}'
    cmd = (f'get pods {" ".join(pods)} {ns} {node_selector} '
           f'-o jsonpath="{{ range .items[*]}}{{@.metadata.name}}:'
           f'{{@.status.phase}};"')

    result = kubectl.run_kubectl(cmd)
    # get pods can return a list of items or a single pod
    pod_list = result.split(";")
    for name,status in [ pod.split(":") for pod in pod_list if pod is not ""]:
        if status not in statuses:
            return False

    return True
