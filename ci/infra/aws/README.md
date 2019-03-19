## Introduction

This terraform project creates a the whole infrastructure needed to run a
vNext cluster on top of AWS using EC2 instances.

The deployment will create a VPC spreading over three availability zones.
The nodes (master and workers) will spread over the availability zones.

For example, a cluster with 3 master nodes will have each one of them in a
different availability zone. A cluster with 5 master nodes will have 2 nodes into
the first and second availability zones, while the third one will have only 1
master node.

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `ec2-user` user. The `ec2-user` user can
perform `sudo` without specifying a password.

## Load balancer

The deployment will also create a AWS load balancer sitting in front of the
kubernetes API server. This is the control plane FQDN to use when defining
the cluster.

## Customization

Copy the `terraform.tfvars.example` to `terraform.tfvars` and
provide reasonable values.
