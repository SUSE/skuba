# skuba-update [![Build Status](https://travis-ci.org/SUSE/skuba-update.svg?branch=master)](https://travis-ci.org/SUSE/skuba-update)

Utility to automatically refresh and update a SUSE CaaSP cluster. As it is
right now, it only works for the zypper package manager.

## Requirements

- zypper 1.14.0 or higher
- python 3.6 or higher

## Development

This is a python project that makes use of setuptools and virtual environment.

To set the development environment consider the following commands:

```bash
# Get into the repository folder
$ cd skuba-update

# Initiate the python3 virtualenv
$ python3 -m venv .env3

# Activate the virutalenv
$ source .env3/bin/activate

# Install development dependencies
$ pip install -r dev-requirements.txt

# We need the shellcheck utility. You can install it from the package manager
# you might be using.
$ sudo zypper install ShellCheck

# Run tests and code style checks
$ make test

# Run tests and code style checks inside of a Docker container
$ DOCKERIZED_UNIT_TESTS=1 make test
```

## Running OS tests

Besides unit tests, there are also OS tests that check that the zypper commands
we use underneath work as expected. You can consider them as integration tests
for core features. You can run these tests like so:

```bash
$ cd test/os/suse
$ make test
# You can also run specific OSes like:
$ make test-sle
```

Moreover, you can also filter the tests that you want to run by using two
environment variables:

- `TESTNAME`: specify the name of the file (without the extension) that you want
  to execute.
- `SKUBA`: if it is not set (default), the same test will be run once by using
  zypper commands, and another time by using a `skuba-update` binary. If it's
  set to `1`, then it will only run the binary version. Otherwise, if it's set
  to `0`, it will only run once (with zypper commands).

An example usage:

```bash
# Run only the 'test-non-interruptive-updates' test, and use SLE.
$ TESTNAME=test-non-interruptive-updates make test-sle

# Same as before but only run with the 'skuba-update' binary.
$ TESTNAME=test-non-interruptive-updates SKUBA=1 make test-sle
```

## License

skuba-update is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/SUSE/skuba-update/blob/master/LICENSE) for the full
license text.
