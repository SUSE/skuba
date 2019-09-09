# Writing tests

`testrunner` offers the `test` command which allows running tests using `testrunner`'s functionality for deploying infrastructure and executing `skuba` commands.

Tests are based on [pytest](https://docs.pytest.org) framework and take advantage of features such as [`fixtures`](https://docs.pytest.org/en/latest/fixture.html) to facilitate test setup and tear down.

Following pytest's standard test organization, tests must be defined in python files with a name following the pattern `xxxx_test.py`, where `xxx` is the name of the test suite. Each test is defined in the test file as an individual function or as a function in a class. Test functions must follow the name convention `test_xxxx` where `xxx` is the name of the test.

See the following example:

```
def test_add_worker(bootstrap, skuba):
    skuba.node_join(role="worker", nr=0)
    masters = skuba.num_of_nodes("master")
    workers = skuba.num_of_nodes("worker")
    assert masters == 1
    assert workers == 1
```
Listing 1. Sample Test

## Using fixtures

You may have noticed in the example above the two parameters to the `test_add_worker`, `setup` and `skuba`. These are `fixtures'.

Testrunner provides the following fixtures:
- conf: an object with the configuration read from the `vars` file.
- platform: a Platform object
- skuba: an Skuba object configured
- target: the name of the target plaform

Tests can define and use additional fixtures, such as the `setup` fixture in the example above, which executes the initialziation of the cluster. When used for this purpose, one interesting feature is the definition of a finalizer function which is executed automatically when a test that uses this fixture ends, either successfully or due to an error.

The example below shows a fixture that provides a bootstrapped cluster. It also automatically cleans up the allocated resources by adding the `cleanup` function as a finalizer:

```
@pytest.fixture
def setup(request, platform, skuba):
    def cleanup():
        platform.cleanup()
    request.addfinalizer(cleanup)

    platform.provision()
    skuba.cluster_init()
    skuba.node_bootstrap()
```

Note: pytest also allow a more idiomatic way of defining teardown logic in fixtures by using python's `yield` statement instead of registering a finalizer, as shown in the code below. However, finalizer functions have the advantage that they will always be called regardless if the fixture setup code raises an exception, provided they are registered before the exception occurs. Therefore, **testrunner encourages using finalizer functions**.

```
@pytest.fixture
def setup(request, platform, skuba):
    platform.provision()
    skuba.cluster_init()
    skuba.node_bootstrap()
    yield               # return from fixture
    platform.cleanup()  # teardown logic
```

## Reusing already deployed infrastructure

Sometime, it is convenient to reuse an already deployed infrastructure when executing tests. This is a common case while tests are beeing developed (as they must be tested by the developer and errors need to be fixed), or when multiple tests which have no side effects can share the same infrastructure.

To address these uses cases, `testrunner`'s [`test` command](../README.md#test-command) provides the `--skip-setup` option which allows skipping the execution of one or more setup fixtures that setup, whithout having to modify the test or the fixtures. If a fixuture depends on other fixtures, those are also skipped automatically.

Consider the following fixtures:

```
@pytest.fixture()
def provision():
# provision infrastructure

@pytest.fixture()
def bootstrap(provision, skuba):
# bootstrap cluster

@pytest.fixture()
def deployment(bootstrap, platform, skuba)
# complete cluster deployment
# joining all nodes

def test_deployment(deployment, skuba:
# test fully deployed cluster

```

Running the following command will executed the test without executing the cluster deployment fixture and neither of the fixtures it depends (`bootstrap`, `provision`)

```
testrunner test --skip-setup deployed -t test_deployment
```


## Running tests with the Testrunner

The `testrunner` command can be used for running tests. It allows selecting a directory, an individual test file (a suite of tests) or an specific test in a test file.

Given the following directory structure:
```
testrunner
vars
 |-- vars.yaml
tests
 |-- test_workers.py

The command below will exectute the `test_add_worker` function defined in `tests/test_workers.py`:

```
testrunner -v vars/vars.yaml test --module tests --suite core_tests.py --test test_add_worker
```



## Using Testrunner library

Testrunner provides a library of functions that wraps `skuba` and `terraform` for executing actions such as provisioning a platform, or runnig any `skuba` command,

### Platform

`Platform` offers the functions required for provisioning a platform for deploying a cluster. Provides the following functions:
- `get_platform(conf)`: returns an instance of the platform initialized with the configuration passed in the `conf` parameter. This configuration can be obtained by means of the `conf` fixture
- `provision`:  executes the provisioning of the platform
- `cleanup`: releases any resource obtained by `provision`
- `get_nodes_ipaddrs(role)`: return the list of ip addresses for the nodes provisioned for a role.
- `get_lb_ipadd`: returns the ip address for the load balancer node

### Skuba

`Skuba` wraps the `skuba` commands:
- `Skuba(conf): creates an instance of the `Skuba` class initialized with the configuration provided in `conf`
- `cluster_init()`: initializes the skuba cluster configuration
- `node_bootstrap()`: bootstraps a cluster
- `node_join(role, nr)`: adds a new node to the cluster with the given role. The node is identified by its index in the provisioned nodes for that role.
- `node_remove(role, nr)` removes a node currently part of the cluster. The node is identified by its role an its id in the list of provisioned nodes for that role.
- `num_of_nodes(role)`: returns the number of nodes in cluster for the given role.
