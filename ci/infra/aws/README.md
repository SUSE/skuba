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

## Known limitations

### IP addresses

`skuba` cannot currently access nodes through a bastion host, so all
the nodes in the cluster must be directly reachable from the machine where
`skuba` is being run. We must also consider that `skuba` must use
the external IPs as `--target`s when initializing or joining the cluster,
while we must specify the internal DNS names for registering the nodes
in the cluster.

### Kubernetes cloud provider

The kubernetes cluster deployed with `skuba` does not use the Kubernetes
cloud provider for AWS. In order to be enable this feature, users need to
modify the cluster configuration generated with `skuba cluster init`.

```yaml
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
nodeRegistration:
  kubeletExtraArgs:
    cloud-provider: "aws"
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
kubernetesVersion: v1.13.0
apiServer:
  extraArgs:
    cloud-provider: "aws"
controllerManager:
  extraArgs:
    cloud-provider: "aws"
```

Refer to the [AWS Cloud Provider](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#aws)
documentation for details on how to use these features in your cluster.

