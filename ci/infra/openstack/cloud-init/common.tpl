#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

# set root password
chpasswd:
  list: |
    root:linux
    ${username}:${password}
  expire: False

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
${register_scc}
${register_rmt}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"
