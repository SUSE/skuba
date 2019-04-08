# End to End tests for caaspctl

This tests are BDD style, using ginkgo Kubernetes framework for doing specific `caaspctl` e2e tests.

# Prerequisites for e2e tests.

1) Having 4 Instances: (1 load-balancer, 1 master, 2 workers).
   This Instances can be 

# Run e2e-tests

`make test-e2e`

# Architecture and design:

This testsuite can be executed indipendently and consumed from each different tests/ci frameworks.

A testsuite is a subdirectory of `tests` and is a indipendent unit. Examples: `tests/cluster-scale`, `tests/cilium`.

Each subdirectory, testsuite will contain tests. E.g the `tests/cluster-scale` can contain several BDD tests which (growth, node removal etc).

The testsuite shares only the `lib` directory which are utilty. 
The Common library is stored on `lib` directory, You should try to put code there to make clean the specs.

You need only pass the IP you can run the tests to any deployed cluster outside in the wild.
Alls hosts/vms should have sshd enabled on port 22. We use `linux` as std password but you can change it with the ENV.variable.

# Developing New Tests:

## Tests requirements:

0) All tests should be idempotent, meanining you can run them XX times, you will have the same results.

1) All tests can be run in parallel.

2) All tests doesn't require or have dependencies each others. Meaining: we can change order in which tests are executed, results will be the same. There is no hidden dependency between tests.

## How to create a new suite:

Generally we should avoid to create much subsuites if they are not needed. 

0) Create a dir like `your_suite_name`
1) Create a pkg accordingly inside the dir. This pkg should be empty, only containing `pkg services` as example.
2) Use `ginkgo bootstrap` for createing the `testsuite` file
3) Use `ginkgo generate name_test` for generating specs. 

See upstream doc for further details.
