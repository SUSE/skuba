#cloud-config

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

ssh_authorized_keys:
${authorized_keys}

# FIXME: zypper repos do not seem to work on the cloud-init in the AWS AMIs

bootcmd:
  - ip link set dev eth0 mtu 1400

runcmd:
${register_scc}
${register_rmt}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"

