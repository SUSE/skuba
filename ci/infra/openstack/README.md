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

Once you perform a [Customization](#Customization) you can use `terraform` to deploy the cluster:

```sh
terraform init
terraform validate
terraform apply
```

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and `sles` user. The normal 'sles' user user can
perform `sudo` without specifying a password.

Neither root nor the normal `sles` user will have password. Both `terraform` and `skuba`
are using SSH key-based authentication. You can always set a password after the
creation of the machines using `sudo passwd sles` (for normal user) or `sudo passwd` (for root).

## Load balancer

The kubernetes api-server instances running inside of the cluster are
exposed by a load balancer managed by OpenStack.

## Customization

Copy the `terraform.tfvars.example` to `terraform.tfvars` and
provide reasonable values.

## Variables

`image_name` - Name of the image to use\
`internal_net` - Name of the internal network to be created\
`stack_name` - Identifier to make all your resources unique and avoid clashes with other users of this terraform project\
`authorized_keys` - A list of ssh public keys that will be installed on all nodes\
`repositories` - Additional repositories that will be added on all nodes\
`packages` - Additional packages that will be installed on all nodes\
`enable_cpi` - Enable Cloud Provider Integration with OpenStack in `cpi.auto.tfvars`\
`ca_file` - Used to specify the path to your custom CA file in `cpi.auto.tfvars`

### Please use one of the following options:
`caasp_registry_code` - Provide SUSE CaaSP Product Registration Code in 
`registration.auto.tfvars` file to register product against official SCC server\
`rmt_server_name` - Provide SUSE Repository Mirroring Tool Server Name in
`registration.auto.tfvars` file to use repositories stored on RMT server
