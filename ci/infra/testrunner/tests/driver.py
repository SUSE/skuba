import os

import pytest

FILEPATH = os.path.realpath(__file__)
TESTRUNNER_DIR = os.path.dirname(os.path.dirname(FILEPATH))

PYTEST_RC = {
    0: "all tests passed successfully",
    1: "some of the tests failed",
    2: "execution was interrupted by the user",
    3: "internal error happened while executing tests",
    4: "pytest command line usage error",
    5: "no tests were collected"
}

class TestDriver:
    def __init__(self, conf, platform):
        self.conf = conf
        self.platform = platform

    def run(self, module=None, test_suite=None,
            test=None, verbose=False, collect=False,
            skip_setup=None, mark=None, junit=None, traceback="short"):
        opts = []

        vars_opt = "--vars={}".format(self.conf.yaml_path)
        opts.append(vars_opt)

        platform_opt = "--platform={}".format(self.platform)
        opts.append(platform_opt)

        if verbose:
            opts.append("-s")

        # Dont capture logs
        opts.append("--show-capture=no")

        # generete detailed test results
        opts.append("-v")

        if collect:
            opts.append("--collect-only")

        if skip_setup is not None:
            opts.append(f"--skip-setup={skip_setup}")

        if junit is not None:
            opts.append(f"--junitxml={TESTRUNNER_DIR}/{junit}.xml")

        if mark is not None:
            opts.append(f'-m {mark}')

        opts.append(f'--tb={traceback}')

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

        if result in [0, 1]:
            raise SystemExit(result)

        if result in [2, 3, 4, 5]:
            raise Exception(f'error executing test {PYTEST_RC[result]}')

        raise Exception(f'unexpected return code from pytest {result}')
