#cloud-config

# set locale
locale: en_GB.UTF-8

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

# FIXME: zypper repos do not seem to work on the cloud-init in the AWS AMIs
#
# # need to disable gpg checks because the cloud image has an untrusted repo
# zypper:
#   repos:
# \$\{repositories}
#   config:
#     gpgcheck: "off"
#     solver.onlyRequires: "true"
#     download.use_deltarpm: "true"
# 
# # need to remove the standard docker packages that are pre-installed on the
# # cloud image because they conflict with the kubic- ones that are pulled by
# # the kubernetes packages
# packages:
# \$\{packages}

bootcmd:
  - ip link set dev eth0 mtu 1400

runcmd:
  - zypper ar --enable --refresh --no-gpgcheck https://download.opensuse.org/repositories/devel:/kubic/openSUSE_Leap_15.1 kubic
  - zypper ref
  - zypper in -y --no-recommends kubernetes-kubeadm kubernetes-kubelet kubernetes-client cri-o cni-plugins kmod
  - /usr/bin/sed -i -e 's/btrfs/overlay2/g' /etc/crio/crio.conf

final_message: "The system is finally up, after $UPTIME seconds"


