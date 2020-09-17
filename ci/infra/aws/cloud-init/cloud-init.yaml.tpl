#cloud-config
# vim: syntax=yaml
#
# ***********************
#   ---- for more examples look at: ------
# ---> https://cloudinit.readthedocs.io/en/latest/topics/examples.html
# ---> https://www.terraform.io/docs/providers/template/d/cloudinit_config.html
# ******************************
#
# This is the configuration syntax that the write_files module
# will know how to understand. encoding can be given b64 or gzip or (gz+b64).
# The content will be decoded accordingly and then written to the path that is
# provided.
#
# Note: Content strings here are truncated for example purposes.

# set locale
locale: en_US.UTF-8

# set timezone
timezone: Etc/UTC

users:
  - name: ${username}
    ssh-authorized-keys:
      ${authorized_keys}
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    groups: sudo
    shell: /bin/bash

# Inject the public keys
ssh_authorized_keys:
${authorized_keys}

# WARNING!!! Do not use cloud-init packages module when SUSE CaaSP Registration
# Code is provided. In this case, repositories will be added in runcmd module
# with SUSEConnect command after packages module is ran
#packages:

bootcmd:
  - ip link set dev eth0 mtu 1500

runcmd:
${register_scc}
${register_rmt}
${register_suma}
${repositories}
${commands}

final_message: "The system is finally up, after $UPTIME seconds"
