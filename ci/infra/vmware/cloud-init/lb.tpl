#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

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
packages:
  - haproxy

# set hostname
hostname: ${hostname}

runcmd:
  # Since we are currently inside of the cloud-init systemd unit, trying to
  # start another service by either `enable --now` or `start` will create a
  # deadlock. Instead, we have to use the `--no-block-` flag.
  # The template machine should have been cleaned up, so no machine-id exists
  - [ dbus-uuidgen, --ensure ]
  - [ systemd-machine-id-setup ]
  # With a new machine-id generated the journald daemon will work and can be restarted
  # Without a new machine-id it should be in a failed state
  - [ systemctl, restart, systemd-journald ]
  # LB firewall is not disabled via skuba, so we need to manually disable it
  - [ systemctl, disable, firewalld, --now ]
  # Set node's hostname from DHCP server
  - sed -i -e '/^DHCLIENT_SET_HOSTNAME/s/^.*$/DHCLIENT_SET_HOSTNAME=\"${hostname_from_dhcp}\"/' /etc/sysconfig/network/dhcp
  - systemctl restart wicked

final_message: "The system is finally up, after $UPTIME seconds"
