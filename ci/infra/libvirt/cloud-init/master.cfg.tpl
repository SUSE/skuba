#cloud-config

# set locale
locale: en_GB.UTF-8

# set timezone
timezone: Etc/UTC

# Set hostname and FQDN
hostname: ${hostname}
fqdn: ${fqdn}

# set root password
chpasswd:
  list: |
    root:linux
    opensuse:linux
  expire: False

ssh_authorized_keys:
${authorized_keys}

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
  - salt-ssh
  - kubernetes-kubeadm
  - kubernetes-kubelet
  - kubernetes-client
  - cri-o
  - cri-tools
  - cni-plugins
  - "-docker"
  - "-containerd"
  - "-docker-runc"
  - "-docker-libnetwork"
${packages}

bootcmd:
  - ip link set dev eth0 mtu 1400

# NOTE: Uncomment the following code if the CRI-O that you are using is
# expecting to be working on top of BTRFS. This may happen in packages from
# repositories which haven't fixed this issue.
#runcmd:
  #- /usr/bin/sed -i -e 's/btrfs/overlay2/g' /etc/crio/crio.conf

final_message: "The system is finally up, after $UPTIME seconds"
