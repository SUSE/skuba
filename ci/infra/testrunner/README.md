
# Testrunner

## Local Dev Machine Setup
Set the following environment variables

```
WORKSPACE="/path/to/your/workspace" # e.g. $HOME/testrunner_workspace it is expected that the skuba code is at $HOME/testrunner_workspace/skuba
TERRAFORM_STACK_NAME="name-of-your-stack" # i.e. $USER
```

or

Copy `testrunner/vars.yaml` to `testrunner/myvars.yaml` set the variables in the there and use the `--vars myvars.yaml` arg

```
workspace: "" # Working directory for testrunner
...
terraform:
  stack_name: "" # name of the stack
```

You can also set things like where the skuba binary and src are
```
SKUBA_BINPATH=/path/to/bin/skuba
SKUBA_SRCPATH=/path/to/src/skuba
```

or

```
skuba:
  binpath: "" # path to skuba binary
  srcpath: "" # path to skuba source
```

Anything else you can override is in `testrunner/vars.yaml` or can also be set as an environment variable.

### Setup For Openstack
1. Download your openrc file from openstack

2. Add your openstack password to the downloaded openrc.sh like the following
```
export OS_USERNAME="YOUR USERNAME"
# With Keystone you pass the keystone password.
#echo "Please enter your OpenStack Password for project $OS_PROJECT_NAME as user $OS_USERNAME: "
#read -sr OS_PASSWORD_INPUT
#export OS_PASSWORD=$OS_PASSWORD_INPUT
export OS_PASSWORD="YOUR PASSWORD"
```
4. Set `OPENSTACK_OPENRC=/path/to/openrc.sh`

or

```
openstack:
  openrc: ""
```

### Setup For VMware

1. Create an environment file e.g. `vmware-env.sh` with the following:
```
#!/usr/bin/env bash

export VSPHERE_SERVER="vsphere.cluster.endpoint.hostname"
export VSPHERE_USER="username@vsphere.cluster.endpoint.hostname"
export VSPHERE_PASSWORD="password"
export VSPHERE_ALLOW_UNVERIFIED_SSL="true"
```

2. Set `VMWARE_ENV_FILE=/path/to/vmware-env.sh`

or

```
vmware:
  env_file: ""
```

3. Be sure to use the `-p|--platform` arg when calling testrunner and set it to `vmware`

## Testrunner Usage


### General

```
./testrunner --help
usage:
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.

       [-h] [-v YAML_PATH] [-p {openstack,vmware,bare-metal,libvirt}]
       [-l {DEBUG,INFO,WARNING,ERROR}]
       {info,get_logs,cleanup,provision,build-skuba,bootstrap,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,ssh,test}
       ...

positional arguments:
  {info,get_logs,cleanup,provision,build-skuba,bootstrap,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,ssh,test}
                          command
    info                  ip info
    get_logs              gather logs from nodes
    cleanup               cleanup created skuba environment
    provision             provision nodes for cluster in your configured
                          platform e.g: openstack, vmware.
    build-skuba           build skuba environment
                          {workspace}/go/src/github.com/SUSE/skuba and build
                          skuba in that directory
    bootstrap             bootstrap k8s cluster with deployed nodes in your
                          platform
    status                check K8s cluster status
    cluster-upgrade-plan  cluster upgrade plan
    join-node             add node in k8s cluster with the given role.
    remove-node           remove node from k8s cluster.
    node-upgrade          plan or apply kubernetes version upgrade in node
    ssh                   Execute command in node via ssh.
    test                  execute tests

optional arguments:
  -h, --help            show this help message and exit
  -v YAML_PATH, --vars YAML_PATH
                        path for platform yaml file. Default is vars.yaml. eg:
                        -v myconfig.yaml
  -p {openstack,vmware,bare-metal,libvirt}, --platform {openstack,vmware,bare-metal,libvirt}
                        The platform you're targeting. Defaults to openstack
  -l {DEBUG,INFO,WARNING,ERROR}, --log-level {DEBUG,INFO,WARNING,ERROR}
                        log level

```

### Provision

```
optional arguments:
  -h, --help            show this help message and exit
  -m MASTER_COUNT, -master-count MASTER_COUNT
                        number of masters nodes to be deployed. eg: -m 2
  -w WORKER_COUNT, --worker-count WORKER_COUNT
                        number of workers nodes to be deployed. eg: -w 2
```

### Node commands (join, remove)
```
  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0

```

### Node Upgrade

```
  -h, --help            show this help message and exit
  -a {plan,apply}, --action {plan,apply}
                        action: plan or apply upgrade

```

### Ssh

  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0
  -c ..., --cmd ...     remote command and its arguments. e.g ls -al. Must be
                        last argument for ssh command

### Test

```
optional arguments:
  -h, --help            show this help message and exit
  -s TEST_SUITE, --suite TEST_SUITE
                        test file name
  -t TEST, --test TEST  test to execute
  -l, --list            only list tests to be executed
  -v, --verbose         show all output
```

### Jenkins Machine Setup
In your Jenkins file, you need to set up environment variables. Then These environment variables will replace
variables in the yaml file.

As default, Jenkins has WORKSPACE environment variable so that workspace will be replaced in Jenkins workspace
```
   environment {
        OPENSTACK_OPENRC = credentials('openrc') or VMWARE_ENV_FILE = credentials('vmware-env')
        TERRAFORM_STACK_NAME  = '' #unique name for this pipeline run
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack' or 'vmware'
   }
```

### Step to create K8s Cluster and start to use K8s cluster
1. Deploy nodes to openstack
```ci/infra/testrunner/testrunner provision```
2. Build skuba and store at `SKUBA_BINPATH` defaults to $WORKSPACE/go/bin/skuba
```ci/infra/testrunner/testrunner build-skuba```
3. Initialize the control plane
```ci/infra/testrunner/testrunner bootstrap```
4. Join nodes
```ci/infra/testrunner/testrunner join-node --role worker --node 0```

5. Use K8s
Once your nodes are bootstrapped, $WORKSPACE/test-cluster folder will be created. Inside test-cluster, Your kubeconfig file will be located in with the name of admin.conf in test-cluster folder.
```
chang@~/Workspace/vNext/test-cluster$ kubectl get pods --all-namespaces --kubeconfig=./admin.conf
NAMESPACE     NAME                                  READY     STATUS    RESTARTS   AGE
kube-system   cilium-6mnrh                          1/1       Running   0          3m
kube-system   cilium-z9rqm                          1/1       Running   0          3m
kube-system   coredns-559fbd6bb4-gw7cn              1/1       Running   0          4m
kube-system   coredns-559fbd6bb4-jqt4r              1/1       Running   0          4m
kube-system   etcd-my-master-0                      1/1       Running   0          3m
kube-system   kube-apiserver-my-master-0            1/1       Running   0          3m
kube-system   kube-controller-manager-my-master-0   1/1       Running   0          3m
kube-system   kube-proxy-782z2                      1/1       Running   0          4m
kube-system   kube-proxy-kf7g5                      1/1       Running   0          3m
kube-system   kube-scheduler-my-master-0            1/1       Running   0          3m
```

### Collected Logs
All collected logs are stored at `path/to/workspace/testrunner_logs/`

Logs that are currently being collected are the cloud-init logs for each of the nodes:

    /var/run/cloud-init/status.json
    /var/log/cloud-init-output.log
    /var/log/cloud-init.log

These are stored each in their own folder named `path/to/workspace/testrunner_logs/{master|worker}_ip_address/`
