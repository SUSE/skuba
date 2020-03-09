import json
import signal
import time
import yaml

PREVIOUS_VERSION = "1.15.2"
CURRENT_VERSION = "1.16.2"


def check_nodes_ready(kubectl):
    # Retrieve the node name and the Ready condition (True, or False)
    cmd = ("get nodes -o jsonpath='{ range .items[*]}{@.metadata.name}{\":\"}"
           "{range @.status.conditions[?(.type==\"Ready\")]}{@.status}{\" \"}'")

    nodes = kubectl.run_kubectl(cmd).strip().split(" ")
    for node in nodes:
        node_name, node_status = node.split(":")
        assert node_status == "True", f'Node {node_name} is not Ready'


def node_is_ready(platform, kubectl, role, nr):
    node_name = platform.get_nodes_names(role)[nr]
    cmd = ("get nodes {} -o jsonpath='{{range @.status.conditions[*]}}"
           "{{@.type}}={{@.status}};{{end}}'").format(node_name)

    return kubectl.run_kubectl(cmd).find("Ready=True") != -1


def node_is_upgraded(kubectl, platform, role, nr):
    node_name = platform.get_nodes_names(role)[nr]
    for attempt in range(20):
        if platform.all_apiservers_responsive():
            # kubernetes might be a little bit slow with updating the NodeVersionInfo
            cmd = ("get nodes {} -o jsonpath="
                   "'{{.status.nodeInfo.kubeletVersion}}'").format(node_name)
            version = kubectl.run_kubectl(cmd)
            if version.find(PREVIOUS_VERSION) == 0:
                time.sleep(2)
            else:
                break
        else:
            time.sleep(2)

    # allow system pods to come up again after the upgrade
    wait(check_pods_ready,
         kubectl,
         namespace="kube-system",
         wait_delay=60,
         wait_backoff=30,
         wait_elapsed=60*10,
         wait_allow=(AssertionError))

    cmd = "get nodes {} -o jsonpath='{{.status.nodeInfo.kubeletVersion}}'".format(node_name)
    return kubectl.run_kubectl(cmd).find(CURRENT_VERSION) != -1


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
            reason = "timeout of {}s exceded".format(timeout)
        except allow as ex:
            reason = "{}: '{}'".format(ex.__class__.__name__, ex)
        finally:
            signal.alarm(0)

        if elapsed > 0 and int(time.time())-start >= elapsed:
            reason = "maximum wait time exceeded: {}s".format(elapsed)
            break

        if retries > 0 and attempts == retries:
            break

        time.sleep(backoff)

        attempts = attempts + 1

    raise Exception("Failed waiting for function {} after {} attemps due to {}".format(func.__name__, attempts, reason))


def setup_kubernetes_version(skuba, kubectl, kubernetes_version=None):
    """
    Initialize the cluster with the given kubernetes_version, bootstrap it and
    join nodes.
    """

    skuba.cluster_init(kubernetes_version)
    skuba.node_bootstrap()

    skuba.join_nodes()

    wait(check_nodes_ready,
         kubectl,
         wait_delay=60,
         wait_backoff=30,
         wait_elapsed=60*10,
         wait_allow=(AssertionError))


def create_skuba_config(kubectl, configmap_data, dry_run=False):
    return kubectl.run_kubectl(
        'create configmap skuba-config --from-literal {0} -o yaml --namespace kube-system {1}'.format(
            configmap_data, '--dry-run' if dry_run else ''
        )
    )


def replace_skuba_config(kubectl, configmap_data):
    new_configmap = create_skuba_config(kubectl, configmap_data, dry_run=True)
    return kubectl.run_kubectl("replace -f -", stdin=new_configmap.encode())


def get_skuba_configuration_dict(kubectl):
    skubaConf_yml = kubectl.run_kubectl(
        "get configmap skuba-config --namespace kube-system -o jsonpath='{.data.SkubaConfiguration}'"
    )
    return yaml.load(skubaConf_yml, Loader=yaml.FullLoader)
