# Testrunner

## Contents

- [Summary](#summary)
- [Configuration](#configuration-parameters)
  - [Work Environment](#work-environment)
  - [Platform](#platform)
    - [Terraform](#terraform)
    - [Openstack](#openstack)
    - [VMware](#vmware)
  - [Skuba](#skuba)
  - [Log](#log)
  - [Test](#test)
- [Environment setup](#environment-setup)
  - [Local setup](#local-setup)
  - [CI setup](#ci-setup)
- [Usage](#usage)
  - [General CLI options](#general-cli-options)
  - [Provision](#provision)
  - [Node commands](#node-commands)
    - [Node Upgrade](#node-upgrade)
  - [Ssh](#ssh)
  - [Test](#ssh)
- [Examples](#examples)
  - [Create K8s Cluster](#create-k8s-cluster)
  - [Collect logs](#collect-logs)

## Summary

Testrunner is a CLI tool for setting up an environment for running e2e tests, making transparent the mechanism used for providing the test infrastructure. It can be used as a stand-alone tool, running either locally or as part of a CI pipeline. It provides commands for provisioning the infrastructure, deploying a k8s cluster using `skuba`,  and running tests. It also provides [a library](tests/README.md) for developing `pytest`-based tests.

## Configuration parameters

Testrunner provides configuration by means of:
 
- A yaml configuration file (defaults to `vars.yaml` in current directory)
- Environment variables that override the configuration. Every configuration option of the form `<section>.<variable>' can be subtituted by an environment variable `SECTION_VARIABLE. Notice that some variables are defined in the "root" section (e.g `workspace`). For example `skuba.binpath` is overriden by `SKUBA_BINPATH` and `workspace` by `WORKSPACE`
- CLI options which override configuration parameters such as the logging level (see [Usage](#usage))

The following sections document the configuration options. The CLI arguments are described in the [Usage section](#usage).

### Work Environment

This section configures the working environment and is generally specific of each user of CI job-

- workspace: path to the testrunner's working directory 
- username: user name for the user. It's optional. If `platform.stack_name` is not set, `username` is used.

```
workspace: "/path/to/your/workspace" 
username: "username"
```

### Platform

There are some arguments that are currently at the top level of the configuration, but are actually related to the platform:

- log_dir: path to the directory where platform logs are collected. Defaults to `<workspace>/testrunner_logs` 
- nodeuser: the user name used to login into the platform nodes. Optional. 
- ssh_option: specifies the location of the key used to access nodes. Can have two possible values:
  - "id_shared": uses the key located at `<skuba src directory>/ci/infra/id_shared` (default)
  - "id_rsa": uses the user's key located at `$HOME/.ssh/id_rsa`

#### Terraform

General setting for terraform-based platforms such as [Openstack](#openstack) and [VMware](#vmware). 

* internal_net: name of the network used when provisioning the platform. Defaults to `stack_name`
* mirror: URL for the repository mirrors to be used when setting up the skuba nodes, replacing the URL of the repositories defined in terraform. Used, for instance, to switch to development repositories or internal repositories when running in the CI pipeline.
* plugin_dir: directory used for retrieving terraform plugins. If not set, plugins are installed using terraform [discovery mechanism](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery)
* retries: maximum number of attempts to recover from failures during terraform provisioning 
* stack name: the unique name of the platform stack on the shared infrastructure, used as prefix by many resources such as networks, nodes, among others. If not specified, the `username` is used.
* tfdir: path to the terraform files. Testrunner must have writing permissions to this directory. Defaults to `skuba.srcpath/ci/infra`.
* tfvars: name of the terraform variables file to be used. Defaults to "terraform.tfvars.json.ci.example"

Example
```
terraform:
  stack_name="my-test-stack"
```

#### Openstack

* openrc: path to the environment setup script

Example:
```
openstack:
  openrc: "/home/myuser/my-openrc.sh"
```

#### VMware

* env_file: path to environment variables file

Example:
```
vmware: 
   env_file: "/path/to/env/file"
```

### Skuba

The Skuba section defines the location and execution options for the `skuba` command. As `testrunner` can be used either from a local development or testing environment or a CI pipeline, the configuration allows to define the location of both the source and the binary. Please notice that the source location is used as default location for other configuration elements, such as terraform files, even if the `skuba` binary is specified. 

* binpath: path to skuba binary
* srcpath: path to skuba source. Used to locate other resources, like terraform configuration, and ssh keys.
* verbosity: verbosity level for skuba command execution

```
skuba:
  binpath: "/usr/bin/"
  srcpath: "/go/src/github.com/SUSE/skuba"
  verbosity: 5
```

### Log

Testrunner sends output to both a console and file logger handlers, configured using the following `log` variables:

* file: name of the file used to send a copy of the log with verbosity `DEBUG`. This file is located under the `workspace` directory.
* level: debug verbosity level to console. Can be any of `DEBUG`, `INFO`, `WARNING`, `ERROR`. Defaults to `INFO`.
* overwrite: boolean that indicates if the content of the log file must be overwritten (`True`) or log entries must be appended at the end of the file if it exists. Defaults to `False` (do not overwrite) 
* quiet: boolean that indicates if `testrunner` will send any output to console (`False`) or not will execute silently (`True`). Quiet mode is useful when `testrunner` is used as a library. Defaults to `False`.

Example:
```
log:
  level: DEBUG
```

### Test

* no_destroy: boolean that indicates if provisioned resources should be deleted when test ends. Defaults to `False`

```
no_destroy: True  #keep resources after test ends
```

## Environment Setup

This section details how to setup `testrunner` 

### Local setup

Copy `vars.yaml` to `/path/to/myvars.yaml`, set the variables according to your environment and needs, and use the `--vars /path/to/myvars.yaml` CLI argument when running `testrunner`.

#### Work Environment

`testrunner` requires a working directory, which is specified in the `workspace` configuration parameter. The content of this directory can be erased or overwritten by `testrunner`. Be sure you create a directory to be used as workspace which is not located under your local working copy of the `skuba` project.


```
workspace: "/path/to/workspace"
username:  "my-user-test"
```


#### skuba and platform

Set the `skuba` and `terraform parameters depending on how you are testing `skuba`: 
* If testing from local source:
```
  skuba:
    srcpath: "/path/to/local/skuba/repo"
    binpath: "path/to/go/bin/directory"
```

Be sure you don't specify the `terraform.tfdir` directory, so terraform configuration from the local `skuba` repo are used.


* If testing from installed package

```
  skuba:
    binpath: "/usr/bin/"

  terraform:
    tfdir: "/path/to/terraform/files"
```

You use your `id_rsa` keys to connect to the cluster nodes, as the `shared_id` is not available.

```
ssh_option: "id_rsa"
```

#### Open Stack

1. Download your openrc file from openstack

2. Optionally, add your openstack password to the downloaded openrc.sh as shown below:
```
# With Keystone you pass the keystone password.
#echo "Please enter your OpenStack Password for project $OS_PROJECT_NAME as user $OS_USERNAME: "
#read -sr OS_PASSWORD_INPUT
#export OS_PASSWORD=$OS_PASSWORD_INPUT
export OS_PASSWORD="YOUR PASSWORD"
```
3. Set the path to the openrc file in the testrunner's vars file:

```
openstack:
  openrc: "/path/to/openrc.sh"
```

or as an environment variable: ```export OPENSTACK_OPENRC=/path/to/openrc.sh```

#### VMware

1. Create an environment file e.g. `vmware-env.sh` with the following:
```
#!/usr/bin/env bash

export VSPHERE_SERVER="vsphere.cluster.endpoint.hostname"
export VSPHERE_USER="username@vsphere.cluster.endpoint.hostname"
export VSPHERE_PASSWORD="password"
export VSPHERE_ALLOW_UNVERIFIED_SSL="true"
```

2. Set the path to the VMware environment file in the testrunner's vars file:

```
vmware:
  env_file: "/path/to/vmware-env.sh"
```
or as an environment variable: `export VMWARE_ENV_FILE=/path/to/vmware-env.sh`

3. Be sure to use the `-p` or `--platform` argument when invoking `testrunner` and set it to `vmware`, otherwise `openstack` is used. 

### Jenkins Setup

In your Jenkins file, you need to set up environment variables which will replace options in the yaml file. This is more convenient than having to edit the yaml file in the CI pipeline. 

#### Work environment

By default, Jenkins has a `WORKSPACE` environment variable so that `workspace` will be replaced automatically.

#### Skuba

Jenkins checks out the `skuba` repository under the `workspace` directory and generates the binaries also under the `workspace`, which are the default locations. Therefore, there is no need to specify any location:

```
skuba:
  srcpath: ""
  binpath: ""
``` 
 
#### Terraform

It is advisable to use a unique id related to the job execution as the terraform stack name:

```
TERRAFORM_STACK_NAME  = "${JOB_NAME}-${BUILD_NUMBER}" 
```

#### Openstack

Set the path to the `openrc` file using jenkins's builtin `credentials` directive.
 
```
 OPENSTACK_OPENRC = credentials('openrc')
```

#### VMware

Set the path to `env_file` using jenkins's builtin `credentials` directive.
```
VMWARE_ENV_FILE = credentials('vmware-env')
```

### Example
 
```
   environment {
        OPENSTACK_OPENRC = credentials('openrc')
        TERRAFORM_STACK_NAME  = "${JOB_NAME}-${BUILD_NUMBER}" 
        GITHUB_TOKEN = credentials('github-token')
        PLATFORM = 'openstack'
   }
```

## Usage

### General CLI options

```
./testrunner --help
usage:
    This script is meant to be run manually on test servers, developer desktops, or Jenkins.
    This script supposed to run on python virtualenv from testrunner. Requires root privileges.
    Warning: it removes docker containers, VMs, images, and network configuration.

       [-h] [-v YAML_PATH] [-p {openstack,vmware,bare-metal,libvirt}]
       [-l {DEBUG,INFO,WARNING,ERROR}]
       {info,get_logs,cleanup,provision,bootstrap,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,ssh,test}
       ...

positional arguments:
  {info,get_logs,cleanup,provision,bootstrap,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,ssh,test}
                          command
    info                  ip info
    get_logs              gather logs from nodes
    cleanup               cleanup created skuba environment
    provision             provision nodes for cluster in your configured
                          platform e.g: openstack, vmware.
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
### Bootstrap

 ```
optional arguments:
  -h, --help            show this help message and exit
  -k KUBERNETES_VERSION, --kubernetes-version KUBERNETES_VERSION
                        kubernetes version
  -c, --cloud-provider
                        The cloud provider will be offered
```

### Node commands

```
  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0

```

#### Node Upgrade

```
  -h, --help            show this help message and exit
  -a {plan,apply}, --action {plan,apply}
                        action: plan or apply upgrade
```

### Ssh

```
  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0
  -c ..., --cmd ...     remote command and its arguments. e.g ls -al. Must be
                        last argument for ssh command
```

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


## Examples 

### Create K8s Cluster

1. Deploy nodes to openstack
```./testrunner provision```
2. Initialize the control plane
```./testrunner bootstrap```
3. Join nodes
```./testrunner join-node --role worker --node 0```

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

### Collect logs

```./testrunner get_logs```

All collected logs are stored at `path/to/workspace/testrunner_logs/`

Logs that are currently being collected are the cloud-init logs for each of the nodes:

    /var/run/cloud-init/status.json
    /var/log/cloud-init-output.log
    /var/log/cloud-init.log

These are stored each in their own folder named `path/to/workspace/testrunner_logs/{master|worker}_ip_address/`
