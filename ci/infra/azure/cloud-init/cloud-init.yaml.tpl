#cloud-config

# Set locale
locale: en_US.UTF-8

# Set timezone
timezone: Etc/UTC

# Inject the public keys
ssh_authorized_keys:
${authorized_keys}

# WARNING!!! Do not use cloud-init packages module when SUSE CaaSP Registration
# Code is provided. In this case, repositories will be added in runcmd module
# with SUSEConnect command after packages module is ran
# packages:

bootcmd:
  - ip link set dev eth0 mtu 1500

runcmd:
${register_scc}
${register_rmt}
${register_suma}
${repositories}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"
