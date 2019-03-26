#!/usr/bin/env python3 -Wd -b

"""
    Runs end-to-end product tests for v4+.
    This script can be run from Jenkins or manually, on developer desktops or servers.
"""

from argparse import Namespace
from functools import wraps
import json
import os
import re
import subprocess
import sys

import requests

from timeout_decorator import timeout

__version__ = "0.0.3"

help = """
This script is meant to be run manually on test servers, developer desktops
and by Jenkins.

Warning: it removes docker containers, VMs, images, and network configuration.

It creates a workspace directory and a virtualenv.

Requires root privileges.

"""

# Please flag requirements for packages with: #requirepkg <packagename>
# Env vars with #requireenv
# ...and other stuff with:  #require

STAGE_NAMES = (
    "info", "github_collaborator_check",
    "initial_cleanup", "create_environment",
    "configure_environment",
    "bootstrap_environment", "grow_environment",
    "gather_logs", "final_cleanup"
)

TFSTATE_USER_HOST="ci-tfstate@hpa6s10.caasp.suse.net"

# Jenkins env vars: BUILD_NUMBER

env_defaults = dict(
    WORKSPACE=os.path.join(os.getcwd(), "workspace"),
    BMCONF="error-bare-metal-config-file-not-set",
)

# global conf
conf = None

ssh_opts = "-oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null " + \
    "-oConnectTimeout=10 -oBatchMode=yes "

def getvar(name):
    """Resolve in order:
    - CLI k/v variable (case insensitive)
    - environment variable (case sensitive)
    - default value
    """
    lc = name.lower()
    if hasattr(conf, lc):
        return getattr(conf, lc)
    if name in os.environ:
        return os.environ[name]
    if name in env_defaults:
        return env_defaults[name]
    raise Exception("env variable '{}' not found".format(name))


def replace_vars(s):
    """Replace jenkins ${} variables"""
    try:
        for match in re.findall('\$\{[\w\-\.]+\}', s):
            varname = match[2:-1]
            val = getvar(varname)
            s = s.replace(match, val, 1)  # replace only the first
        return s
    except Exception as e:
        print("Error while replacing '{}'".format(s))
        print(e)
        raise

run_name = replace_vars("${JOB_NAME}-${BUILD_NUMBER}")

# TODO: Replacing Jenkins variables like ${WORKSPACE} is a temporary hack
# to ease the migration from groovy.

# TODO: reimplement dry run

def sh(cmd, env=None):
    """emulate Jenkins `sh`"""
    cmd = replace_vars(cmd)
    path = replace_vars("${WORKSPACE}")
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
    cmd = replace_vars(cmd)
    print("$ {}".format(cmd))
    if conf.dryrun:
        return
    return subprocess.Popen(cmd, shell=True)

def shp(path, cmd, env=None):
    """emulate Jenkins `sh`"""
    cmd = replace_vars(cmd)
    path = replace_vars(path)
    if not os.path.isabs(path):
        path = os.path.join(replace_vars("${WORKSPACE}"), path)

    print(">  in {}".format(path))
    print("$ {}".format(cmd))
    if conf.dryrun:
        return

    subprocess.check_call(cmd, cwd=path, shell=True, env=env)

def create_workspace_dir():
    path = replace_vars("${WORKSPACE}")
    try:
        os.makedirs(path)
    except:
        print(path, "created")
        pass

## nested output blocks
if 'JENKINS_HOME' in os.environ:
    DOT = '●'
    DOT_exit = '●'
else:
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
    key_fn = locate_id_shared()
    sh("chmod 400 " + key_fn)

def locate_tfstate(platform):
    assert platform in ("openstack", "vmware")
    return os.path.join(replace_vars("${WORKSPACE}"),
        "caaspctl/ci/infra/{}/terraform.tfstate".format(platform))

@step()
def fetch_tfstate(platform):
    chmod_id_shared()
    fn = locate_tfstate(platform)
    key_fn = locate_id_shared()
    sh("scp {} -i {} {}:~/tfstates/{} {}".format(
        ssh_opts, key_fn, TFSTATE_USER_HOST, run_name, fn))

@step()
def push_tfstate(platform):
    chmod_id_shared()
    key_fn = locate_id_shared()
    fn = locate_tfstate(platform)
    sh("scp {} -i {} {} {}:~/tfstates/{}".format(
        ssh_opts, key_fn, fn, TFSTATE_USER_HOST, run_name))

@timeout(5)
@step()
def info():
    """Node info"""
    print("Env vars: {}".format(sorted(os.environ)))

    sh('ip a')
    sh('ip r')
    sh('cat /etc/resolv.conf')


@timeout(125)
@step()
def initial_cleanup():
    """Cleanup"""
    #sh('rm -rf ${WORKSPACE} || : ')
    #create_workspace_dir()
    sh('mkdir -p ${WORKSPACE}/logs')
    sh('chmod a+x ${WORKSPACE}')
    # TODO: implement cleanups for vsphere etc
    if conf.stack_type == 'openstack-terraform':
        try:
            fetch_tfstate("openstack")
            cleanup_openstack_terraform()
        except:
            print("Nothing to clean up")


@timeout(90)
@step()
def github_collaborator_check():
    if conf.no_checkout or conf.no_collab_check:
        print("Skipping collaborator check")
        return
    print("Starting GitHub Collaborator Check")
    org = "SUSE"
    repo = 'avantgarde-caaspctl'
    user = getvar('CHANGE_AUTHOR')
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

@step()
def fetch_openstack_terraform_output():
    shp("caaspctl/ci/infra/openstack", "source ${OPENRC}; "
        "terraform output -json > ${WORKSPACE}/tfout.json")

def ssh(ipaddr, cmd):
    key_fn = locate_id_shared()
    cmd = "ssh " + ssh_opts + " -i {key_fn} {username}@{ip} -- '{cmd}'".format(
        key_fn=key_fn, ip=ipaddr, cmd=cmd, username=conf.username)
    sh(cmd)

def authorized_keys():
    fn = locate_id_shared() + ".pub"
    with open(fn) as f:
        shared_pubkey = f.read().strip()
    return shared_pubkey

@step()
def boot_openstack():
    # Implement a simple state machine to handle tfstate files
    # and prevent leaving around "forgotten" stacks
    print("Test SSH")

    # generate terraform variables file
    fn = os.path.join(replace_vars("${WORKSPACE}"), "terraform.tfvars")
    with open(fn, 'w') as f:
        f.write(generate_tfvars_file())
    print("Wrote %s" % fn)

    print("Init terraform")
    shp("caaspctl/ci/infra/openstack", "terraform init")
    print("------------------------")
    print()
    print("To clean up OpenStack manually, run:")
    print(replace_vars("BUILD_NUMBER=${BUILD_NUMBER} "
        "JOB_NAME=${JOB_NAME} OPENRC=<replace-me> ./testrunner "
        "stack-type=openstack-terraform stage=initial_cleanup"))
    print()
    print("------------------------")
    for retry in range(1, 5):
        print("Run terraform plan - execution n. %d" % retry)
        shp("caaspctl/ci/infra/openstack", "source ${OPENRC};"
            " terraform plan -var-file='${WORKSPACE}/terraform.tfvars' -out ${WORKSPACE}/tfout"
        )
        print("Running terraform apply - execution n. %d" % retry)
        try:
            shp("caaspctl/ci/infra/openstack", "source ${OPENRC};"
                " terraform apply -auto-approve ${WORKSPACE}/tfout")
            push_tfstate("openstack")
            break

        except:
            print("Failed terraform apply n. %d" % retry)
            # push the tfstate anyways in case something is created
            push_tfstate("openstack")
            if retry == 4:
                print("Last failed attempt, cleaning up and exiting")
                cleanup_openstack_terraform()
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
        shp("vmware",
            "terraform output -json > ${WORKSPACE}/tfout.json")

    print_ipaddr_summary()

def gorun(rundir, cmd, extra_env=None):
    env = {
        'GOPATH': replace_vars("${WORKSPACE}") + '/go',
        'PATH': os.environ['PATH']
    }
    if extra_env:
        env.update(extra_env)
    shp(rundir, cmd, env=env)

def caaspctl(rundir, cmd):
    env={"SSH_AUTH_SOCK": pick_ssh_agent_sock()}
    binpath = os.path.join(replace_vars("${WORKSPACE}"), 'go/bin/caaspctl')
    gorun(rundir, binpath + ' ' + cmd, extra_env=env)


@timeout(90)
@step()
def configure_environment():
    """Configure Environment"""
    sh("mkdir -p ${WORKSPACE}/go/src/suse.com")
    # TODO: better idempotency?
    try:
        sh("test -d ${WORKSPACE}/go/src/suse.com/caaspctl || "
           "cp -a ${WORKSPACE}/caaspctl ${WORKSPACE}/go/src/suse.com/")
    except:
        pass
    print("Building caaspctl")
    gorun("${WORKSPACE}/go/src/suse.com/caaspctl", "make")


def load_tfstate():
    fn = replace_vars("${WORKSPACE}/caaspctl/ci/infra/openstack/terraform.tfstate")
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
    sh("rm /go/src/suse.com/caaspctl/test-cluster -rf")
    caaspctl("${WORKSPACE}/go/src/suse.com/caaspctl",
        "cluster init --control-plane %s test-cluster" %
        get_lb_ipaddr())

def locate_id_shared():
    return replace_vars("${WORKSPACE}/caaspctl/ci/infra/id_shared")

@step()
def kubeadm_reset():
    # TODO: temporary hack - will be done by caaspctl
    ipaddr = get_masters_ipaddrs()[0]
    ssh(ipaddr, 'sudo kubeadm reset -f')

@step()
def caaspctl_node_bootstrap():
    caaspctl("${WORKSPACE}/go/src/suse.com/caaspctl/test-cluster",
        "node bootstrap --user {username} --sudo --target "
        "{ip} my-master-0".format(ip=get_masters_ipaddrs()[0], username=conf.username))

@step()
def caaspctl_cluster_status():
    caaspctl("${WORKSPACE}/go/src/suse.com/caaspctl/test-cluster",
        "cluster status")

@step()
def caaspctl_node_join(role="worker", nr=0):
    if role == "master":
        ip_addr = get_masters_ipaddrs()[nr]
    else:
        ip_addr = get_workers_ipaddrs()[nr]

    caaspctl("${WORKSPACE}/go/src/suse.com/caaspctl/test-cluster",
        "node join --role {role} --user {username} --sudo --target "
          "{ip} my-{role}-{nr}".format(role=role, ip=ip_addr, nr=nr, username=conf.username))

def pick_ssh_agent_sock():
    return os.path.join(replace_vars("${WORKSPACE}"), "ssh-agent-sock")

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
    key_fn = locate_id_shared()
    sh("ssh-add " + key_fn, env={"SSH_AUTH_SOCK": sock_fn})
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
    with open(replace_vars("${WORKSPACE}/environment.json")) as f:
        return json.load(f)


@step()
def setup_testinfra_tox(env, cmd):
    shp("${WORKSPACE}/automation/testinfra", cmd, env=env)

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
    sh("mkdir -p ${WORKSPACE}/artifacts")
    path = os.path.join(path, glob)
    try:
        sh("rsync -a " + path + " ${WORKSPACE}/artifacts")
    except:
        print("rsync error")

@step()
def archive_logs():
    """Archive Logs"""
    archive_artifacts('${WORKSPACE}', 'logs/**')

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
def _cleanup_openstack_terraform():
    shp("caaspctl/ci/infra/openstack", "source ${OPENRC};"
        " terraform destroy -auto-approve"
        " -var internal_net=net-${JOB_NAME}-${BUILD_NUMBER}"
        " -var stack_name=${JOB_NAME}-${BUILD_NUMBER}")

@step()
def cleanup_openstack_terraform():
    """Cleanup Openstack (twice)"""
    _cleanup_openstack_terraform()
    _cleanup_openstack_terraform()
    push_tfstate("openstack")


@timeout(40)
@step()
def cleanup_bare_metal():
    shp('caaspctl/ci/infra/bare-metal/deployer',
        './deployer --release ${JOB_NAME}-${BUILD_NUMBER}'
             " --conffile deployer.conf.json")

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
    conf.stack_type = 'caasp-kvm'
    conf.stage = None   # None: run all stages
    conf.change_author = ""
    conf.no_checkout = False
    conf.no_collab_check = False
    conf.no_destroy = False
    conf.workers = "3"
    conf.job_name = getvar("JOB_NAME")
    conf.build_number = getvar("BUILD_NUMBER")
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
    conf.image = replace_vars("file://${WORKSPACE}/automation/downloads/kvm-devel")
    conf.generate_pipeline = False
    conf.username = "sles"

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

image_name = "SLES15-SP1-JeOS.x86_64-15.1-OpenStack-Cloud-RC1"

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
    '''.format(job_name=run_name, authorized_keys=authorized_keys(), username=conf.username)
    return tpl


def main():
    global conf
    print("Testrunner v. {}".format(__version__))
    conf = parse_args()

    if conf.generate_pipeline:
        generate_pipeline()
        sys.exit()

    print("Using workspace: {}".format(getvar("WORKSPACE")))
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
