# infrastructure

Infrastructure deployment scripts.

## Supported platforms

This project currently supports three different platforms to deploy on top of:

* [OpenStack](openstack/README.md)
* [VMware](vmware/README.md)
* [Amazon Web Services](aws/README.md)
* [libvirt](libvirt/README.md)

For all of them it is important to have your public ssh key within the `authorized_keys`,
this is done by `cloud-init` through a terraform variable called `authorized_keys`.

Please, refer to the specific README.md files inside each directory for further
information.
