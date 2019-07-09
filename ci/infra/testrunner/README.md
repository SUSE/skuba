
# Testrunner

## Local Dev Machine Setup For Openstack
1. Create a private directory 
```
sudo mkdir /Private
``` 
2. Download container-openrc.sh from openstack, change the file, and store the file in /Private

3. Add openstack password to downloaded container-openrc.sh like the following, save and move to /Private
```
export OS_USERNAME="YOUR USERNAME"
# With Keystone you pass the keystone password.
#echo "Please enter your OpenStack Password for project $OS_PROJECT_NAME as user $OS_USERNAME: "
#read -sr OS_PASSWORD_INPUT
#export OS_PASSWORD=$OS_PASSWORD_INPUT
export OS_PASSWORD="YOUR PASSWORD"
```
4. Edit and update `ci/infra/testrunner/vars/openstack.yaml`
```
workspace: "" # The top folder where skuba is stored
username: ""  # User deployed stack name
openrc: ""    # Path to openrc.sh file
skuba:        # skuba locations
  srcpath:    # Path to skuba srch project (defaults to `./skuba`)
  binpath     # Path to skuba bin directory (defaults to `<workspace>/go/bin/`)
terraform:
  plugin_dir: # Path to directory used for installing providers 
  stack_name: # Unique name for the terraform deployment
  tfdir:      # Path to the directory with terraform templates (defaults to <skuba.srcpath>/ci/infra/`)
              # under this directory there must be a subdirectory per platform (openstack, vmware, baremetal)
  tfvars:     # name of the tfvars file to be used (defatuls to `terraform.tfvars.ci.example`)
```

## Local Dev Machine Setup For VMware

1. Create an environment file e.g. `vmware-env.sh` with the following:
```
#!/usr/bin/env bash

export VSPHERE_SERVER="vsphere.cluster.endpoint.hostname"
export VSPHERE_USER="username@vsphere.cluster.endpoint.hostname"
export VSPHERE_PASSWORD="password"
export VSPHERE_ALLOW_UNVERIFIED_SSL="true"
```

2. Edit and update `ci/infra/testrunner/vars/vmware.yaml`
```
workspace: "" # The top folder where skuba is stored
username: "" # User deployed stack name
env_file: "" # Path to vmware-env.sh file
```

3. Be sure to use the `--vars` arg when calling testrunner and supply the path to `ci/infra/testrunner/vars/vmware.yaml`

## Testrunner Usage


### General

```
./testrunner -h
Starting ./testrunner script
usage:
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.

       [-h] [-v YAML_PATH]
      {info,log,cleanup,provision,build-skuba,bootstrap,status,join-node,remove-node,reset-node,test}
       ...

positional arguments:
  {info,log,cleanup,provision,build-skuba,bootstrap,status,join-node,remove-node,reset-node,test}
    info                ip info
    log                 gather logs from nodes
    cleanup             cleanup created skuba environment
    provision           provision nodes for cluster in your configured
                        platform e.g: openstack, vmware.
    build-skuba         build skuba environment
                        {workspace}/go/src/github.com/SUSE/skuba and build
                        skuba in that directory
    bootstrap           bootstrap k8s cluster with deployed nodes in your
                        platform
    status              check K8s cluster status
    join-node           add node in k8s cluster with the given role.
    remove-node         remove node from k8s cluster.
    reset-node          reset node reverting state previous to bootstap/join.
    test                execute tests

optional arguments:
  -h, --help            show this help message and exit
  -v YAML_PATH, --vars YAML_PATH
                        path for platform yaml file. Default is
                        vars/openstack.yaml in
                        {workspace}/ci/infra/testrunner. eg: -v
                        vars/myconfig.yaml
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

### Node commands (join, remove, reset)

```
  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0

```

### Test

```
optional arguments:
  -h, --help            show this help message and exit
  -s TEST_SUITE, --suit TEST_SUITE
                        test file name
  -t TEST, --test TEST  test to execute
  -l, --list            only list tests to be executed
  -v, --verbose         show all output
```

### Jenkins Machine Setup
1. In your Jenkins file, you need to set up environment variables. Then These environment variables will replace
variables in yaml file.

This pipeline script is same as openrc: "/Private/container-openrc.sh" in openstack.yaml.
As default, Jenkins has WORKSPACE environment variable so that workspace will be replaced in Jenkins workspace
```
   environment {
        OPENRC = credentials('openrc') or ENV_FILE = credentials('vmware-env') 
        STACK_NAME  = '' #unique name for this pipeline run
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack' or 'vmware'
   }
```
2. Set the package download mirror

Edit the `vars/<platform>.yaml` file and set the mirror for downloading packages for node setup
```
mirror: "ibs-mirror.prv.suse.net"
```

### Step to create K8s Cluster and start to use K8s cluster 
1. Cleanup before deploying nodes
```ci/infra/testrunner/testrunner -x ``` 
2. Deploy nodes in openstack  
```ci/infra/testrunner/testrunner -t ```  
3. Create skuba env and Build skuba and store in go bin dir
```ci/infra/testrunner/testrunner -c ```
4. Bootstraping a cluster
```ci/infra/testrunner/testrunner -b ```

Once bootstrapping is done you will be ready to use K8s cluster.

5. To extend the cluster, you can add more node with 
```ci/infra/testrunner/testrunner -a -m 2 -w 2 ```

6. Use K8s
Once your nodes are bootstrapped, {worksapce}/test-cluster folder will be created. Inside test-cluster, Your kubeconfig file will be located in with the name of admin.conf in test-cluster folder.
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
