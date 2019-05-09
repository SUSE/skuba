#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

ssh_authorized_keys:
${authorized_keys}

# manage_resolv_conf: true
# resolv_conf:
#   nameservers: ['8.8.4.4', '8.8.8.8']

# FIXME: zypper repos do not seem to work on the cloud-init in the AWS AMIs

bootcmd:
  - ip link set dev eth0 mtu 1400

runcmd:
${register_scc}
${register_rmt}
  - zypper ref
  - zypper in -y --no-recommends ${packages}
  - /usr/bin/sed -i -e 's/btrfs/overlay2/g' /etc/crio/crio.conf

final_message: "The system is finally up, after $UPTIME seconds"

