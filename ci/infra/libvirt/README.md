## Introduction

Terraform cluster definition leveraging the libvirt provider.

The whole infra is based on openSUSE Leap 15.0 built for public cloud usage.

The image customization relies on cloud-init.

## Machine access

By default all the machines will have the following users:

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `opensuse` user with `linux` password. The `opensuse` user can
perform `sudo` without specifying a password.

## Load balancer

Terraform will create a static DHCP configuration to be used.

The load balancer will be named `ag-lb.ag-test.net` and will always have the
`10.17.1.0` IP address.

## Topology

The cluster will be made by these machines:

  * Load balancer
  * X master nodes: have `kubeadm`, `kubelet` and `kubectl` preinstalled
  * Y worker nodes: have `kubeadm`, `kubelet` and `kubectl` preinstalled

The master nodes will be named `ag-master-{N}.ag-test.net` and will always have the
`10.17.2.{0,1,...}` IP addresses.

The worker nodes will be named `ag-worker-{N}.ag-test.net` and will always have the
`10.17.3.{0,1,...}` IP addresses.

All the nodes can ping each other and can resolve their FQDN.


# PRO-tip

Download the image referenced inside of the checkout of this repository:

```
$ wget `grep qcow2 variables.tf | awk {'print $3'} | sed -e 's/"//g'`
```

Then `cp terraform.tfvars.example terraform.tfvars` and edit it to reference
the image you just downloaded.
