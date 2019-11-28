import os

import pytest

FILEPATH = os.path.realpath(__file__)
TESTRUNNER_DIR = os.path.dirname(os.path.dirname(FILEPATH))


class PyTestOpts:

    NO_CAPTURE_LOGS = "--show-capture=no"

    SHOW_OUTPUT = "-s"

    VERBOSE = "-v"

    COLLECT_TESTS = "--collect-only"


class TestDriver:
    def __init__(self, conf, platform):
        self.conf = conf
        self.platform = platform

    def run(self, module=None, test_suite=None,
            test=None, verbose=False, collect=False,
            skip_setup=None, mark=None, junit=None):
        opts = []

        vars_opt = "--vars={}".format(self.conf.yaml_path)
        opts.append(vars_opt)

        platform_opt = "--platform={}".format(self.platform)
        opts.append(platform_opt)

        if verbose:
            opts.append(PyTestOpts.SHOW_OUTPUT)

        # Dont capture logs
        opts.append(PyTestOpts.NO_CAPTURE_LOGS)

        # generete detailed test results
        opts.append(PyTestOpts.VERBOSE)

        if collect:
            opts.append(PyTestOpts.COLLECT_TESTS)

        if skip_setup is not None:
            opts.append(f"--skip-setup={skip_setup}")

        if junit is not None:
            opts.append(f"--junitxml={TESTRUNNER_DIR}/{junit}.xml")

        if mark is not None:
            opts.append(f'-m {mark}')

        test_path = module if module is not None else "tests"

        if test_suite:
            if not test_suite.endswith(".py"):
                raise ValueError("Test suite must be a python file")
            test_path = os.path.join(test_path, test_suite)

        if test:
            if not test_suite:
                raise ValueError("Test suite is required for selecting a test")
            test_path = "{}::{}".format(test_path, test)

        # Path must be the last argument
        opts.append(test_path)

        # Before running the tests, switch to the directory of the testrunner.py
        os.chdir(TESTRUNNER_DIR)

        result = pytest.main(args=opts)

        if not junit and result > 0:
            raise AssertionError("Running {} failed.\nExit Code: {}".format(test_path, result))
