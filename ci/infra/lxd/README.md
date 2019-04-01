## Introduction

Terraform cluster definition leveraging the libvirt provider.

The whole infra is based on openSUSE Leap 15.0 built for public cloud usage.

The image customization relies on cloud-init.

## Pre-requisites

* _LXD_

  The easiest way to install LXD is with a Snap: https://snapcraft.io/lxd.
  Just do a `snap install lxd`. Then you will have to add your username to
  the `lxc` group (for accessing the LXD socket without being root).

* _terraform/LXD_

  You whill have to compile the LXD provider by yourself with a Golang compiler
  (and your `GOPATH` properly set).
  Do a `go get -v -u github.com/sl1pm4t/terraform-provider-lxd`.
  Maybe you will have to add the provider to your `~/.terraformrc` if terraform does not find
  the provider automatically. For example:
  ```
  providers {
    lxd = "/users/me/go/src/github.com/sl1pm4t/terraform-provider-lxd/terraform-provider-lxd"
  }
  ```

## Machine access

By default all the machines will have the following users:

* All the instances have a `root` user with `linux` password.
* It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

## Load balancer

Terraform will create a static DHCP configuration to be used.

## Topology

The cluster will be made by these machines:

  * A load balancer
  * X master nodes: have `kubeadm`, `kubelet` and `kubectl` preinstalled
  * Y worker nodes: have `kubeadm`, `kubelet` and `kubectl` preinstalled

All node should be able to ping each other and resolve their FQDN.
