#!/usr/bin/env python3 -Wd -b

"""
    Runs end-to-end product tests for v4+.
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

from argparse import Namespace
from functools import wraps
import json
import os
import subprocess
import sys

import requests

from timeout_decorator import timeout

__version__ = "0.0.3"

help = """
This script is meant to be run manually on test servers, developer desktops
and by Jenkins.

Warning: it removes docker containers, VMs, images, and network configuration.

It creates a virtualenv.

Requires root privileges.

"""

# Please flag requirements for packages with: #requirepkg <packagename>
# Env vars with #requireenv
# ...and other stuff with:  #require

STAGE_NAMES = (
    "info", "github_collaborator_check", "git_rebase",
    "initial_cleanup", "create_environment",
    "configure_environment",
    "bootstrap_environment", "grow_environment",
    "gather_logs", "final_cleanup"
)

TFSTATE_USER_HOST="ci-tfstate@hpa6s10.caasp.suse.net"

# Jenkins env vars: BUILD_NUMBER JOB_NAME GITHUB_TOKEN OPENRC CHANGE_AUTHOR

# global conf
conf = None

ssh_opts = "-oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null " + \
    "-oConnectTimeout=10 -oBatchMode=yes "

def ws_join(path):
    if path.startswith('/'):
        path = path[1:]
    return os.path.join(conf.workspace, path)

# TODO: reimplement dry run

def sh(cmd, env=None):
    """emulate Jenkins `sh`"""
    path = conf.workspace
    print(">  in {}".format(path))
    print("$ {}".format(cmd))
    if conf.dryrun:
        return

    p = subprocess.call(cmd, cwd=path, stderr=sys.stdout.buffer, shell=True,
                        env=env)
    if p != 0:
        raise Exception("'{}' exited with {}".format(cmd, p))

def sh_fork(cmd):
    """emulate Jenkins `sh`"""
    print("$ {}".format(cmd))
    if conf.dryrun:
        return
    return subprocess.Popen(cmd, shell=True)

def shp(path, cmd, env=None):
    """emulate Jenkins `sh`"""
    if not os.path.isabs(path):
        path = os.path.join(conf.workspace, path)

    print(">  in {}".format(path))
    print("$ {}".format(cmd))
    if conf.dryrun:
        return

    subprocess.check_call(cmd, cwd=path, shell=True, env=env)

def create_workspace_dir():
    try:
        os.makedirs(conf.workspace)
    except:
        print(conf.workspace, "created")
        pass

## nested output blocks
DOT = '\033[34m●\033[0m'
DOT_exit = '\033[32m●\033[0m'

_stepdepth = 0
def step(foo=None):
    def deco(f):
        @wraps(f)
        def wrapped(*args, **kwargs):
            global _stepdepth
            _stepdepth += 1
            print("{}  {} {}".format(DOT * _stepdepth, f.__name__,
                                     f.__doc__ or ""))
            r = f(*args, **kwargs)
            print("{}  exiting {}".format(DOT_exit * _stepdepth, f.__name__))
            _stepdepth -= 1
            return r
        return wrapped
    return deco


def chmod_id_shared():
    key_fn = conf.id_shared_fn
    sh("chmod 400 " + key_fn)

def locate_tfstate(platform):
    assert platform in ("openstack", "vmware")
    return os.path.join(conf.workspace,
        "caaspctl/ci/infra/{}/terraform.tfstate".format(platform))

@step()
def fetch_tfstate(platform, run_name):
    chmod_id_shared()
    fn = locate_tfstate(platform)
    key_fn = conf.id_shared_fn
    sh("scp {} -i {} {}:~/tfstates/{} {}".format(
        ssh_opts, key_fn, TFSTATE_USER_HOST, run_name, fn))

@step()
def push_tfstate(platform, run_name):
    chmod_id_shared()
    key_fn = conf.id_shared_fn
    fn = locate_tfstate(platform)
    sh("ssh {} -i {} {} -- 'mkdir -p ~/tfstates/{}'".format(
        ssh_opts, key_fn, TFSTATE_USER_HOST, os.path.dirname(run_name)))
    sh("scp {} -i {} {} {}:~/tfstates/{}".format(
        ssh_opts, key_fn, fn, TFSTATE_USER_HOST, run_name))


@timeout(7)
@step()
def info():
    """Node info"""
    print("Env vars: {}".format(sorted(os.environ)))

    sh('ip a')
    sh('ip r')
    sh('cat /etc/resolv.conf')

    try:
        r = requests.get('http://169.254.169.254/2009-04-04/meta-data/public-ipv4', timeout=2)
        r.raise_for_status()
    except (requests.HTTPError, requests.Timeout) as err:
        print(err)
        print('Meta Data service unavailable could not get external IP addr')
    else:
        print('External IP addr: {}'.format(r.text))


@timeout(125)
@step()
def initial_cleanup():
    """Cleanup"""
    sh('mkdir -p {}/logs'.format(conf.workspace))
    sh('chmod a+x {}'.format(conf.workspace))
    # TODO: implement cleanups for vsphere etc
    if conf.stack_type == 'openstack-terraform':
        cleanup_openstack_terraform()


@timeout(90)
@step()
def github_collaborator_check():
    if conf.no_checkout or conf.no_collab_check:
        print("Skipping collaborator check")
        return
    print("Starting GitHub Collaborator Check")
    org = "SUSE"
    repo = 'avantgarde-caaspctl'
    user = conf.change_author
    token = os.getenv('GITHUB_TOKEN')
    url = "https://api.github.com/repos/{}/{}/collaborators/{}"
    url = url.format(org, repo, user)
    if user is "":
        return

    # Check if a change is from collaborator, or not.
    # Require approval for non-collaborators. As non-collaborators are
    # already considered untrusted by Jenkins, Jenkins will load the
    # Pipeline and library from the target branch and NOT from the
    # outside collaborators fork / pull request.
    headers = {
        "Accept": "application/vnd.github.hellcat-preview+json",
        "Authorization": "token {}".format(token),
    }
    r = requests.get(url, headers=headers)
    # 204 yes, 404 no   :-/
    if r.status_code == 204:
        print("Test execution for collaborator {} allowed".format(user))
        return

    msg = "Test execution for unknown user {} NOT allowed".format(user)
    print(msg)
    raise Exception(msg)


@timeout(90)
@step()
def git_rebase():
    if conf.branch_name.lower() == "master":
        print("Rebase not required for master.")
        return

    try:
        shp("caaspctl", 'git -c "user.name=${CHANGE_AUTHOR}" -c "user.email=${CHANGE_AUTHOR_EMAIL}" rebase origin/master')
    except subprocess.CalledProcessError as ex:
        print(ex)
        print("Rebase failed, manual rebase is required.")
        shp("caaspctl", "git rebase --abort")
        sys.exit(1)
    except Exception as ex:
        print(ex)
        print("Unknown error exiting.")
        sys.exit(2)


@step()
def fetch_openstack_terraform_output():
    shp("caaspctl/ci/infra/openstack", "source {}; "
        "terraform output -json > {}/tfout.json".format(
            conf.openrc, conf.workspace))

def ssh(ipaddr, cmd):
    key_fn = conf.id_shared_fn
    cmd = "ssh " + ssh_opts + " -i {key_fn} {username}@{ip} -- '{cmd}'".format(
        key_fn=key_fn, ip=ipaddr, cmd=cmd, username=conf.username)
    sh(cmd)

def authorized_keys():
    fn = conf.id_shared_fn + ".pub"
    with open(fn) as f:
        shared_pubkey = f.read().strip()
    return shared_pubkey

@step()
def boot_openstack():
    # Implement a simple state machine to handle tfstate files
    # and prevent leaving around "forgotten" stacks
    print("Test SSH")

    # generate terraform variables file
    fn = os.path.join(conf.workspace, "terraform.tfvars")
    with open(fn, 'w') as f:
        f.write(generate_tfvars_file())
    print("Wrote %s" % fn)

    print("Init terraform")
    shp("caaspctl/ci/infra/openstack", "terraform init")
    shp("caaspctl/ci/infra/openstack", "terraform version")
    print("------------------------")
    print()
    print("To clean up OpenStack manually, run:")
    print(("./testrunner stack-type=openstack-terraform job_name={} "
          "build_number={} openrc={} stage=initial_cleanup").format(
          conf.job_name, conf.build_number, conf.openrc))
    print()
    print("------------------------")
    plan_cmd = ("source {};"
        " terraform plan -var-file='{}/terraform.tfvars'"
        " -out {}/tfout".format(conf.openrc, conf.workspace, conf.workspace)
    )
    apply_cmd = ("source {}; terraform apply -auto-approve {}/tfout".format(
        conf.openrc, conf.workspace))
    for retry in range(1, 5):
        print("Run terraform plan - execution n. %d" % retry)
        shp("caaspctl/ci/infra/openstack", plan_cmd)
        print("Running terraform apply - execution n. %d" % retry)
        try:
            shp("caaspctl/ci/infra/openstack", apply_cmd)
            push_tfstate("openstack", conf.run_name)
            break

        except:
            print("Failed terraform apply n. %d" % retry)
            # push the tfstate anyways in case something is created
            push_tfstate("openstack", conf.run_name)
            if retry == 4:
                print("Last failed attempt, cleaning up and exiting")
                cleanup_openstack_terraform(clean_previous=False)
                raise Exception("Failed OpenStack deploy")

    fetch_openstack_terraform_output()

def print_ipaddr_summary():
    print("-" * 20)
    print()
    print("LB IP addr: " + get_lb_ipaddr())
    print("Masters IP addrs: " + " ".join(sorted(get_masters_ipaddrs())))
    print("Workers IP addrs: " + " ".join(sorted(get_workers_ipaddrs())))
    print()
    print("-" * 20)

@step()
def create_environment():
    """Create Environment"""
    if conf.stack_type == 'caasp-kvm':
        raise NotImplementedError # TODO

    elif conf.stack_type == 'openstack-terraform':
        boot_openstack()

    elif conf.stack_type == 'bare-metal':
        shp("${WORKSPACE}/caaspctl/ci/infra/bare-metal/deployer",
            "./deployer ${JOB_NAME}-${BUILD_NUMBER} --deploy-nodes "
            "--master-count 1 --worker-count 1"
            " --conffile deployer.conf.json")
        sh("cp ${WORKSPACE}/caaspctl/ci/infra/bare-metal/deployer/environment.json ${WORKSPACE}/")
        bare_metal_cloud_init()

    elif conf.stack_type == 'vmware-terraform':
        #requireenv VSPHERE_USER
        #requireenv VSPHERE_PASSWORD
        shp("vmware", "terraform init")
        shp("vmware",
            "terraform apply -auto-approve"
            " -var internal_net=containers-ci"
            " -var stack_name=${JOB_NAME}-${BUILD_NUMBER}")
        shp("vmware", "terraform output -json > tfout.json")

    print_ipaddr_summary()

def gorun(rundir, cmd, extra_env=None):
    env = {
        'GOPATH': ws_join('go'),
        'PATH': os.environ['PATH']
    }
    if extra_env:
        env.update(extra_env)
    shp(rundir, cmd, env=env)

def caaspctl(rundir, cmd):
    env={"SSH_AUTH_SOCK": pick_ssh_agent_sock()}
    binpath = ws_join('go/bin/caaspctl')
    gorun(rundir, binpath + ' ' + cmd, extra_env=env)


@timeout(120)
@step()
def configure_environment():
    """Configure Environment"""
    sh("mkdir -p go/src/github.com/SUSE")
    # TODO: better idempotency?
    try:
        sh("test -d go/src/github.com/SUSE/caaspctl || "
           "cp -a caaspctl go/src/github.com/SUSE/")
    except:
        pass
    gorun(conf.workspace, "go version")
    print("Building caaspctl")
    gorun("go/src/github.com/SUSE/caaspctl", "make")


def load_tfstate():
    fn = ws_join("caaspctl/ci/infra/openstack/terraform.tfstate")
    print("Reading {}".format(fn))
    with open(fn) as f:
        return json.load(f)

def get_lb_ipaddr():
    j = load_tfstate()
    return j["modules"][0]["outputs"]["ip_ext_load_balancer"]["value"]

def get_masters_ipaddrs():
    j = load_tfstate()
    return j["modules"][0]["outputs"]["ip_masters"]["value"]

def get_workers_ipaddrs():
    j = load_tfstate()
    return j["modules"][0]["outputs"]["ip_workers"]["value"]

@step()
def caaspctl_cluster_init():
    print("Cleaning up any previous test-cluster dir")
    sh("rm /go/src/github.com/SUSE/caaspctl/test-cluster -rf")
    caaspctl("go/src/github.com/SUSE/caaspctl",
        "cluster init --control-plane %s test-cluster" %
        get_lb_ipaddr())

def locate_id_shared():
    return ws_join("caaspctl/ci/infra/id_shared")

@step()
def kubeadm_reset():
    # TODO: temporary hack - will be done by caaspctl
    ipaddr = get_masters_ipaddrs()[0]
    ssh(ipaddr, 'sudo kubeadm reset -f')

@step()
def caaspctl_node_bootstrap():
    caaspctl("go/src/github.com/SUSE/caaspctl/test-cluster",
        "node bootstrap --user {username} --sudo --target "
        "{ip} my-master-0".format(ip=get_masters_ipaddrs()[0], username=conf.username))

@step()
def caaspctl_cluster_status():
    caaspctl("go/src/github.com/SUSE/caaspctl/test-cluster",
        "cluster status")

@step()
def caaspctl_node_join(role="worker", nr=0):
    if role == "master":
        ip_addr = get_masters_ipaddrs()[nr]
    else:
        ip_addr = get_workers_ipaddrs()[nr]

    caaspctl("go/src/github.com/SUSE/caaspctl/test-cluster",
        "node join --role {role} --user {username} --sudo --target "
          "{ip} my-{role}-{nr}".format(role=role, ip=ip_addr, nr=nr, username=conf.username))

def pick_ssh_agent_sock():
    return ws_join("ssh-agent-sock")

@timeout(10)
@step()
def setup_ssh():
    chmod_id_shared()
    print("Starting ssh-agent ")
    # use a dedicated agent to minimize stateful components
    sock_fn = pick_ssh_agent_sock()
    try:
        sh("pkill -f 'ssh-agent -a %s'" % sock_fn)
        print("Killed previous instance of ssh-agent")
    except:
        pass
    sh("ssh-agent -a %s" % sock_fn)
    print("adding id_shared ssh key")
    sh("ssh-add " + conf.id_shared_fn, env={"SSH_AUTH_SOCK": sock_fn})
    # TODO kill agent on cleanup


@timeout(400)
@step()
def bare_metal_cloud_init():
    #requirepkg cloud-init
    # TODO: move cloud-init outside of openstack dir
    shp("caaspctl/ci/infra/openstack", "cloud-init status")
    #shp("caaspctl/ci/infra/openstack", "cloud-init collect-logs ")

@timeout(600)
@step()
def bootstrap_environment():
    """Bootstrap Environment"""
    setup_ssh()
    caaspctl_cluster_init()
    kubeadm_reset()
    # bootstrap on the first master and then join the first worker. The other workers and masters are joined in `grow_environment`
    caaspctl_node_bootstrap()
    caaspctl_node_join(role="worker", nr=0)
    try:
        caaspctl_cluster_status()
    except:
        pass

@timeout(600)
@step()
def grow_environment():
    # master-0 and worker-0 are already in the cluster
    """Grow Environment by one worker and 2 masters"""
    caaspctl_node_join(role="worker", nr=1)
    caaspctl_node_join(role="master", nr=1)
    caaspctl_node_join(role="master", nr=2)
    try:
        caaspctl_cluster_status()
    except:
        pass

@step()
def create_environment_workers_bare_metal():
    # Warning: requires deployer.conf.json
    shp('caaspctl/ci/infra/bare-metal/deployer',
        './deployer ${JOB_NAME}-${BUILD_NUMBER} --deploy-nodes --logsdir ${WORKSPACE}/logs'
        " --conffile deployer.conf.json")
    shp('caaspctl/ci/infra/bare-metal/deployer',
        "cp environment.json ${WORKSPACE}/environment.json")
    # FIXME generate a new form of environment.json
    shp('caaspctl/ci/infra/bare-metal/deployer',
        '${WORKSPACE}/automation/misc-tools/generate-ssh-config ${WORKSPACE}/environment.json')
    archive_artifacts('${WORKSPACE}', 'environment.json')


def load_env_json():
    with open(ws_join("environment.json")) as f:
        return json.load(f)


@timeout(30)
@step()
def setup_testinfra():
    #FIXME implement tests
    return

@timeout(30 * 10) # implement parallel run
@step()
def run_testinfra():
    #FIXME implement tests
    return

@timeout(125)
@step()
def add_node():
    raise NotImplementedError

@step()
def run_conformance_tests():
    """Run K8S Conformance Tests"""
    # TODO
    pass

@timeout(300)
@step()
def _gather_logs(minion):
    return


@timeout(60)
@step()
def gather_logs():
    """Gather Kubic Logs"""
    if conf.dryrun:
        print("DRYRUN: skipping gather_logs")
        return

    ipaddrs = get_masters_ipaddrs() + get_workers_ipaddrs()
    for ipa in ipaddrs:
        print("--------------------------------------------------------------")
        print("Gathering logs from {}".format(ipa))
        ssh(ipa, "cat /var/run/cloud-init/status.json")
        print("--------------------------------------------------------------")
        ssh(ipa, "cat /var/log/cloud-init-output.log")


def archive_artifacts(path, glob):
    sh("mkdir -p artifacts")
    path = os.path.join(path, glob)
    try:
        sh("rsync -a " + path + " {}/artifacts".format(conf.workspace))
    except:
        print("rsync error")

@step()
def archive_logs():
    """Archive Logs"""
    archive_artifacts(conf.workspace, 'logs/**')

@timeout(15)
@step()
def cleanup_kvm():
    #TODO
    raise NotImplementedError

def show_vmware_status():
    # TODO cleanup
    pass

@timeout(15)
@step()
def cleanup_vmware():
    show_vmware_status()
    # TODO cleanup
    show_vmware_status()


@timeout(60)
def _cleanup_openstack_terraform(run_name):
    cmd = ("source {orc};"
        " terraform destroy -auto-approve"
        " -var internal_net=net-{run}"
        " -var stack_name={run}").format(orc=conf.openrc, run=run_name)
    shp("caaspctl/ci/infra/openstack", cmd)

@step()
def cleanup_openstack_terraform(clean_previous=True):
    """Cleanup OpenStack"""
    # OpenStack can leave running stacks around. This functions
    # cleans stacks from previous runs and the current one.
    shifter = [-2, -1, 0] if clean_previous else [0]
    for shift in shifter:
        bn = conf.build_number + shift
        if bn < 1:
            continue
        run_name = "{}-{}".format(conf.job_name, bn)
        try:
            fetch_tfstate("openstack", run_name)
            print("Cleaning up run n. {}".format(bn))
            _cleanup_openstack_terraform(run_name)
            # Upload "empty" tfstate
            push_tfstate("openstack", run_name)
        except:
            print("Nothing to clean up for run n. {}".format(bn))


@timeout(40)
@step()
def cleanup_bare_metal():
    shp('caaspctl/ci/infra/bare-metal/deployer',
        "./deployer --release {run} --conffile deployer.conf.json".format(
        run=conf.run_name))

@timeout(40)
@step()
def cleanup_hyperv():
    # TODO
    raise NotImplementedError


@step()
def final_cleanup():
    """Cleanup"""
    if conf.no_destroy:
        print("no-destroy was passed: skipping cleanup")
        return
    if conf.stack_type == 'caasp-kvm':
        cleanup_kvm()
    if conf.stack_type == 'vmware-terraform':
        cleanup_vmware()
    elif conf.stack_type == 'openstack-terraform':
        cleanup_openstack_terraform()
    elif conf.stack_type == 'bare-metal':
        cleanup_bare_metal()


def parse_args():
    """Handle free-form CLI parameters
    """
    conf = Namespace()
    conf.dryrun = False
    conf.stack_type = 'openstack-terraform'
    conf.stage = None   # None: run all stages
    conf.change_author = os.getenv("CHANGE_AUTHOR")
    conf.no_checkout = False
    conf.no_collab_check = False
    conf.no_destroy = False
    conf.workers = "3"
    conf.job_name = os.getenv("JOB_NAME")
    conf.build_number = os.getenv("BUILD_NUMBER")
    conf.workspace = os.getenv("WORKSPACE")
    conf.id_shared_fn = None
    conf.openrc = os.getenv("OPENRC")
    conf.master_count = "3"
    conf.worker_count = "3"
    conf.admin_cpu = "4"
    conf.admin_ram = "8192"
    conf.master_cpu = "4"
    conf.master_ram = "4096"
    conf.worker_cpu = "4"
    conf.worker_ram = "4096"
    conf.netlocation = "provo"
    conf.channel = "devel"
    conf.replica_count = "5"
    conf.replicas_creation_interval_seconds = "5"
    conf.podname = "default"
    conf.generate_pipeline = False
    conf.username = "sles"
    conf.branch_name = "master"

    if '-h' in sys.argv or '--help' in sys.argv:
        print("Help:\n\n")
        print(help)
        print("\nSupported options:\n")
        for k, v in sorted(conf.__dict__.items()):
            k = k.replace('_', '-')
            if v == False:
                print("    {}".format(k))
            else:
                print("    {}={}".format(k, v))
        print()
        sys.exit()

    for a in sys.argv[1:]:
        if '=' in a:
            # extract key-value args
            k, v = a.split('=', 1)[0:2]
        else:
            k, v = a, True

        k = k.replace('-', '_')
        if k in conf:
            conf.__setattr__(k, v)
        else:
            print("Unexpected conf param {}".format(k))
            sys.exit(1)

    try:
        conf.build_number = int(conf.build_number)
    except ValueError:
        print("ERROR: unable to parse build number as an integer")
        sys.exit(1)

    conf.run_name = "{}-{}".format(conf.job_name, conf.build_number)
    assert conf.workspace, "A workspace env var or CLI param is required"

    return conf

def check_root_user():
    if os.getenv('EUID') != "0":
        print("Error: this script needs to be run as root")
        sys.exit(1)

def generate_pipeline():
    """Generate stub Jenkins pipeline"""
    # TODO: show PARAMS as a parameter, default with type=openstack-terraform
    # TODO: collect artifacts
    tpl = """
pipeline {
    agent any
    environment {
        OPENRC = credentials("ecp-cloud-shared")
        GITHUB_TOKEN = credentials("github-token")
        GITLAB_TOKEN = credentials("gitlab-token")
        PARAMS = ""
    }
    stages {
        stage('Init') { steps {
            sh "rm ${WORKSPACE}/* -rf"
            sh "git clone https://${GITHUB_TOKEN}@github.com/SUSE/caaspctl"
        } }
        %s
   }
   post {
        unsuccessful {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=gather_logs ${PARAMS}"
            sh "caaspctl/ci/infra/testrunner/testrunner stage=final_cleanup ${PARAMS}"
        }
    }
}
    """
    stage_tpl = """
        stage('%s') { steps {
            sh "caaspctl/ci/infra/testrunner/testrunner stage=%s ${PARAMS}"
        } }
    """

    stages_block = ""
    for sn in STAGE_NAMES:
        stages_block += stage_tpl % (sn.replace('_', ' ').title(), sn)
    print(tpl % stages_block)


def generate_tfvars_file():
    """Generate terraform tfvars file"""
    tpl = '''
internal_net = "{job_name}"
stack_name = "{job_name}"

image_name = "SLES15-SP1-JeOS-RC1-with-fixed-kernel-default"

repositories = [
  {{
    caasp_40_devel_sle15sp1 = "http://download.suse.de/ibs/Devel:/CaaSP:/4.0/SLE_15_SP1/"
  }},
  {{
    sle15sp1_pool = "http://download.suse.de/ibs/SUSE:/SLE-15-SP1:/GA/standard/"
  }},
  {{
    sle15sp1_update = "http://download.suse.de/ibs/SUSE:/SLE-15-SP1:/Update/standard/"
  }},
  {{
    sle15_pool = "http://download.suse.de/ibs/SUSE:/SLE-15:/GA/standard/"
  }},
  {{
    sle15_update = "http://download.suse.de/ibs/SUSE:/SLE-15:/Update/standard/"
  }},
  {{
    suse_ca = "http://download.suse.de/ibs/SUSE:/CA/SLE_15_SP1/"
  }}
]

packages = [
  "ca-certificates-suse",
  "kubernetes-kubeadm",
  "kubernetes-kubelet",
  "kubernetes-client"
]

masters = 3
workers = 2

authorized_keys = [
  "{authorized_keys}"
]
username = "{username}"
    '''.format(job_name=conf.run_name, authorized_keys=authorized_keys(),
               username=conf.username)
    return tpl


def main():
    global conf
    print("Testrunner v. {}".format(__version__))
    conf = parse_args()
    if conf.id_shared_fn is None:
        conf.id_shared_fn = ws_join("caaspctl/ci/infra/id_shared")

    if conf.stack_type == 'openstack-terraform':
        assert conf.openrc, "An openrc env var or CLI param is required"

    if conf.generate_pipeline:
        generate_pipeline()
        sys.exit()

    print("Conf: {}".format(conf))
    print("PATH: {}".format(os.getenv("PATH")))

    if not conf.dryrun:
        create_workspace_dir()

    if conf.stage is None:
        # run all stages and exit
        for sn in STAGE_NAMES:
            globals()[sn]()
        return

    # run one stage
    if conf.stage not in STAGE_NAMES:
        print("Unknown stage name. Valid names are:\n")
        for sn in STAGE_NAMES:
            print("  %s" % sn)
        sys.exit(1)

    # call stage function by name
    assert conf.stage in globals()
    globals()[conf.stage]()


if __name__ == '__main__':
    main()
