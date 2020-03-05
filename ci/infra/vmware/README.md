## Introduction

These terraform definitions are going to create the CaaSP v4 cluster on top of VMWare vSphere cluster.

This code was developed and tested on VMware vSphere cluster based on VMware ESXi 6.7.20000.

## Deployment

Prepare a VM template machine in vSphere by following [vmware-deployment guide](https://susedoc.github.io/doc-caasp/master/caasp-deployment/single-html/#_vm_preparation_for_creating_a_template).

It doesn't matter if you deploy the VM template for SLES15-SP1 manually by using ISO or you use pregenerated vmdk SLES15-SP1 JeOS image but in both cases you'll need `cloud-init-vmware-guestinfo` package (from SUSE CaaS Platform module), `cloud-init` package (from Public Cloud Module) and its dependent packages installed. The respective services must be enabled:

```sh
systemctl enable cloud-init cloud-init-local cloud-config cloud-final
```

Next you need to define following environment variables in your current shell with proper value:

```sh
# HINT: Please enter just a hostname without specifing a protocol in VSPHERE_SERVER variable (using https by default).
export VSPHERE_SERVER="vsphere.cluster.endpoint.hostname"
export VSPHERE_USER="username"
export VSPHERE_PASSWORD="password"
export VSPHERE_ALLOW_UNVERIFIED_SSL="true"
```

Once you perform a [Customization](#Customization) you can use `terraform` to deploy the cluster:

```sh
terraform init
terraform validate
terraform apply
```

## Machine access

It is important to have your public ssh key within the `authorized_keys`, this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `sles` user, password is not set. User can login only as `sles` user over SSH by using his private ssh key. The `sles` user can perform `sudo` without specifying a password.

## Load balancer

VMWare vSPhere doesn't offer a load-balancer solution. Please expose port 6443 for the Kubernetes api-servers on the master nodes on a local load-balancer using round-robin 1:1 port forwarding.

NOTE: Development version of these VMWare Terraform definitions will deploy preconfigured load-balancer VM node which is using haproxy software. Use its IP address in `skuba cluster init --control-plane <ip-load-balancer> <cluster-name>` command. For accessing haproxy statistics open http://ip-load-balancer:9000/stats in your browser.

## Customization

IMPORTANT: Please define unique `stack_name` value in `terrafrom.tfvars` file to not interfere with other deployments.

Copy the `terraform.tfvars.example` to `terraform.tfvars` and provide reasonable values.

## Variables

`vsphere_datastore` - Provide the datastore to use in vSphere\
`vsphere_datacenter` - Provide the datacenter to use in vSphere\
`vsphere_datastore_cluster` - Provide the datastore cluster to use on the vSphere server\
`vsphere_network` - Provide the network to use in vSphere - this network must be able to access the ntp servers and the nodes must be able to reach each other\
`vsphere_resource_pool` - Provide the resource pool the machines will be running in\
`template_name` - The template name the machines will be copied from\
`firmware` - Replace the default "bios" value with "efi" in case your template was created by using EFI firmware\
`stack_name` - Identifier to make all your resources unique and avoid clashes with other users of this terraform project\
`authorized_keys` - A list of ssh public keys that will be installed on all nodes\
`repositories` - Additional repositories that will be added on all nodes\
`packages` - Additional packages that will be installed on all nodes

### Please use one of the following options:

`caasp_registry_code` - Provide SUSE CaaSP Product Registration Code in `registration.auto.tfvars` file to register product against official SCC server\
`rmt_server_name` - Provide SUSE Repository Mirroring Tool Server Name in `registration.auto.tfvars` file to use repositories stored on RMT server
