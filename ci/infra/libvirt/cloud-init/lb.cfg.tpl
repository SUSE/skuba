#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

# Set FQDN
fqdn: ${fqdn}

# set root password
chpasswd:
  list: |
    root:linux
    ${username}:${password}
  expire: False

ssh_authorized_keys:
${authorized_keys}

bootcmd:
  - ip link set dev eth0 mtu 1400

# need to disable gpg checks because the cloud image has an untrusted repo
zypper:
  repos:
${repositories}
  config:
    gpgcheck: "off"
    solver.onlyRequires: "true"
    download.use_deltarpm: "true"

packages:
  - haproxy

write_files:
- path: /etc/haproxy/haproxy.cfg
  content: |
    # Used high values so "kubectl exec" or "kubectl logs -f"
    # do not exit due to inactivity in the connection
    defaults
      timeout connect 10s
      timeout client 86400s
      timeout server 86400s

    # NAT access: ssh -L 9000:10.17.1.0:9000 $VM_HOST
    listen stats
      bind    *:9000
      mode    http
      stats   uri       /stats

    frontend apiserver
      bind :6443
      default_backend apiserver-backend

    backend apiserver-backend
      option httpchk GET /healthz
      ${backends}

runcmd:
  # Since we are currently inside of the cloud-init systemd unit, trying to
  # start another service by either `enable --now` or `start` will create a
  # deadlock. Instead, we have to use the `--no-block-` flag.
  - [ systemctl, enable, --now, --no-block, haproxy ]

final_message: "The system is finally up, after $UPTIME seconds"
