## Introduction

These terraform definitions are going to create the whole
cluster on top of openstack.

## Deployment

Make sure to download an openrc file from your OpenStack instance, e.g.:

`https://engcloud.prv.suse.net/project/api_access/openrc/`

and source it:

```sh
source container-openrc.sh
```

Also make sure to have your ssh key within OpenStack, by adding your key to the
key_pairs first.

Then you can use `terraform` to deploy the cluster

```sh
terraform init
terraform apply
```

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `opensuse` user. The `opensuse` user can
perform `sudo` without specifying a password.

## Load balancer

The kubernetes api-server instances running inside of the cluster are
exposed by a load balancer managed by OpenStack.

## Customization

Copy the `terraform.tfvars.example` to `terraform.tfvars` and
provide reasonable values.

## Variables

`caasp_registry_code` - Provide SUSE CaaSP Product Registration Code in 
`registration.auto.tfvars` file to register product against official repositories
