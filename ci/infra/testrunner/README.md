# Testrunner

## Contents

- [Summary](#summary)
- [Design](#design)
- [Configuration](#configuration-parameters)
  - [Packages](#packages)
  - [Utils](#utils)
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
  - [Provision](#provision-command)
  - [Node commands](#node-commands)
    - [Node Upgrade](#node-upgrade-command)
    - [Ssh](#ssh-command)
  - [Test](#test-command)
- [Examples](#examples)
  - [Create K8s Cluster](#create-k8s-cluster)
  - [Collect logs](#collect-logs)
  - [Install using registration code](#install-using-registration-code)
  - [Install packages from mirror](#install-packages-from-mirror)

## Summary

Testrunner is a CLI tool for setting up an environment for running e2e tests, making transparent the mechanism used for providing the test infrastructure. It can be used as a stand-alone tool, running either locally or as part of a CI pipeline. It provides commands for provisioning the infrastructure, deploying a k8s cluster using `skuba`,  and running tests. It also provides [a library](tests/README.md) for developing `pytest`-based tests.


## Design

The `testrunner` is composed by multiple components as described in the figure below:
* The `testrunner.py` frontend cli application, which receives CLI options, loads a configuration and executes the selected command. The configuration is loaded from a yaml file (by default, [vars.yaml](vars.yaml)) and mixed with the environment variables.
* A set of supporting classes which offer functionality for setting up and interacting with a cluster and executing tests. These classes wrap external tools such as `skuba`, `terraform`, `kubectl` and `pytests`. The commands offered by the `testrunner.py` cli application mostly expose the functionalities of these classes and add only the required glue code.
* In the case of the test command, the `testrunner` does not only wrap the `pytest` testing tool ([tests/driver.py](tests/driver.py), but also offers a test library ([tests/conftest.py](tests/conftest.py)) which implements reusable test functionalities as well as `fixtures` to facilitate [test development](tests/README.md). These fixtures use the libraries offered by testrunner for setting up the test infrastructure, deploying the cluster, interact with the cluster and the nodes, cleaning up after execution, execute common validation checks, among others.

The objective of this design is to maintain a clear separation between the following concerns:
* User interface including configuration (`testrunner.py` and `BaseConfig` class)
* Wrapping external classes (supporting libraries for `skuba`, `terraform`, `kubectl`)
* Reusable test functionalities (`conftest.py`)
* Test logic (diverse test suites)

It is worth noticing that tests can be executed directly using pytest, but it is more convenient to execute them using the testrunner, resulting in a consistent user experience.

```                                           
                                              | env variables
                                              v
              +--------------+         +----------------+     +---------------+
  CLI options |              |<--------|                |     | Configuration |
  ----------->|  Testrunner  |         | Initialization |<----|  (vars.yaml)  |
              |              |----+    |                |     |               |
              +--------------+    |    +----------------+     +---------------+
                    |             |
       Test command |             | Setup commands 
                    v             v (provision, bootstrap, ...)
        +--------------+    +--------------+
        |    pytest    |    |  Supporting  | Wrap skuba, 
        |    wrapper   |    |  libraries   | terraform,
        |              |    |              | kubectl
        +--------------+    +--------------+
                |                   ^
   Invokes with |                   | Use
   cli optons   v                   |
        +--------------+    +--------------+
        |              |    |              | reusable test
        |    pytest    |    | Test library | fixtures
        |              |    |              |
        +--------------+    +--------------+
                |                   ^
       Execute  |                   |
                v                   |
        +--------------+            | Use
        |              |            |
        |    Tests     |------------+
        |              |
        +--------------+
```

## Configuration parameters

Testrunner provides configuration by means of:
 
- A yaml configuration file (defaults to `vars.yaml` in current directory)
- Environment variables that override the configuration. Every configuration option of the form `<section>.<variable>' can be subtituted by an environment variable `SECTION_VARIABLE. For example `skuba.binpath` is overriden by `SKUBA_BINPATH`.
- CLI options which override configuration parameters such as the logging level (see [Usage](#usage))

The following sections document the configuration options. The CLI arguments are described in the [Usage section](#usage).

#### Packages
The `packages` section configures the source of the packages to be installed in the nodes:

* additional_pkgs: list with additional packages to be installed in the nodes. For example, for installing SUSE certificates for self-signed packages in development environments:
```
packages:
  additional_pkgs:
  - "ca-certificates-suse"
```
* additional_repos: repositories to be added to the nodes. For example, for installing maintenance updates. It takes the form of a map:
```
packages:
  additional_repos:
    repo1: url/to/repo1
    repo2: url/to/repo2
```
* mirror: URL for the repository mirrors to be used when setting up the skuba nodes, replacing the URL of the repositories defined in terraform. Used, for instance, to switch to development repositories or internal repositories when running in the CI pipeline.
* registry_code: code use for registering CaaSP product. If specified, the registries from the tfvars are ignored. Additional repositories can still be defined using the `maintenance` configuration parameter.


### Utils

This section configures the utils module used for executing commands.

* ssh_key: specifies the location of the key used to access nodes. The default is to use the user's key located at `$HOME/.ssh/id_rsa`.
* ssh_sock: name of the socket used to communicate with the ssh-agent. Default is /tmp/testrunner_ssh_sock'

Example:
```
utils:
  ssh_sock: "/path/to/ssh-agent/socket"
```
### Platform

This section configures general platform-independent parameters. Platform dependent parameters are defined in the corresponding sections (Terraform, Openstack, VMware)

- log_dir: path to the directory where platform logs are collected. Defaults to `$WORKSPACE/platform_logs`

```
log_dir: "/path/to/log/dir/
```

#### Terraform

General setting for terraform-based platforms such as [Openstack](#openstack) and [VMware](#vmware). 

* internal_net: name of the network used when provisioning the platform. Defaults to `stack_name`
* lb: specifications for the load balancer(s)
* nodeuser: the user name used to login into the platform nodes. Optional.
* master: specifications for the master(s)
* plugin_dir: directory used for retrieving terraform plugins. If not set, plugins are installed using terraform [discovery mechanism](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery)
* retries: maximum number of attempts to recover from failures during terraform provisioning 
* stack name: the unique name of the platform stack on the shared infrastructure, used as prefix by many resources such as networks, nodes, among others. Default is "$USER" 
* tfdir: path to the terraform files. Testrunner must have writing permissions to this directory. Defaults to `$WORKSPACE/skuba/ci/infra`.
* tfvars: name of the terraform variables file to be used. Defaults to "terraform.tfvars.json.ci.example"
* workdir: working directory on which tfout file will be generated. Default is `$WORKSPACE`
* worker: specifications for the worker(s)

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

The Skuba section defines the location and execution options for the `skuba` command. As `testrunner` can be used either from a local development or testing environment or a CI pipeline, the configuration allows to define the location of the binary.

* binpath: path to skuba binary. Default is "$WORKSPACE/go/bin/skuba"
* cluster: name of the cluster. Default is "test-cluster"
* verbosity: verbosity level for skuba command execution
* workdir: working directory on which cluster is initialized. Default is "$WORKSPACE"

Example:
```
skuba:
  binpath: "/usr/bin/skuba"
  verbosity: 10
```

### Kubectl

The kubectl section defines the configuration of the kubectl tool. 
* binpath: path to the kubectl binary. Defaults to `/usr/bin/kubectl`
* kubeconf: path to the kubeconfig file. defaults to `<workspace>/test-cluster/admin.conf`

### Log

Testrunner sends output to both a console and file logger handlers, configured using the following `log` variables:

* file: path to the file used to send a copy of the log with verbosity `DEBUG`. Default is "$WORKSPACE/testrunner.log"
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

Different components of `testrunner` requires a working directory. In particular, skuba for maintaining the test cluster configuration. The content of this directory can be erased or overwritten by `testrunner`. Be sure you create a directory to be used as workspace which is NOT located under your local working copy of the `skuba` project. By default, the working directory is taken from environment variable `WORKSPACE`:

```
export WORKSPACE="/path/to/workspace"
```


#### skuba and platform

Set the `skuba` and `terraform parameters depending on how you are testing `skuba`: 
* If testing from local source:
```
  skuba:
    binpath: "/path/to/go/bin/directory"
```

Be sure you `terraform.tfdir` directory to point to the `ci/infra` directory in the local `skuba` repo:
```
  terraform:
    tfdir: "/path/to/local/skuba/repo/ci/infra"
```


* If testing from installed package

Use skuba binary installed from the package
```
  skuba:
    binpath: "/usr/bin/"
```

You must copy the terraform files installed from the package to a work directory and set the `tfdir` directory accordingly:
```
  terraform:
    tfdir: "/path/to/terraform/files"
```

You must provide the ssh key to connect to the cluster nodes, as the `shared_id` key used in development is not available. By default, your `id_rsa` key will be used, but you can provide any key:

```
  utils:
    ssh_key: "path/to/id_rsa"
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

#### Libvirt

`testrunner` can provision a cluster of virtual machines using terraform libvirt provider. The only noticeable difference with other platforms is the dependency on the terraform libvirt provider plugin which is not available from the official terraform plugin site, neither is delivered as part of the CaaSP packages. However this is available from the development [CaaSP repositories](http://download.suse.de/ibs/Devel:/CaaSP:/4.0/SLE_15_SP1/) for SLE15-SP1 or from the public [openSUSE repositories](https://build.opensuse.org/package/show/systemsmanagement:terraform:unstable/terraform-provider-libvirt) for non SLE15-SP1 hosts. Notice it requires an updated version of libvirt (4.1.0 or above).

There three configuration variables required for libvirt operations, all configurable from the configuration yaml file or by environment variables like any other variable in the yaml file:

```yaml
libvirt:
  uri: "qemu:///system" #os.getenv("LIBVIRT_URI")
  keyfile: "" #os.getenv("LIBVIRT_KEYFILE")
  image_uri: "" #os.getenv("LIBVIRT_IMAGE_URI")
```

`uri` is the URI used by the libvirt client to connect the libvirt host `qemu:///system` for local libvirt services.
`keyfile` is the path of the keyfile used to connect to the libbirt `uri`. This path is included as part of the `uri` as a query. Consider a remote ssh uri as `qemu+ssh://<user>@<libvirt_host>/system`, defining a `keyfile` turns it into `qemu+ssh://<user>@<libvirt_host>/system?keyfile=<keyfile_path>`
`image_uri` is the URI that will be used to pull the image for the VMs deployment, usually this points to some JeOS image.Note the image is expected to include cloud-init.

### Jenkins Setup

In your Jenkins file, you need to set up environment variables which will replace options in the yaml file. This is more convenient than having to edit the yaml file in the CI pipeline. 

#### Work environment

By default, Jenkins has a `WORKSPACE` environment variable so that `workspace` will be replaced automatically.

#### Skuba

Jenkins checks out the `skuba` repository under the `workspace` directory and generates the binaries also under the `workspace`, which are the default locations. Therefore, there is no need to specify any location:

```
skuba:
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

       [-h] [-v YAML_PATH] [-p {openstack,vmware,bare-metal,libvirt}] [-c]
       [-l {DEBUG,INFO,WARNING,ERROR}]
       {info,config,get_logs,cleanup,provision,bootstrap,deploy,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,join-nodes,ssh,test,inhibit_kured}
       ...

positional arguments:
  {info,get_logs,cleanup,provision,bootstrap,deploy,status,cluster-upgrade-plan,join-node,remove-node,node-upgrade,join-nodes,ssh,test,inhibit_kured}
                        command
    info                ip info
    config              print configuration 
    get_logs            gather logs from nodes
    cleanup             cleanup created skuba environment
    provision           provision nodes for cluster in your configured
                        platform e.g: openstack, vmware.
    bootstrap           bootstrap k8s cluster
    deploy              initializes, bootstrap and join all nodes k8s
    status              check K8s cluster status
    cluster-upgrade-plan
                        Cluster upgrade plan
    check-node          check node health
    join-node           add node in k8s cluster with the given role.
    remove-node         remove node from k8s cluster.
    node-upgrade        upgrade kubernetes version in node
    join-nodes          add multiple provisioned nodes k8s.
    ssh                 Execute command in node via ssh.
    test                execute tests
    inhibit_kured       Prevent kured to reboot nodes


optional arguments:
  -h, --help            show this help message and exit
  -v YAML_PATH, --vars YAML_PATH
                        path for platform yaml file. Default is vars.yaml. eg:
                        -v myconfig.yaml
  -p {openstack,vmware,bare-metal,libvirt}, --platform {openstack,vmware,bare-metal,libvirt}
                        The platform you're targeting. Default is openstack
  -l {DEBUG,INFO,WARNING,ERROR}, --log-level {DEBUG,INFO,WARNING,ERROR}
                        log level
  -c, --print-conf      prints the configuration

```

### Provision command

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
  -c, --cloud-provider  Use cloud provider integration
  -t TIMEOUT, --timeout TIMEOUT
                        timeout for waiting a node to become ready (seconds)

```

### Deploy

```
optional arguments:
  -h, --help            show this help message and exit
  -k KUBERNETES_VERSION, --kubernetes-version KUBERNETES_VERSION
                        kubernetes version
  -c, --cloud-provider  Use cloud provider integration
  -t TIMEOUT, --timeout TIMEOUT
                        timeout for waiting a node to become ready (seconds)
```

### Join nodes

```
  -h, --help            show this help message and exit
  -m MASTERS, --masters MASTERS
                        Specify how many masters to join. Default is all
  -w WORKERS, --workers WORKERS
                        Specify how many workers to join. Default is all
  -t TIMEOUT, --timeout TIMEOUT
                        timeout for waiting the master nodes to become ready (seconds)
```

### Check cluster 

Checks the status of the cluster. If no check is specified, all checks that apply to
the stage are executed.

```
  -c CHECKS [CHECKS ...], --check CHECKS [CHECKS ...]
                        check to be executed (multiple checks can be specified)
  -s STAGE, -stage STAGE
                        only execute checks that apply to this stage
```

### Node commands

Common parameters

```
  -h, --help            show this help message and exit
  -r {master,worker}, --role {master,worker}
                        role of the node to be added or deleted. eg: --role
                        master
  -n NODE, --node NODE  node to be added or deleted. eg: -n 0

```
#### Join Node

Joins node to cluster with the given role

```
  -t TIMEOUT, --timeout TIMEOUT
                        timeout for waiting a node to become ready (seconds)
```

#### Node Upgrade command

Upgrades node

```
  -h, --help            show this help message and exit
  -a {plan,apply}, --action {plan,apply}
                        action: plan or apply upgrade
```

#### Ssh command

Executes command in a node

```
  -c ..., --cmd ...     remote command and its arguments. e.g ls -al. Must be
                        last argument for ssh command
```

#### Check command

Checks the status of a node. If no check is specified, all checks that apply to
the node's role are executed.

```
  -c CHECKS [CHECKS ...], --check CHECKS [CHECKS ...]
                        check to be executed (multiple checks can be specified)
  -s STAGE, -stage STAGE
                        only execute checks that apply to this stage
```

### Test command

```
optional arguments:
  -h, --help            show this help message and exit
  -f MARK, --filter MARK
                        Filter the tests based on markers
  -j JUNIT, --junit JUNIT
                        Name of the xml file to record the results to.
  -m MODULE, --module MODULE
                        folder with the tests
  -s TEST_SUITE, --suite TEST_SUITE
                        test file name
  -t TEST, --test TEST  test to execute
  -l, --list            only list tests to be executed
  -v, --verbose         show all output from testrunner libraries
  --skip-setup {provisioned,bootstrapped,deployed}
                        Skip the given setup step. 'provisioned' For when you
                        have already provisioned the nodes. 'bootstrapped' For
                        when you have already bootstrapped the cluster.
                        'deployed' For when you already have a fully deployed
                        cluster.
  --traceback {long,short,line,no}
                        level of detail in traceback for test failure

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

All collected logs are stored at `path/to/workspace/platform_logs/`

Logs that are currently being collected are the cloud-init logs for each of the nodes:

    /var/run/cloud-init/status.json
    /var/log/cloud-init-output.log
    /var/log/cloud-init.log

These are stored each in their own folder named `path/to/workspace/platform_logs/{master|worker}_ip_address/`

### Install using registration code

1. Configure the registration code to be passed to nodes:

`vars.yaml`
```
packages:
  registry_code: "<registry code>"
```
2. Configure `testrunner` to use a `skuba` binary compatible with the version installed in the nodes:

`vars.yaml`
```
skuba:
  bin_path: "/path/to/skuba"
```

### Install packages from mirror

Specify the mirror an enable the installation of certifates package:

`vars.yaml`
```
packages:
  mirror: "my.mirror.site"
  certificates: "certificates-package"
```

