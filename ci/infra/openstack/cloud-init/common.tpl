#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

# Inject the public keys
ssh_authorized_keys:
${authorized_keys}

ntp:
  enabled: true
  ntp_client: chrony
  config:
    confpath: /etc/chrony.conf
  servers:
${ntp_servers}

# need to disable gpg checks because the cloud image has an untrusted repo
zypper:
  repos:
${repositories}
  config:
    gpgcheck: "off"
    solver.onlyRequires: "true"
    download.use_deltarpm: "true"

# need to remove the standard docker packages that are pre-installed on the
# cloud image because they conflict with the kubic- ones that are pulled by
# the kubernetes packages
# WARNING!!! Do not use cloud-init packages module when SUSE CaaSP Registraion
# Code is provided. In this case repositories will be added in runcmd module 
# with SUSEConnect command after packages module is ran
#packages:

bootcmd:
  - ip link set dev eth0 mtu 1400

runcmd:
  # workaround for bsc#1119397 . If this is not called, /etc/resolv.conf is empty
  - netconfig -f update
  # Workaround for bsc#1138557 . Disable root and password SSH login
  - sed -i -e '/^PermitRootLogin/s/^.*$/PermitRootLogin no/' /etc/ssh/sshd_config
  - sed -i -e '/^#ChallengeResponseAuthentication/s/^.*$/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config
  - sed -i -e '/^#PasswordAuthentication/s/^.*$/PasswordAuthentication no/' /etc/ssh/sshd_config
  - sshd -t || echo "ssh syntax failure"
  - systemctl restart sshd
${register_scc}
${register_rmt}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"
