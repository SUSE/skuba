import os, json
from timeout_decorator import timeout
from utils import step
from utils import Utils
from utils import Constant

class Skuba:

    def __init__(self, conf):
        self.conf = conf
        self.utils = Utils(self.conf)
        self.cwd = "{}/test-cluster".format(self.conf.workspace)

    def _verify_skuba_bin_dependency(self):
        skuba_bin_path = os.path.join(self.conf.workspace, 'go/bin/skuba')
        if not os.path.isfile(skuba_bin_path):
            raise FileNotFoundError("{}skuba not found at {}. Please run create-skuba and try again".format(
                Constant.RED, skuba_bin_path, Constant.RED_EXIT))

    def _verify_bootstrap_dependency(self):
        if not os.path.exists(os.path.join(self.conf.workspace, "test-cluster")):
            raise Exception("{}test-cluster not found. Please run bootstrap and try again{}".format(
                Constant.RED, Constant.RED_EXIT))

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

    @step
    def cluster_init(self):
        self._load_tfstate()

        print("Cleaning up any previous test-cluster dir")
        self.utils.runshellcommand("rm -rf {}".format(self.cwd))
        cmd = "cluster init --control-plane {} test-cluster".format(self._get_lb_ipaddr())
        # Override work directory, because init must run in the parent directory of the
        # cluster directory
        self._run_skuba(cmd, cwd=self.conf.workspace)

    @step
    def node_bootstrap(self):
        self._load_tfstate()
        self._verify_bootstrap_dependency()

        cmd = "node bootstrap --user {username} --sudo --target \
                 {ip} my-master-0".format(ip=self._get_masters_ipaddrs()[0], username=self.conf.nodeuser)
        self._run_skuba(cmd)

    @step
    def node_join(self, role="worker", nr=0):
        self._load_tfstate()
        self._verify_bootstrap_dependency()
        
        try:
            if role == "master":
                ip_addr = self._get_masters_ipaddrs()[nr]
            else:
                ip_addr = self._get_workers_ipaddrs()[nr]
        except:
            raise("{}Error: there is not enough node to add {} node in cluster{}".format(
                Constant.RED, role, Constant.RED_EXIT))

        cmd = "node join --role {role} --user {username} --sudo --target {ip} my-{role}-{nr}".format(
            role=role, ip=ip_addr, nr=nr, username=self.conf.nodeuser)
        try:
            self._run_skuba(cmd)
        except:
            raise ("{}Error: {}{}".format(Constant.RED, cmd, Constant.RED_EXIT))

    @step
    def node_remove(self, role="worker", nr=0):
        self._load_tfstate()
        self._verify_bootstrap_dependency()

        if nr <= 0:
            raise ("{}Error: there is not enough node to remove {} node in cluster{}".format(
                Constant.RED, role,Constant.RED_EXIT))

        if role == "master":
            ip_addr = self._get_masters_ipaddrs()[nr]
        else:
            ip_addr = self._get_workers_ipaddrs()[nr]

        cmd = "node remove my-{role}-{nr}".format(role=role, ip=ip_addr, nr=nr, username=self.conf.nodeuser)
        try:
            self._run_skuba(cmd)
        except:
            raise ("{}Error: {}{}".format(Constant.RED, cmd, Constant.RED_EXIT))

    def cluster_status(self):
        self._verify_bootstrap_dependency()
        self._run_skuba("cluster status")

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

    def _run_skuba(self, cmd, cwd=None):
        """Running skuba command in cwd.
        The cwd defautls to {workspace}/test-cluster but can be overrided
        for example, for the init command that must run in {workspace}
        """
        self._verify_skuba_bin_dependency()

        if cwd is None:
           cwd=self.cwd

        env = {"SSH_AUTH_SOCK": os.path.join(self.conf.workspace, "ssh-agent-sock")}

        binpath = os.path.join(self.conf.workspace, 'go/bin/skuba')
        self.utils.runshellcommand(binpath + " "+ cmd, cwd=cwd, env=env)

