# End to End tests for caaspctl

The test can be run locally or in CI.

# Requirements:

- the infrastructure is already deployed. 

# HOW TO RUN:

You can run the `e2e-tests` in 2 ways:

1) with IP adress with the supported env. variables (see later)

`CONTROLPLANE=10.86.2.23 MASTER00=10.86.1.103 WORKER00=10.86.2.130 make test-e2e` 

2) without IP but with a terraform state,  specify the platform you have deployed the infrastructure.

`IP_FROM_TF_STATE=TRUE PLATFORM=openstack make test-e2e`

This method will read the tfstate file and read the IPs of host and pass them to `ginkgo`

Boths methods are convenients: 1) method is usefull when we don't have the terraform state.

3) Use a custom ginkgo binary:

`~/go/src/github.com/SUSE/caaspctl> GINKGO_BIN_PATH="$PWD/ginkgo" IP_FROM_TF_STATE=TRUE PLATFORM=openstack make test-e2e`
In the following example we assume you have builded ginkgo from vendor.

# Env. Variable:

## Guidelines:

Syntax: use `_` underscores for separating words and use UPPERCASE for naming.

Adding NEW variables:

- IPs of nodes when needed can be added in the python and golang code.

In general adding new behaviour variable should be discussed within the team, since we need to keep the variable minimalist as possible.

All needed variable should have been already implemented, only the NODE ips variable should be added. 

### Currenlty supported:

### IPs:

- CONTROLPLANE = IP of host which will be the controlplane
- MASTER00 = IP of 1st master
- WORKER00 = IP of 1st worker

SEE 1) example in HOW TO RUN

As showed, in future we will have `WORKER01`, `MASTER01`, `MASTER02` etc. 

### Behaviour variables:

Read from a tfstate file, you need both variable passed

- `IP_FROM_TF_STATE`: if set to `TRUE` this will read terraform states.
- `PLATFORM`: this specify the provider used. (libvirt, openstack, vmware, etc)

SEE 2) example in HOW TO RUN

## Binary Localtion (optional)

- `CAASPCTL_BIN_PATH` for specify the full path of a caaspctl binary. ( e.g if you use an RPM)
- `GINKGO_BIN_PATH` use this var for passing a fullpath to a ginkgo bin which will be used by tests.

By default this variable point to GOPATH for devel purposes.

# Internal Development:

The Env. variables are set by `ci/tasks/e2e-tests.py`.

This script pass ENV. Variables to ginkgo and call the ginkgo binary which will run the e2e tests.

In this way we can run easy tests locally or in remote CI
