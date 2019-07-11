import os
import pytest 

class PyTestOpts:

    SHOW_OUTPUT = "-s"

    COLLECT_TESTS = "--collect-only"

class TestDriver:


    def __init__(self, conf, platform):
        self.conf = conf
        self.platform = platform
        
    def run(self, module="tests", test_suite=None, test=None, verbose=False, collect=False):
        
        opts = []
        
        vars_opt = "--vars={}".format(self.conf.yaml_path)
        opts.append(vars_opt)

        platform_opt = "--platform={}".format(self.platform)
        opts.append(platform_opt)

        if verbose:
            opts.append(PyTestOpts.SHOW_OUTPUT)

        if collect:
            opts.append(PyTestOpts.COLLECT_TESTS)

        test_path = module

        if test_suite:
            if not test_suite.endswith(".py"):
                raise ValueError("Test suite must be a python file")
            test_path = os.path.join(module,test_suite)
        
        if test:
            test_path += "{}::{}".format(test_path, test)

        # Path must be the last argument
        opts.append(test_path)

        pytest.main(args=opts) 
