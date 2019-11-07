import json
import signal
import time


def check_nodes_ready(kubectl):
    # Retrieve the node name and the Ready condition (True, or False)
    cmd = ("get nodes -o jsonpath='{ range .items[*]}{@.metadata.name}{\":\"}"
           "{range @.status.conditions[?(.type==\"Ready\")]}{@.status}{\" \"}'")

    nodes = kubectl.run_kubectl(cmd).strip().split(" ")
    for node in nodes:
        node_name, node_status = node.split(":")
        assert node_status == "True", f'Node {node_name} is not Ready'

def check_pods_ready(kubectl, namespace=None, pods=[], statuses=['Running', 'Succeeded']):
    
    ns = f'{"--namespace="+namespace if namespace else ""}' 
    kubectl_cmd = f'get pods {" ".join(pods)} {ns} -o json'

    result = json.loads(kubectl.run_kubectl(kubectl_cmd))
    # get pods can return a list of items or a single pod
    pod_list =  []
    if result.get('items'):
        pod_list = result['items']
    else:
        pod_list.append(result)
    for pod in pod_list:
        pod_status = pod['status']['phase']
        pod_name   = pod['metadata']['name']
        assert pod_status in statuses, f'Pod {pod_name} status {pod_status} not in expected statuses: {", ".join(statuses)}'

def wait(func, *args, **kwargs):

    class TimeoutError(Exception):
        pass

    timeout = kwargs.pop("wait_timeout", 0)
    delay   = kwargs.pop("wait_delay", 0)
    backoff = kwargs.pop("wait_backoff", 0)
    retries = kwargs.pop("wait_retries", 0)
    allow   = kwargs.pop("wait_allow", ())
    elapsed = kwargs.pop("wait_elapsed", 0)

    if retries > 0 and elapsed > 0:
        raise ValueError("wait_retries and wait_elapsed cannot both have a non zero value")

    if retries == 0 and elapsed == 0:
        raise ValueError("either wait_retries  or wait_elapsed must have a non zero value")

    def _handle_timeout(signum, frame):
        raise TimeoutError()

    start = int(time.time())
    attempts = 1
    reason=""

    time.sleep(delay)
    while True:
        signal.signal(signal.SIGALRM, _handle_timeout)
        signal.alarm(timeout)
        try:
            return func(*args, **kwargs)
        except TimeoutError:
            if elapsed > 0 and int(time.time())-start >= elapsed:
               reason = "maximum wait time exceeded: {}s".format(elapsed)
               break
            reason = "timeout of {}s exceded".format(timeout)
        except allow as ex:
            reason = "{}: '{}'".format(ex.__class__.__name__, ex)
        finally:
            signal.alarm(0)

        if retries > 0 and attempts == retries:
            break

        time.sleep(backoff)

        attempts = attempts + 1

    raise Exception("Failed waiting for function {} after {} attemps due to {}".format(func.__name__, attempts, reason))

