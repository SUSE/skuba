# CI tasks:

The main goal of CI tasks is to execute small automation tasks, which will be called by makefile targets.
CI tasks should be used as bridge between local development and remote CI.

# e2e-tests

This script is used in CI also for local development.

HOwto use it:

if you have already deployed your cluster, you can either pass individual IPS or read the terraform tfstate
