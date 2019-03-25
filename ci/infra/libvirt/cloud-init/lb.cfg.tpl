#cloud-config

# set locale
locale: en_GB.UTF-8

# set timezone
timezone: Etc/UTC

# Set hostname and FQDN
hostname: ${hostname}
fqdn: ${fqdn}

# set root password
chpasswd:
  list: |
    root:linux
    opensuse:linux
  expire: False

ssh_authorized_keys:
${authorized_keys}

bootcmd:
  - ip link set dev eth0 mtu 1400

# need to disable gpg checks because the cloud image has an untrusted repo
zypper:
  repos:
    - id: caasp
      name: caasp
      baseurl: https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0
      enabled: 1
      autorefresh: 1
      gpgcheck: 0
  config:
    gpgcheck: "off"
    solver.onlyRequires: "true"
    download.use_deltarpm: "true"

packages:
  - haproxy

write_files:
- path: /etc/haproxy/haproxy.cfg
  content: |
    global
      log /dev/log daemon
      maxconn 32768
      chroot /var/lib/haproxy
      user haproxy
      group haproxy
      daemon
      stats socket /var/lib/haproxy/stats user haproxy group haproxy mode 0640 level operator
      tune.bufsize 32768
      tune.ssl.default-dh-param 2048
      ssl-default-bind-ciphers ALL:!aNULL:!eNULL:!EXPORT:!DES:!3DES:!MD5:!PSK:!RC4:!ADH:!LOW@STRENGTH

    defaults
      log     global
      mode    tcp
      option  log-health-checks
      option  log-separate-errors
      option  dontlog-normal
      option  dontlognull
      option  httplog
      option  socket-stats
      retries 3
      option  redispatch
      maxconn 10000
      timeout connect     5s
      timeout client     50s
      timeout server    450s

    frontend apiserver
      bind 0.0.0.0:6443
      default_backend apiserver-backend

    backend apiserver-backend
      option httpchk GET /healthz
      default-server inter 3s fall 3 rise 2
      ${backends}

runcmd:
  - [ systemctl, enable, haproxy ]

final_message: "The system is finally up, after $UPTIME seconds"
