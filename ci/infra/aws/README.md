## Introduction

This terraform project creates the infrastructure needed to run a
vNext cluster on top of AWS using EC2 instances.

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

You can add the access and secret keys in Terraform configuration file like `my-cluster.auto.tfvars`:

```sh
# customize any variables defined in variables.tf
stack_name = "my-k8s-cluster"

access_key = "<KEY>"

secret_key = "<SECRET>"

authorized_keys = ["ssh-rsa AAAAB3NzaC1y..."]
```

### Creating the infrastructure

You can create the infrastructure byy _applying_ the script with:

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

```sh
$ terraform output
ip_masters = [
    35.180.108.76
]
ip_workers = [
    35.181.119.73
]
lb_dns_name = my-k8s-cluster-kube-lb-5809f2c4a03cd884.elb.eu-west-3.amazonaws.com
```

Then you can initialize the cluster with `caaspctl cluster init`, using the Load Balancer (`lb_dns_name` in the Terraform output) as the control plane endpoint:

```
$ caaspctl cluster init --control-plane my-k8s-cluster-kube-lb-5809f2c4a03cd884.elb.eu-west-3.amazonaws.com  my-devenv-cluster
** This is a BETA release and NOT intended for production usage. **
[init] configuration files written to /home/user/my-devenv-cluster
```

At this point we can bootstrap the first master:

```sh
$ cd my-devenv-cluster
$ ./caaspctl node bootstrap --target 35.180.108.76 --sudo --user ec2-user --ignore-preflight-errors NumCPU ec2-35-180-108-76.eu-west-3.compute.amazonaws.com
```

### Using the cluster

```sh
$ kubectl cluster-info
```

// TODO

## Known limitations

### IP addresses

`caaspctl` cannot currently access nodes through a bastion host, so all
the nodes in the cluster must be directly reachable from the machine where
`caaspctl` is being run. We must also consider that `caaspctl` cannot
currently differentiate between external IPs and internal ones when
configuring the Kubernetes components.
All these things mean that you can either:

1) assign Elastic IPs to all the nodes in the cluster and use them
as the node's IPs, providing accessibility from the machine where 
`caaspctl` is being run as well as accessibility between the Kubernetes
components (although machines)

2) alternatively, you could run `caaspctl` in an instance in the same
subnet as the nodes being created. That would mean you would have to:
    * create or reuse a VPC with a valid subnet
    * create or reuse a EC2 instance that is connected to that subnet.
    * copy the `caaspctl` binary to this instance and run it there.

### Kubernetes cloud provider

The kubernetes cluster deployed with `caaspctl` does not use the Kubernetes
cloud provider for AWS. In order to be enble this feature, users need to
modify the cluster configuration generated with `caaspctl cluster init`.

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
