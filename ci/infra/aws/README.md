## Introduction

This terraform project creates the infrastructure needed to run a
cluster on top of AWS using EC2 instances.

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `ec2-user` user. The `ec2-user` user can
perform `sudo` without specifying a password.

## Load balancer

The deployment will also create a AWS load balancer sitting in front of the
kubernetes API server. This is the control plane FQDN to use when defining
the cluster.

## Starting the cluster

### Configuration

You can use the `terraform.tfvars.example` as a template for configuring
the Terraform variables, or create your own file like `my-cluster.auto.tfvars`:

```sh
# customize any variables defined in variables.tf
stack_name = "my-k8s-cluster"

access_key = "<KEY>"

secret_key = "<SECRET>"

authorized_keys = ["ssh-rsa AAAAB3NzaC1y..."]
```

### Creating the infrastructure

You can create the infrastructure by _applying_ the script with:

Then you can use `terraform` to deploy the cluster

```sh
$ terraform init
$ terraform apply
```

Alternatively, you could pass some of the configuration from environment
variables on the command line, like this:

```sh
$ TF_VAR_authorized_keys=\[\"`cat ~/.ssh/id_rsa.pub`\"\] terraform plan
```

### Creating the Kubernetes cluster

Once the infrastructure has been created, you can obtain the details with
`terraform output`:

```console
$ terraform output
control_plane.private_dns = [
    ip-10-1-1-55.eu-central-1.compute.internal
]
control_plane.public_ip = [
    3.121.219.168
]
elb_address = k8s-elb-1487845812.eu-central-1.elb.amazonaws.com
nodes.private_dns = [
    ip-10-1-1-157.eu-central-1.compute.internal
]
nodes.public_ip = [
    54.93.246.74
]
```

Then you can initialize the cluster with `skuba cluster init`, using the Load Balancer (`lb_dns_name` in the Terraform output) as the control plane endpoint:

```console
$ skuba cluster init --control-plane k8s-elb-1487845812.eu-central-1.elb.amazonaws.com  my-devenv-cluster
** This is a BETA release and NOT intended for production usage. **
[init] configuration files written to /home/user/my-devenv-cluster
```

At this point we can bootstrap the first master:

```console
$ cd my-devenv-cluster
$ ./skuba node bootstrap --target 3.121.219.168 --sudo --user ec2-user ip-10-1-1-55.eu-central-1.compute.internal
```

And the  you can add a worker node with:

```console
$ cd my-devenv-cluster
$ ./skuba node join --role worker --target 54.93.246.74 --sudo --user ec2-user ip-10-1-1-157.eu-central-1.compute.internal
```

### Using the cluster

You must first point the `KUBECONFIG` environment variable to the `admin.conf`
file created in the cluster configuration directory:

```console
$ export KUBECONFIG=/home/user/my-devenv-cluster/admin.conf
```

And then you are ready for running some `kubectl` command like:

```console
$ kubectl get nodes
```

## Enable Cloud provider Interface

### Requirements

Before proceeding you must have created IAM policies matching the ones described
[here](https://github.com/kubernetes/cloud-provider-aws#iam-policy), one
for the master nodes and one for the worker nodes.

Once this is done you have to specify their name inside of the following
terraform variables:

  * `iam_profile_master`
  * `iam_profile_worker`

**Note well:** this must be done before the infrastructure is created.

### Cluster creation

At the time of writing the `skuba cluster bootstrap` command cannot yet create
cluster definitions that handle the cloud provider interface enablement.

You can however enable that manually:

  * Run `skuba cluster init` in the usual way
  * Edit the contents of `kubeadm-init.conf`
  * Edit the contents of `kubeadm-join.conf.d/master.conf.template`
  * Edit the contents of `kubeadm-join.conf.d/worker.conf.template`

The `kubeadm-init.conf` must be changed to include these sections:

```yaml
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
apiServer:
  extraArgs:
    cloud-provider: "aws"
controllerManager:
  extraArgs:
    cloud-provider: "aws"
    allocate-node-cidrs: "false"
```

The `kubeadm-join.conf.d/master.conf.template` and
the `kubeadm-join.conf.d/worker.conf.template` files must be changed to
include these sections:

```yaml
apiVersion: kubeadm.k8s.io/v1beta1
kind: JoinConfiguration
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
```

**Note well:** node must be bootstrapped/joined using their FQDN in order to
have the CPI find them. For example:

```
$ skuba node bootstrap -u ec2-user -s -t ip-172-28-1-225.eu-central-1.compute.internal ip-172-28-1-225.eu-central-1.compute.internal
$ skuba node join --role worker -u ec2-user -s -t ip-172-28-1-15.eu-central-1.compute.internal ip-172-28-1-15.eu-central-1.compute.internal
```

Refer to the [AWS Cloud Provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#aws)
documentation for details on how to use these features in your cluster.

## Known limitations

### IP addresses

`skuba` cannot currently access nodes through a bastion host, so all
the nodes in the cluster must be directly reachable from the machine where
`skuba` is being run. We must also consider that `skuba` must use
the external IPs as `--target`s when initializing or joining the cluster,
while we must specify the internal DNS names for registering the nodes
in the cluster.


