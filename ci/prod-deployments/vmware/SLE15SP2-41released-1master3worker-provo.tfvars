#####################
# libvirt variables #
#####################
# A remote host running libvirt
# eg: an host with more than 8GB or RAM
libvirt_uri = "qemu:///system"

# A custom pool defined on the libvirt host
# eg: a pool backed by a fast SSD
pool = "default"

#####################
# Cluster variables #
#####################
master_count = 2
worker_count = 1

# A range that doesn't conflict with the SUSE network
# and allows a similar naming.
#network = "172.30.0.0/22"
#net_mode = "route"

img_source_url = "openSUSE-Leap-15.0-OpenStack.x86_64-0.0.4-Buildlp150.12.113.qcow2"

# define the base url of the repository to use
# repo_baseurl = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
