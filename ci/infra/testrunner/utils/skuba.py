import json
import os

from timeout_decorator import timeout

from utils.format import Format
from utils.utils import (step, Utils)


class Skuba:

    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)

    # TODO: this function is currently not used. Identify points where it should
    # be invoked
    def _verify_tf_dependency(self):
        if not os.path.exists(self.conf.terraform_json_path):
            raise Exception(Format.alert("tf file not found. Please run terraform and try again{}"))

    def _verify_skuba_bin_dependency(self):
        skuba_bin_path = os.path.join(self.conf.workspace, 'go/bin/skuba')
        if not os.path.isfile(skuba_bin_path):
            raise FileNotFoundError(Format.alert("skuba not found at {}. Please run create-skuba and try again".format(skuba_bin_path)))

    def _verify_bootstrap_dependency(self):
        if not os.path.exists(os.path.join(self.conf.workspace, "test-cluster")):
            raise Exception(Format.alert("test-cluster not found. Please run bootstrap and try again"))

    @timeout(600)
    @step
    def create_skuba(self):
        """Configure Environment"""
        self.utils.runshellcommand("rm -fr go")
        self.utils.runshellcommand("mkdir -p go/src/github.com/SUSE")
        self.utils.runshellcommand("cp -a skuba go/src/github.com/SUSE/")
        self.utils.gorun("go version")
        print("Building skuba")
        self.utils.gorun("make")

    @step
    def cleanup(self):
        """Cleanup skuba working environment"""
        # TODO: check why (and if) the following two commands are needed
        cmd = 'mkdir -p {}/logs'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        # This is pretty aggressive but modules are also present
        # in workspace and they lack the 'w' bit so just set
        # everything so we can do whatever we want during cleanup
        cmd = 'chmod -R 777 {}'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        dirs = [os.path.join(self.conf.workspace, "test-cluster"),
                os.path.join(self.conf.workspace, "go"),
                os.path.join(self.conf.workspace, "logs"),
                os.path.join(self.conf.workspace, "ssh-agent-sock"),
                os.path.join(self.conf.workspace, "test-cluster")]

        cleanup_failure = False
        for dir in dirs:
            try: 
                self.utils.runshellcommand("rm -rf {}".format(dir))
            except Exception as ex:
                cleanup_failure = True
                print("Received the following error {}".format(ex))
                print("Attempting to finish cleaup")

        if cleanup_failure:
            raise Exception("Failure(s) during cleanup")

    def num_of_nodes(self):

        test_cluster = os.path.join(self.conf.workspace, "test-cluster")
        binpath = os.path.join(self.conf.workspace, 'go/bin/skuba')
        cmd = "cd " + test_cluster + "; " +binpath + " cluster status"
        output = self.utils.runshellcommand_withoutput(cmd)
        return output.count("master"), output.count("worker")

    def _load_tfstate(self):
        fn = os.path.join(self.conf.terraform_dir, "terraform.tfstate")
        print("Reading {}".format(fn))
        return output.count("master"), output.count("worker")

    def _load_tfstate(self):
        fn = os.path.join(self.conf.terraform_dir, "terraform.tfstate")
        print("Reading {}".format(fn))
        with open(fn) as f:
            self.state= json.load(f)

    def _get_lb_ipaddr(self):
        return self.state["modules"][0]["outputs"]["ip_ext_load_balancer"]["value"]

    def _get_masters_ipaddrs(self):
        return self.state["modules"][0]["outputs"]["ip_masters"]["value"]

    def _get_workers_ipaddrs(self):
        return self.state["modules"][0]["outputs"]["ip_workers"]["value"]

    @timeout(600)
    @step
    def gather_logs(self):
        self._load_tfstate()

        logging_error = False

        try:
            ipaddrs = self._get_masters_ipaddrs() + self._get_workers_ipaddrs()
            for ipa in ipaddrs:
                print("--------------------------------------------------------------")
                print("Gathering logs from {}".format(ipa))
                self.utils.ssh_run(ipa, "cat /var/run/cloud-init/status.json")
                print("--------------------------------------------------------------")
                self.utils.ssh_run(ipa, "cat /var/log/cloud-init-output.log")
        except Exception as ex:
            logging_error = True
            print("Error while collecting logs from cluster \n {}".format(ex))

        if logging_error:
            raise Exception("Failure(s) while collecting logs")


