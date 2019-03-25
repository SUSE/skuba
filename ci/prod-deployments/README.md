# Deployments and Data of production CI

This directory is a place older for data used by Prod CI.

# Layout of director

All the platform contains `tfvars` file used for by CI Pipelines.

An execption is the baremetal directory which contains autoyast files used for deploying in CI.
( we cannot use terraform for baremetal).

## Naming

The naming schema is following:

`Product`-`CAASP-VERSION`-`CLUSTER-topology`-`remote-server`

example:
`SLES15-41devel-4master3worker-nue.tfvars`
