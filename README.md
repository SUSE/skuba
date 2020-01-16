# skuba

Tool to manage the full lifecycle of a cluster.

## Table of Content

- [Prerequisites](#prerequisites)
- [Installation](#installation)
  * [Development](#development)
  * [Staging](#staging)
  * [Release](#release)
- [Creating a cluster](#creating-a-cluster)
  * [cluster init](#cluster-init)
  * [node bootstrap](#node-bootstrap)
- [Growing a cluster](#growing-a-cluster)
  * [node join](#node-join)
    + [master node join](#master-node-join)
    + [worker node join](#worker-node-join)
- [Shrinking a cluster](#shrinking-a-cluster)
  * [node remove](#node-remove)
- [kubectl-caasp](#kubectl-caasp)
- [Demo](#demo)
- [CI](ci/README.md)
- [Update](skuba-update/README.md)

## Prerequisites

The required infrastructure for deploying CaaSP needs to exist beforehand, it's
required for you to have SSH access to these machines from the machine that you
are running `skuba` from. `skuba` requires you to have added your SSH
keys to the SSH agent on this machine, e.g:

```sh
ssh-add ~/.ssh/id_rsa
```

If you want to perform an HA deployment you also need to set up a load balancer,
depending on your needs this setup can be as advanced as required.

The target nodes must have some packages already preinstalled:

  * cri-o
  * kubelet
  * kubeadm

The terraform based deployments are taking care of fulfilling these requirements.

## Installation

```sh
go get github.com/SUSE/skuba/cmd/skuba
```

### Development

A development build will:

* Pull container images from `registry.suse.de/devel/caasp/4.0/containers/containers/caasp/v4`

To build it, run:

```sh
make
```

### Staging

A staging build will:

* Pull container images from `registry.suse.de/suse/sle-15-sp1/update/products/casp40/containers/caasp/v4`

To build it, run:

```sh
make staging
```

### Release

A release build will:

* Pull container images from `registry.suse.com/caasp/v4`

To build it, run:

```sh
make release
```

## Creating a cluster

Go to any directory in your machine, e.g. `~/clusters`. From there, execute:

### cluster init

The `init` process creates the definition of your cluster. Ideally there's
nothing to tweak in the general case, but you can go through the generated
configurations and check if everything is fine for your taste.

```
skuba cluster init --control-plane load-balancer.example.com company-cluster
```

This command will have generated a basic project scaffold in the `company-cluster`
folder. You need to change the directory to this new folder in order to run the rest
of the commands in this README.

### node bootstrap

You need to bootstrap your first master node of the cluster. For this purpose
you have to be inside the `company-cluster` folder.

```
skuba node bootstrap --user opensuse --sudo --target <IP/fqdn> my-master
```

You can check `skuba node bootstrap --help` for further options, but the
previous command means:

* Bootstrap node using a SSH connection to target `<IP/fqdn>`
  * Use `opensuse` user when opening the SSH session
  * Use `sudo` when executing commands inside the machine
* Name the node `my-master`: this is what Kubernetes will use to refer to your node

When this command has finished, some secrets will have been copied to your
`company-cluster` folder. Namely:

* Generated secrets will be copied inside the `pki` folder
* The administrative `admin.conf` file of the cluster has been copied in
  root of the `company-cluster` folder

## Growing a cluster

### node join

Joining a node allows you to grow your Kubernetes cluster. You can join master nodes as
well as worker nodes to your existing cluster. For this purpose you have to be inside the
`company-cluster` folder.

This task will automatically create a new bootstrap token on the existing cluster that will
be used for the kubelet TLS bootstrap to happen on the new node. The token will be fed
automatically to the configuration used to join the new node.

This task will create the configuration file inside the `kubeadm-join.conf.d` folder as well
with a file named `<IP/fqdn>.conf` that will contain the join configuration used. If this file
existed before it will be honored, only overriding a small subset of settings automatically:

* Bootstrap token to the one generated on demand
* Kubelet extra args
  * `node-ip` if the `--target` is an IP address
  * `hostname-override` to the `node-name` provided as an argument
  * `cni-bin-dir` directory location if required
* Node registration name to `node-name` provided as an argument

#### master node join

This command will join a new master node to the cluster. This will also increase the etcd
member count by one.

```
skuba node join --role master --user opensuse --sudo --target <IP/fqdn> second-master
```

#### worker node join

This command will join a new worker node to the cluster.

```
skuba node join --role worker --user opensuse --sudo --target <IP/fqdn> my-worker
```

## Shrinking a cluster

### node remove

It's possible to remove master and worker nodes from the cluster. All the required tasks to remove
the target node will be performed automatically:

* Drain the node (also cordoning it)
* Mask and disable the kubelet service
* If it's a master node:
  * Remove persisted information
    * etcd store
    * PKI secrets
  * Remove etcd member from the etcd cluster
  * Remove the endpoint from the `kubeadm-config` config map
* Remove node from the cluster

For removing a node you only need to provide the name of the node known to Kubernetes:

```
skuba node remove my-worker
```

Or, if you want to remove a master node:

```
skuba node remove second-master
```

## kubectl-caasp

This project also comes with a kubectl plugin that has the same layout as `skuba`. You can
call to the same commands presented in `skuba` as `kubectl caasp` when installing the
`kubectl-caasp` binary in your path.

The purpose of the tool is to provide a quick way to see if nodes have pending
upgrades. The tool is currently returning fake data.

```
$ kubectl caasp cluster status
NAME      OS-IMAGE                              KERNEL-VERSION           KUBELET-VERSION   CONTAINER-RUNTIME   HAS-UPDATES   HAS-DISRUPTIVE-UPDATES   CAASP-RELEASE-VERSION
master0   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
master1   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
master2   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
worker0   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
worker1   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
worker2   SUSE Linux Enterprise Server 15 SP1   4.12.14-197.29-default   v1.16.2           cri-o://1.16.0      no            no                       4.1.0
```

## Demo

This is a quick screencast showing how it's easy to deploy a multi master node
on top of AWS. The procedure is the same as the deployment on OpenStack or on
libvirt.

The deployment is done on AWS via the terraform files shared inside of the `infra`
repository.

Videos:

  * [infrastructure creation](https://asciinema.org/a/wy9bqNjzszRN030sUIGM7f9j6)
  * [cluster creation](https://asciinema.org/a/PjblNTwwx0Z7ujyQPEu8SNHgF)

The videos are uncut, as you will see the whole deployment takes around 7 minutes:
4 minutes for the infrastructure, 3 minutes for the actual cluster.

The demo uses a small script to automate the sequential invocations of `skuba`.
Anything can be used to do that, including bash.
