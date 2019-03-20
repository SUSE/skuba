#!/usr/bin/env python3 -Wd -b

from argparse import ArgumentParser
import base64
from datetime import timedelta, datetime
from time import sleep, time
import http.client
import json
import logging
import os.path
import socket
import subprocess
import sys
import urllib.parse

import yaml
from prometheus_client import CollectorRegistry, Summary, push_to_gateway
from prometheus_client import Counter

from environment_json import create_environment_json

__version__ = "0.0.12"
conf = None

class TimedFormatter(logging.Formatter):
    def format(self, record):
        t = timedelta(seconds=int(record.relativeCreated / 1000))
        record.elapsed = str(t)
        return super(TimedFormatter, self).format(record)


HWManager = None  # iLO or IPMI proxied by Bare Metal Manager

log = logging.getLogger()
log.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
handler.setFormatter(TimedFormatter('%(elapsed)s %(message)s'))
log.addHandler(handler)

# This is used to grab servers from previous deployments that never
# released their hardware and is used only when there are not enough spare
# servers available.
SERVERS_TIMEOUT = 7200

AUTOYAST_URL_PATH = "/autoyast"
AUTOYAST_SIG_CHUNK = """\
    <signature-handling>
      <accept_file_without_checksum config:type="boolean">true</accept_file_without_checksum>
      <accept_non_trusted_gpg_key config:type="boolean">true</accept_non_trusted_gpg_key>
      <accept_unknown_gpg_key config:type="boolean">true</accept_unknown_gpg_key>
      <accept_unsigned_file config:type="boolean">true</accept_unsigned_file>
      <accept_verification_failed config:type="boolean">false</accept_verification_failed>
      <import_gpg_key config:type="boolean">true</import_gpg_key>
    </signature-handling>
"""
AUTOYAST_AUTHORIZED_KEYS_CHUNK = """\
      <script>
        <chrooted config:type="boolean">true</chrooted>
        <filename>inject_authorized_key.sh</filename>
        <interpreter>shell</interpreter>
        <source><![CDATA[
#!/bin/sh
mkdir -p /root/.ssh
chmod 600 /root/.ssh
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC2G7k0zGAjd+0LzhbPcGLkdJrJ/LbLrFxtXe+LPAkrphizfRxdZpSC7Dvr5Vewrkd/kfYObiDc6v23DHxzcilVC2HGLQUNeUer/YE1mL4lnXC1M3cb4eU+vJ/Gyr9XVOOReDRDBCwouaL7IzgYNCsm0O5v2z/w9ugnRLryUY180/oIGeE/aOI1HRh6YOsIn7R3Rv55y8CYSqsbmlHWiDC6iZICZtvYLYmUmCgPX2Fg2eT+aRbAStUcUERm8h246fs1KxywdHHI/6o3E1NNIPIQ0LdzIn5aWvTCd6D511L4rf/k5zbdw/Gql0AygHBR/wnngB5gSDERLKfigzeIlCKf Unsafe Shared Key" >> /root/.ssh/authorized_keys
          ]]>
        </source>
      </script>
"""

caasp_node_pxecfg_tpl = """
default menu.c32
prompt 0
timeout 300
ONTIMEOUT CaaSP-dev-AutoYAST

MENU TITLE PXE Menu

LABEL CaaSP-dev-AutoYAST
  MENU LABEL Install CaaSP 1.0 Beta 3 (AutoYAST from {admin_host_ipaddr})
  KERNEL {tftpdir}/linux
  APPEND initrd={tftpdir}/initrd ramdisk_size=65536 install=http://{tftp_ipaddr}/distros/{tftpdir}/DVD1/ language=en_US ifcfg=*=dhcp autoyast=http://{tftp_ipaddr}/autoyast/caasp/{admin_host_ipaddr}/worker_mangled.xml insecure=1  loghost={loghost_ipaddr}

MENU INCLUDE pxelinux.cfg/local-fragment
"""

avantgarde_pxecfg_tpl = """
default menu.c32
prompt 0
timeout 30
ONTIMEOUT CaaSP-AutoYAST-avantgarde

MENU TITLE PXE Menu

LABEL CaaSP-AutoYAST-avantgarde
  MENU LABEL Install CaaSP avantgarde AutoYAST
  KERNEL sles15sp1/linux
  APPEND initrd=sles15sp1/initrd ramdisk_size=65536 install=http://10.84.44.3/distros/sles15sp1/DVD1 language=en_US netsetup=dhcp netdevice=eth0 insecure=1 autoyast=http://10.84.44.3/autoyast/avantgarde.xml loghost={loghost_ipaddr}

MENU INCLUDE pxelinux.cfg/local-fragment
"""

avantgarde_leap15_pxecfg_tpl = """
default menu.c32
prompt 0
timeout 30
ONTIMEOUT CaaSP-AutoYAST-avantgarde-leap15

MENU TITLE PXE Menu

LABEL CaaSP-AutoYAST-avantgarde-leap15
  MENU LABEL Install CaaSP avantgarde-leap15 AutoYAST
  KERNEL leap15.0/linux
  APPEND initrd=leap15.0/initrd ramdisk_size=65536 install=http://10.84.44.3/distros/leap15.0/DVD1 language=en_US netsetup=dhcp netdevice=eth0 insecure=1 autoyast=http://10.84.44.3/autoyast/avantgarde-leap15.xml loghost={loghost_ipaddr}

MENU INCLUDE pxelinux.cfg/local-fragment
"""

wipe_partition_tables_pxecfg = """
default menu.c32
timeout 30
ONTIMEOUT WipePartitionTables
MENU TITLE PXE Menu
LABEL WipePartitionTables
  MENU LABEL Wipe Partition Tables
  KERNEL hpscriptingtoolkit1040/vmlinuz
  APPEND initrd=hpscriptingtoolkit1040/initrd.img root=/dev/ram0 rw ramdisk_size=740100 ide=nodma ide=noraid pnpbios=off network=1 sstk_mount={tftp_ipaddr}:/srv/distros/hpe-scripting-toolkit-linux-10.50-41 sstk_mount_type=nfs sstk_mount_options=ro,nolock numa=off sstk_conf=toolkit.conf sstk_script=/wipe_partitions.sh
"""

# Prometheus

PROMETHEUS_ADDR = '10.84.72.105:9091'
PROMETHEUS_SSH = 'root@10.84.72.105'
PROMETHEUS_JOB_NAME = 'deploy_testbed'

prometheus_reg = CollectorRegistry()

deploy_admin_node_time = Summary(
    'deploy_testbed_admin_node',
    'Time spent doing admin node deploy',
    registry=prometheus_reg
)
deploy_nodes_time = Summary(
    'deploy_testbed_nodes',
    'Time spent doing nodes deploy',
    registry=prometheus_reg
)


tsclient = None

def make_parent_dirs(fn):
    os.makedirs(os.path.dirname(fn), exist_ok=True)

class APIError(Exception):
    pass

class TestbedServiceClient():
    """Handle interactions with the Bare Metal Manager (BMM)
    """

    def __init__(self):
        self._api = conf["bmm_api_address"] if conf else "localhost:8880"
        self._api_token = conf["bmm_token"]

    def _parse(self, response):
        """Parse JSON response
        """
        try:
            return json.loads(response)
        except Exception as e:
            if 'Internal Server Error' in response:
                for line in response.splitlines():
                    log.error(line)
                raise Exception("Server error")
            else:
                log.error("Unable to parse %r" % response)
                raise

    def _api_get(self, path):
        ctx = http.client.ssl._create_stdlib_context()
        conn = http.client.HTTPSConnection(self._api, timeout=20, context=ctx)
        log.info("calling {}/<**token**>{}".format(self._api, path))
        tpath = '/' + self._api_token + path
        # Do not leak out token in production
        # log.info("calling {}{}".format(self._api, tpath))
        try:
            conn.request('GET', tpath)
            response = conn.getresponse().read().decode('utf-8')
        except socket.timeout:
            log.info("socket timeout")
            raise APIError('testbed daemon socket timeout')
        j = self._parse(response)
        if j["status"] != "ok":
            raise APIError(response)
        return j

    def _api_get_raw(self, path):
        ctx = http.client.ssl._create_stdlib_context()
        conn = http.client.HTTPSConnection(self._api, timeout=20, context=ctx)
        log.info("calling {}/<**token**>{}".format(self._api, path))
        tpath = '/' + self._api_token + path
        try:
            conn.request('GET', tpath)
            return conn.getresponse().read()
        except socket.timeout:
            log.info("socket timeout")
            raise APIError('testbed daemon socket timeout')

    def _api_post(self, path, params):
        ctx = http.client.ssl._create_stdlib_context()
        conn = http.client.HTTPSConnection(self._api, timeout=20, context=ctx)
        params = urllib.parse.urlencode(params)
        tpath = '/' + self._api_token + path
        conn.request('POST', tpath, params)
        log.info("calling {}{}".format(self._api, path))
        response = conn.getresponse().read().decode('utf-8')
        j = self._parse(response)
        if j["status"] != "ok":
            raise APIError(response)
        return j

    def fetch_admin_node_ssh_key(self, admin_host_ipaddr):
        log.info("Grabbing admin node SSH private key")
        resp = self._api_get('/ssh/fetch_admin_node_ssh_key/{}'.format(admin_host_ipaddr))

    def fetch_machine_id(self, admin_host_ipaddr, ipaddr):
        url = '/ssh/fetch_machine_id/{}/{}'.format(admin_host_ipaddr, ipaddr)
        resp = self._api_get(url)
        return resp["machine_id"]

    def deploy_ssh_key(self, admin_host_ipaddr):
        log.info("Configuring authorized_keys on admin host")
        resp = self._api_get('/ssh/deploy_ssh_key/{}'.format(admin_host_ipaddr))

    def upload_yml_to_master(self, master_ipaddr, yml, filename):
        log.info("Uploading yml to master")
        filename = os.path.basename(filename)
        resp = self._api_post('/ssh/upload_yml_to_master/',
                {"master_ipaddr": master_ipaddr, "yml": yml, "filename":
                    filename}
        )

    def upload_pxe_conf(self, macaddr, content):
        log.info("Uploading PXE conf")
        resp = self._api_post('/pxe/upload_pxe_conf/',
                {"conf": content, "macaddr": macaddr}
        )

    def upload_worker_mangled_xml2(self, xml, urlpath):
        log.info("Uploading worker_mangled.xml to {}".format(urlpath))
        resp = self._api_post('/autoyast/upload_worker_mangled_xml2/',
                {"xml": xml, "urlpath": urlpath}
        )

    def fetch_dhcp_logs(self, from_date):
        tstamp = from_date.strftime("%s")
        resp = self._api_get('/dhcp/all/{}'.format(tstamp))
        return resp["entries"]

    def pick_tftp_dir(self, iso_list_url):
        resp = self._api_post('/iso/pick_tftp_dir', dict(iso_list_url=iso_list_url))
        return resp["tftpdir"]

    def probe_ssh_port(self, ipaddr):
        resp = self._api_get('/ssh/probe_ssh_port/{}'.format(ipaddr))
        return resp["v"]


    def fetch_servers_list(self, testname, master_count, worker_count,
            want_admin=True, want_nodes=True):
        """Fetch server list from BMM. Lock new servers if needed.

        :returns: [(servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr), ... ]
        """
        # The first server is the admin
        num_servers = 1 + master_count + worker_count
        resp = self._api_get('/hw/list/{}'.format(testname))
        servers = resp['v']
        if len(servers) == 0:
            # Lock new servers
            resp = self._api_get('/hw/lock/{}/{}/{}'.format(testname, num_servers, SERVERS_TIMEOUT))
            servers = resp['v']
            log.info("Locked {} servers".format(len(servers)))
        elif len(servers) < num_servers:
            raise Exception("Only {} servers are locked".format(len(servers)))
        elif len(servers) > num_servers:
            log.warn("Too many servers are locke: {}".format(len(servers)))

        serverlist = [
            (s['name'], s['serial'], '', s['ilo_ipaddr'], s['macaddr0'], s['macaddr1'])
            for s in servers
        ]
        assert want_admin or want_nodes
        if want_admin:
            if want_nodes:
                return serverlist
            return serverlist[:1]
        return serverlist[1:]

    def release_servers(self, testname):
        self._api_get('/hw/release/{}'.format(testname))

    def manage_iso(self):
        """Return ISO status"""
        return self._api_get('/iso/manage_iso')

    def update_iso(self, iso_list_url, iso_pattern):
        """Start ISO fetch if needed, return status"""
        return self._api_post('/iso/update_iso', dict(
                iso_list_url=iso_list_url,
                iso_pattern=iso_pattern
            ))

    def fetch_syslog_logs(self, ipaddr, from_timestamp):
        """Fetch Syslog logs from the last boot, for a host, up to a given timestamp
        """
        return self._api_get_raw('/logs/get/{}/{}'.format(ipaddr, from_timestamp))


class RemoteHWManager(TestbedServiceClient):
    def __init__(self, ipaddr, ipmi_user=None, ipmi_pass=None):
        self.ipaddr = ipaddr
        super(RemoteHWManager, self).__init__()

    def power_off(self):
        resp = self._api_get('/hosts/power_off/{}'.format(self.ipaddr))

    def power_on(self):
        resp = self._api_get('/hosts/power_on/{}'.format(self.ipaddr))

    def set_one_time_network_boot(self, check=False):
        resp = self._api_get('/hosts/set_one_time_network_boot/{}'.format(self.ipaddr))

    def get_host_power_status(self):
        resp = self._api_get('/hosts/get_host_power_status/{}'.format(self.ipaddr))
        return resp["power"]


def parse_args():
    ap = ArgumentParser(description='deploy script')
    ap.add_argument('-p', '--prometheus', action='store_true',
            help='store metrics into Prometheus')
    ap.add_argument('--conffile', help='config filename')
    ap.add_argument('--start-iso-fetching',
            help='start downloading a new ISO if available',
            action='store_true')
    ap.add_argument('--download-urls-fname',
            help='path of download-urls.json',
            default='../../misc-files/download-urls.json')
    ap.add_argument('--channel', help='ISO channel (default: devel)',
            default='devel')
    ap.add_argument('--wait-iso-fetching',
            help='start downloading any new ISO and wait for the current ISO download to complete',
            action='store_true')
    ap.add_argument('--bogus-env-json', help='create temporary environment.json', action='store_true')
    ap.add_argument('--deploy-nodes', help='deploy nodes', action='store_true')
    ap.add_argument('--velum-setup', help='setup velum user/CA',
            action='store_true')
    ap.add_argument('--velum-deploy', help='deploy from  velum',
            action='store_true')
    ap.add_argument('--prometheus-certs', help='install kube API client certs into Prometheus', action='store_true')
    ap.add_argument('--tftpdir', help='tftp dir name (default: pick latest)',
            default=None)
    ap.add_argument('testname')
    ap.add_argument('--tftp-ipaddr', help='TFTP server IP addr',
            default='10.84.44.3')  # TODO: make required parameter
    ap.add_argument('--master-count', type=int, default=1)
    ap.add_argument('--worker-count', type=int, default=2)
    ap.add_argument('--release',
            help='release locked server, if any, and exit immediately',
            action='store_true')
    ap.add_argument('--poweroff', help='power off all nodes',
            action='store_true')
    ap.add_argument('--logsdir', default=os.path.abspath("./logs"))
    args = ap.parse_args()
    args.testname = args.testname.replace('/', '-')  # TODO protect using URL encoding
    return args

def write_pxe_file(args, pxe_macaddr, cnf):
    tsclient.upload_pxe_conf(pxe_macaddr, cnf)

def fetch_and_write_syslog_logs(ipaddr, t0, fname):
    fname = os.path.join(conf['logsdir'], fname)
    log.debug("Writing {}".format(fname))
    make_parent_dirs(fname)
    syslog_logs = tsclient.fetch_syslog_logs(ipaddr, t0)
    with open(fname, 'wb') as f:
        f.write(syslog_logs)


def power_off_nodes(args):
    """Power off nodes
    """
    servers = tsclient.fetch_servers_list(args.testname, args.master_count, args.worker_count, want_admin=True, want_nodes=True)
    log.info("Powering off nodes")
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        i.power_off()

    sleep(10)
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        if i.get_host_power_status():
            log.info("Powering {} off again".format(servername))
            i.power_off()

    log.info("Powering off completed")

def parse_dhcp_logs(from_date, macaddr):
    """Parse dhcpd logs, extract last ipaddr from DHCPACK if found
    """
    macaddr = macaddr.lower()
    entries = tsclient.fetch_dhcp_logs(from_date)
    return entries.get(macaddr, None)

def wait_dhcp_acks(from_date, servers, max_failing_nodes):
    available_hosts = set()
    log.info("Waiting for DHCP ACK for the following servers:")
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        log.info("  {} {} {}".format(servername, ilo_ipaddr, eth0_macaddr))
    while True:
        entries = tsclient.fetch_dhcp_logs(from_date)
        for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
            try:
                ipaddr = entries[eth0_macaddr]
            except KeyError:
                continue
            h = (servername, serial, eth0_macaddr, ipaddr, "unknown")
            if h in available_hosts:
                continue
            log.info("Found host {} {} {} having IP address {}".format(servername, ilo_ipaddr, eth0_macaddr, ipaddr))
            available_hosts.add(h)

        if len(available_hosts) >= len(servers) - max_failing_nodes:
            break

        if len(servers) - len(available_hosts) < 3:
            srv = set(s[0] for s in servers)
            ava = set(s[0] for s in available_hosts)
            missing = srv - ava
            log.debug("Still missing: {}".format(" ".join(sorted(missing))))

        sleep(30)

    return available_hosts


@deploy_nodes_time.time()
def deploy_nodes(args, max_failing_nodes=0):
    """Deploy all nodes"""
    servers = tsclient.fetch_servers_list(
        args.testname, args.master_count, args.worker_count,
        want_admin=True, want_nodes=True)

    pxecfg = avantgarde_leap15_pxecfg_tpl.format(
        loghost_ipaddr=args.tftp_ipaddr
    )

    # Upload PXE configuration files into BMM
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        pxe_macaddr = eth0_macaddr
        write_pxe_file(args, pxe_macaddr, pxecfg)

    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        log.info("setting netboot {}".format(servername))
        try:
            i.set_one_time_network_boot()
        except Exception as e:
            log.info(e)

    sleep(10)
    rebooted = Counter('deploy_testbed_rebooted_nodes', 'Rebooted nodes',
            registry=prometheus_reg)
    power_up_time = datetime.now()
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        # assert i.get_one_time_boot() == 'network', i.get_one_time_boot()
        # assert i.get_host_power_status() == 'OFF', i.get_host_power_status()
        log.info("powering on {}".format(servername))
        i.power_on()
        rebooted.inc()

    sleep(10)

    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        if not i.get_host_power_status():
            log.warn("powering {} on again".format(servername))
            i.power_off()

    sleep(10)

    av = Counter('deploy_testbed_available_nodes', 'Available and powered up nodes',
            registry=prometheus_reg)

    not_powering_up_hosts_cnt = 0
    for servername, serial, desc, ilo_ipaddr, ilo_iface_macaddr, eth0_macaddr in servers:
        i = HWManager(ilo_ipaddr)
        if i.get_host_power_status():
            av.inc()
        else:
            log.info("BROKEN HOST {} - not powering up".format(servername))
            not_powering_up_hosts_cnt += 1

    if not_powering_up_hosts_cnt > max_failing_nodes:
        raise Exception("{} hosts not powering up".format(
            not_powering_up_hosts_cnt))

    available_hosts = wait_dhcp_acks(power_up_time, servers, max_failing_nodes)
    return available_hosts

def fetch_nodes_syslog_logs(t0, available_hosts):
    """Fetch nodes syslog loggs and store it on disk"""
    # available_hosts is a set of (servername, serial, eth0_macaddr, ipaddr), ...
    for servername, serial, eth0_macaddr, ipaddr, minionId in available_hosts:
        fn = 'worker_syslog_{}_{}_{}.log'.format(servername, serial, ipaddr)
        fetch_and_write_syslog_logs(ipaddr, t0, fn)

def install_prometheus_certs(kubeconfig):
    log.info("Fetching API client keys for Prometheus")
    y = yaml.load(kubeconfig)
    k = base64.b64decode(y['users'][0]['user']['client-key-data'])
    with open('client_key', 'w') as f:
        f.write(k.decode())
    subprocess.check_call(["/usr/bin/sudo", "scp",  "client_key",
        PROMETHEUS_SSH + ":/srv/prometheus/prometheus/kube_api_client_key"])

    c = base64.b64decode(y['users'][0]['user']['client-certificate-data'])
    with open('client_cert', 'w') as f:
        f.write(c.decode())
    subprocess.check_call(["/usr/bin/sudo", "scp", "client_cert",
        PROMETHEUS_SSH + ":/srv/prometheus/prometheus/kube_api_client_cert"])
    log.info("Reloading Prometheus")
    subprocess.check_call(["/usr/bin/sudo", "ssh", PROMETHEUS_SSH, "/usr/bin/systemctl", "reload", "prometheus.service"])

def runcmd(cmd):
    if isinstance(cmd, str):
        cmd = cmd.split()
    subprocess.check_call(cmd)

def runcmd_out(cmd):
    if isinstance(cmd, str):
        cmd = cmd.split()
    return subprocess.check_output(cmd).splitlines()


def generate_environment_json(available_hosts):
    create_environment_json(available_hosts)
    return
    # TODO
    #tsclient.deploy_ssh_key(admin_host_ipaddr)

    #tsclient.fetch_admin_node_ssh_key(admin_host_ipaddr)

    #available_hosts2 = []
    #machine_id = tsclient.fetch_machine_id(admin_host_ipaddr, admin_host_ipaddr)
    #available_hosts2.append(("admin", admin_host_ipaddr,
    #    "bogus_mac", admin_host_ipaddr, machine_id))

    #log.info("Fetching machine IDs")
    #for servername, serial, macaddr, ipaddr in available_hosts:
    #    log.info("Attempting to fetch {} {} machine ID".format(servername, ipaddr))
    #    while True:
    #        msg = ""
    #        try:
    #            machine_id = tsclient.fetch_machine_id(admin_host_ipaddr, ipaddr)
    #            available_hosts2.append((servername, serial, macaddr, ipaddr,
    #                machine_id))
    #            break
    #        except APIError as e:
    #            # log out message only when it changes
    #            if str(e) != msg:
    #                msg = str(e)
    #                log.info("The host is not ready yet: {}".format(msg))
    #            sleep(10)

    #log.info("Fetching done")

    #if use_bogus_hosts:
    #    log.info("Adding bogus hosts to environment.json to run Velum setup")
    #    available_hosts2.append(('bogus', 'bogus', 'bogus', '8.8.8.8', 'bogus'))

    #create_environment_json(admin_host_ipaddr, available_hosts2)


def run_velum_client(script_name):
    """Run velum-bootstrap from rbenv
    """
    env = dict(
        VERBOSE="true",
        ENVIRONMENT="../../environment.json",
    )
    cmd = os.path.expanduser("~/.rbenv/shims/bundle")
    cmd += " exec rspec spec/features/{}".format(script_name)
    workdir = os.path.abspath("./automation/velum-bootstrap")
    log.info("Running %r from %r", cmd, workdir)
    log.info("Using env %r", env)
    subprocess.check_call(
        cmd,
        env=env,
        cwd=workdir,
        shell=True
    )

def handle_iso(args):
    """handle fetching a new ISO if needed
    :return: tftpdir
    """
    duf = args.download_urls_fname
    log.debug("Reading %s", duf)
    with open(duf) as f:
        j = json.load(f)
        if j["format_version"] == 1:
            baseurl = j["baseurl"][args.channel]
        else:
            baseurl = j["baseurl"][args.channel]["default"]
    iso_list_url = os.path.join(baseurl, 'images/iso')

    # regexp - general enough for all Build<NNN> Media1 ISOs
    iso_pattern = 'SUSE\\-CaaS\\-Platform\\-\\d+.\\d+\\-DVD\\-x86_64\\-Build(\\d+)\\.?\\d*\\-Media1\\.iso$'

    # FIXME: this uses a different base url (the images-sle15 component) and a more restrictive pattern
    # To be moved to a dedicated "channel"
    iso_list_url = "http://download.suse.de/ibs/Devel:/CASP:/Head:/ControllerNode/images-sle15/iso"
    iso_pattern = 'SUSE\\-CaaS\\-Platform\\-\\d+.\\d+\\-DVD\\-x86_64\\-Build(\\d+)\\.\\d+\\-Media1\\.iso$'

    if args.start_iso_fetching or args.wait_iso_fetching:
        # The BMM will start fetching a new ISO, if available
        log.info("Checking for new ISO at URL: %s", iso_list_url)
        log.info("Pattern: %s", iso_pattern)
        status = tsclient.update_iso(iso_list_url, iso_pattern)
        if status["running"] is None:
            log.info("No new ISO to download")
        else:
            log.info("ISO download started - URL: {}".format(
                status["running"]))

    if args.wait_iso_fetching:
        while True:
            status = tsclient.manage_iso()
            # TODO: ignore running download for a different iso_list_url
            # TODO: handle parallel downloads
            if status["running"] is None:
                break
            log.info(
                "Waiting for ISO to finish downloading. "
                "Progress: {} URL: {} ETA: {}".format(
                    status["progress"],
                    status["running"],
                    status.get("eta", "unknown")
                )
            )
            sleep(20)

    return tsclient.pick_tftp_dir(iso_list_url)


def main():
    global HWManager, conf, tsclient
    args = parse_args()
    log.info("Testbed deploy script v {}".format(__version__))

    if os.environ.get('CONFFILE'):
        args.conffile = os.environ['CONFFILE']
        log.info("Using conffile %r" % args.conffile)
    if args.conffile:
        with open(args.conffile) as f:
            conf = json.load(f)
        assert "bmm_api_address" in conf
        log.info("BMM address: %r" % conf["bmm_api_address"])

    assert conf
    conf["logsdir"] = args.logsdir
    HWManager = RemoteHWManager
    tsclient = TestbedServiceClient()

    if args.release:
        tsclient.release_servers(args.testname)
        return

    args.tftpdir = handle_iso(args)
    log.info("TFTP dir: %r" % args.tftpdir)

    if args.poweroff:
        power_off_nodes(args)

    elif args.deploy_nodes:
        t0 = int(time())
        power_up_time = datetime.now() - timedelta(hours=4)
        power_off_nodes(args)
        available_hosts = deploy_nodes(args)
        generate_environment_json(available_hosts)
        log.info("Waiting for host to respond to SSH")
        while True:
            if tsclient.probe_ssh_port(ipaddr) == "open":
                break
            sleep(10)
        log.info("Waiting 30s")
        sleep(30)
        fetch_nodes_syslog_logs(t0, available_hosts)
        log.info("Nodes deployment - done")

    elif args.prometheus:
        log.info("Pushing metrics to Prometheus")
        try:
            push_to_gateway(PROMETHEUS_ADDR, job=PROMETHEUS_JOB_NAME,
                registry=prometheus_reg)
        except Exception as e:
            log.error("Error pushing to Prometheus", exc_info=True)

    elif args.release:
        tsclient.release(args.testname)


if __name__ == '__main__':
    main()
