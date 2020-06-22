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

# WARNING!!! Do not use cloud-init packages module when SUSE CaaSP Registraion
# Code is provided. In this case repositories will be added in runcmd module
# with SUSEConnect command after packages module is ran
#packages:

# set hostname
hostname: ${hostname}

runcmd:
  # Set node's hostname from DHCP server
  - sed -i -e '/^DHCLIENT_SET_HOSTNAME/s/^.*$/DHCLIENT_SET_HOSTNAME=\"${hostname_from_dhcp}\"/' /etc/sysconfig/network/dhcp
  - systemctl restart wicked
${register_scc}
${register_rmt}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"
