# Addons

Addons extend the functionality of kubernetes. You can think of them as
plugins that hook into Kubernetes providing new infrastructure services,
for example, network connectivity among pods.

## Usage

All the go files in this directory (apart from the xxx_test.go files)
are equipped with the functions and variables to deploy and configure
the addons correctly.

All addons in this directory get deployed automatically by Skuba. They
get registered through the init() function, which is called as soon as
another package imports the addons package. To learn more about this go
functionality read:

https://golang.org/doc/effective_go.html#init

## How to create a new Addon

To create a new addon, check the rest of the addons source file to
understand what is required. You should express the version mapping
with kubernetes in:

https://github.com/SUSE/skuba/blob/master/internal/pkg/skuba/kubernetes/versions.go

Each addon must also declare its `AddOnType`. Any number of addons of type
`CniAddOn`can be created. However, only a single addon of type `CniAddOn`
can be used in the cluster and as such only one addon providing a CNI
plugin will be activated and deployed.

The CNI addon that is activated defaults to cilium, but can be toggled by
passing `--cni-plugin` to `skuba cluster init`.

Example:

```sh
skuba cluster init --control-plane load-balancer.example.com --cni-plugin cilium company-cluster
```
