## Introduction

This terraform project creates the infrastructure needed to run a
cluster on top of AWS using EC2 instances.

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `ec2-user` user. The `ec2-user` user can
perform `sudo` without specifying a password.

Only the master nodes have a public IP associated with. All the worker nodes
are located on a private subnet.

The network structure resembles the one describe inside of this
[AWS document](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html).

## Load balancer

The deployment will also create a AWS load balancer sitting in front of the
kubernetes API server. This is the control plane FQDN to use when defining
the cluster.

## Starting the cluster

### Credentials

The three following arguments must be provided:

* AWS_ACCESS_KEY_ID: This is the AWS access key.
* AWS_SECRET_ACCESS_KEY This is the AWS secret key.
* AWS_DEFAULT_REGION This is the AWS region. A list of region names
can be found [here](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions)

To do so, source the following variables,
for security reasons turn off bash history:

```sh
set +o history
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_DEFAULT_REGION="eu-central-1"
set -o history
```

It can also be stored in a file, for example `aws-credentials`

```
AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
AWS_DEFAULT_REGION="eu-central-1"
```

and sourced:

```sh
set -a; source aws-credentials; set +a
```

### Configuration

You can use the `terraform.tfvars.example` as a template for configuring
the Terraform variables, or create your own file like `my-cluster.auto.tfvars`:

```sh
# customize any variables defined in variables.tf
stack_name = "my-k8s-cluster"

authorized_keys = ["ssh-rsa AAAAB3NzaC1y..."]
```

The terraform files will create a new dedicated VPC for the kubernetes cluster.
It's possible to join this VPC with other existing ones by specifying the IDs
of the VPC to join inside of the `peer_vpc_ids` variable.

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
control_plane.private_dns = {
    i-1234567890abcdef0 = ip-10-1-1-55.eu-central-1.compute.internal
}
control_plane.public_ip = {
    i-1234567890abcdef1 = 3.121.219.168
}
elb_address = k8s-elb-1487845812.eu-central-1.elb.amazonaws.com
nodes.private_dns = {
    i-1234567890abcdef2 = ip-10-1-1-157.eu-central-1.compute.internal
}
```

Then you can initialize the cluster with `skuba cluster init`, using the Load Balancer (`elb_address` in the Terraform output) as the control plane endpoint:

```console
$ skuba cluster init --control-plane k8s-elb-1487845812.eu-central-1.elb.amazonaws.com --cloud-provider aws my-devenv-cluster
** This is a BETA release and NOT intended for production usage. **
[init] configuration files written to /home/user/my-devenv-cluster
```

At this point we can bootstrap the first master:

```console
$ cd my-devenv-cluster
$ skuba node bootstrap --user ec2-user --sudo --target ip-10-1-1-55.eu-central-1.compute.internal ip-10-1-1-55.eu-central-1.compute.internal
```

And the  you can add a worker node with:

```console
$ cd my-devenv-cluster
$ skuba node join --role worker --user ec2-user --sudo --target ip-10-1-1-157.eu-central-1.compute.internal ip-10-1-1-157.eu-central-1.compute.internal
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

By default these terraform files **do not** create these policies. This is not
done because some corporate users are entitled to create AWS resources, but
due to security reasons, they do not have the privileges to create new IAM
policies.

If you do not have the privileges to create IAM policies you have to request
to your organization the creation of such policies.
Once this is done you have to specify their name inside of the following
terraform variables:

  * `iam_profile_master`
  * `iam_profile_worker`

**Note well:** this must be done before the infrastructure is created.


On the other hand, if you have the privileges to create IAM policies you can
let these terraform files take care of that for you by doing these operations.
This is done automatically by leaving the `iam_profile_master` and
`iam_profile_worker` variables unspecified.

### Cluster creation

The cloud provider integration must be enabled when creating the cluster
definition:

```
skuba cluster init --control-plane <ELB created by terraform> --cloud-provider aws my-cluster
```

**WARNING:** nodes must be bootstrapped/joined using their FQDN in order to
have the CPI find them. For example:

```
$ skuba node bootstrap -u ec2-user -s -t ip-172-28-1-225.eu-central-1.compute.internal ip-172-28-1-225.eu-central-1.compute.internal
$ skuba node join --role worker -u ec2-user -s -t ip-172-28-1-15.eu-central-1.compute.internal ip-172-28-1-15.eu-central-1.compute.internal
...
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

### Availability zones

Right now all the nodes are created inside of the same availability zone.

It is possible to filter the available AZ by configuring `availability_zones_filter`.

The available filters can be found [here in the AWS API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAvailabilityZones.html)
