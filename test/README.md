# End to End tests for caaspctl

This tests are BDD style, using ginkgo Kubernetes framework for doing specific `caaspctl` e2e tests.

This doc is about the demo.


# Prerequisites for e2e tests.

1) Having 4 Instances: (1 load-balancer, 1 master, 2 workers).
   This variables are passed via ENV. see `vim ../ci/tasks/e2e-tests.py`

# Run e2e-tests

`make test-e2e`
`DEBUG=True make test-e2e`

## Architecture and design:

This testsuite can be executed indipendently and consumed from each different tests/ci frameworks.

A testsuite is a subdirectory of `tests` and is a indipendent unit. Examples: `tests/cluster-scale`, `tests/cilium`.

Each subdirectory, testsuite will contain tests. E.g the `tests/cluster-scale` can contain several BDD tests which (growth, node removal etc).

The testsuite shares only the `lib` directory which are utilty. 
The Common library is stored on `lib` directory, You should try to put code there to make clean the specs.

You need only pass the IP you can run the tests to any deployed cluster outside in the wild.
Alls hosts/vms should have sshd enabled on port 22. We use `linux` as std password but you can change it with the ENV.variable.
