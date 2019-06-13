## Introduction

These terraform definitions are going to create the whole
cluster on KVM via terraform-provider-libvirt.

## Prerequisites

Follow instructions at https://github.com/dmacvicar/terraform-provider-libvirt#installing to install terraform-provider-libvirt.

## Deployment

Use `terraform` to deploy the cluster. `-parallelism=1` used in apply command avoids potential concurrent issues in terraform-provider-libvirt.

```sh
terraform init
terraform apply -parallelism=1
```

## Machine access

It is important to have your public ssh key within the `authorized_keys`, this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `sles` user, password is not set. User can login only as `sles` user over SSH by using his private ssh key. The `sles` user can perform `sudo` without specifying a password.

## Load balancer

The kubernetes api-server instances running inside of the cluster are
exposed by a load balancer managed by OpenStack.

## Customization

IMPORTANT: Please define unique `stack_name` value in `terrafrom.tfvars` file to not interfere with other deployments.

Copy the `terraform.tfvars.example` to `terraform.tfvars` and provide reasonable values.

## Variables

`image_uri` - URL of the image to use
`stack_name` - Identifier to make all your resources unique and avoid clashes with other users of this terraform project
`authorized_keys` - A list of ssh public keys that will be installed on all nodes
`repositories` - Additional repositories that will be added on all nodes
`packages` - Additional packages that will be installed on all nodes

### Please use one of the following options:
`caasp_registry_code` - Provide SUSE CaaSP Product Registration Code in `registration.auto.tfvars` file to register product against official SCC server  
`rmt_server_name` - Provide SUSE Repository Mirroring Tool Server Name in `registration.auto.tfvars` file to use repositories stored on RMT server  
