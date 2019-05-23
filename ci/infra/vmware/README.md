## Introduction

These terraform definitions are going to create the whole cluster including load-balancer node/vm on top of VMWare vSphere cluster.

This code was developed and tested on VMware vSphere cluster based on VMware ESXi 6.7.20000.

## Deployment

Prepare a template machine in vSphere by following [vmware-deployment guide](https://susedoc.github.io/doc-caasp/adoc/caasp-deployment/single-html/#_vm_preparation_for_creating_a_template).

It doesn't matter if you deploy the vm template for SLES_SP1 manually by using ISO or you use pregenerated vmdk image SLES15_SP1 JeOS but in both cases you'll need `cloud-init` package installed and the respective services must be enabled:

```sh
systemctl enable cloud-init cloud-init-local cloud-config cloud-final
```

```sh
sed -i -e '/mount_default_fields/{adatasource_list: [ NoCloud, OpenStack, None ]
}' /etc/cloud/cloud.cfg
```

Next you need to define following environment variables in your current shell with proper value:

```sh
# HINT, please enter just a hostname without specifing a protocol in VSPHERE_SERVER variable (using https by default)
export VSPHERE_SERVER="vsphere.cluster.endpoint.hostname"
export VSPHERE_USER="an_user"
export VSPHERE_PASSWORD="passwd"
export VSPHERE_ALLOW_UNVERIFIED_SSL="true"
```

Then you can use `terraform` to deploy the cluster

```sh
terraform init
terraform apply
```

## Machine access

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `sles` user, password is not set. User can login only as `sles` user over SSH by using his private ssh key. The `sles` user can perform `sudo` without specifying a password.

## Load balancer

vSPhere doesn't offer a load-balancer solution by itself but this terraform code will deploy a basic load-balancer vm based on haproxy which will be configured to expose ports 6443 of api-severs on all master nodes by doing round-robin 1:1 portforwarding.

## Customization

IMPORTANT, please define unique `stack_name` value in `terrafrom.tfvars` file to not interfere with other deployments.

Copy the `terraform.tfvars.example` to `terraform.tfvars` and provide reasonable values.

## Variables

`vsphere_datastore` - Provide the datastore to use on the vSphere server
`vsphere_datacenter` - Provide the datacenter to use on the vSphere server
`vsphere_network` - Provide the network to use on the vSphere server - this network must be able to access the ntp servers and the nodes must be able to reach each other
`vsphere_resource_ppol` - Provide the resource pool the machines will be running in
`template_name` - The template name the machines will be copied from
`stack_name` - A prefix that all of the booted machines will use
`authorized_keys` - A list of ssh public keys that will be installed on all nodes
`repositories` - Additional repositories that will be added on all nodes
`packages` - Additional packages that will be installed on all nodes
`caasp_registry_code` - Provide SUSE CaaSP Product Registration Code in
`registration.auto.tfvars` file to register product against official repositories
