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

## License

skuba-update is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/SUSE/skuba-update/blob/master/LICENSE) for the full
license text.
