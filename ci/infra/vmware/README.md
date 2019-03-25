# Vmware Terraform deployments:

Check out the wikipages about the vmware infra. server.

The `cluster.tf` will deploy a worker/master cluster configurable with number of workers/master ( see `variable.tf`)

The image customization relies on cloud-init.

## Machine access

By default all the machines will have the following users:

It is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

All the instances have a `root` and a `opensuse` user with `linux` password. The `opensuse` user can
perform `sudo` without specifying a password.

## Load balancer

Not yet implemented.
