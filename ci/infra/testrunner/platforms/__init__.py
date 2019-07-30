from platforms.openstack import Openstack
from platforms.vmware import VMware


def get_platform(conf, platform):
    if platform.lower() == "openstack":
        platform = Openstack(conf)
    elif platform.lower() == "vmware":
        platform = VMware(conf)
    elif platform.lower() == "bare-metal":
        raise Exception("bare-metal is not available")
    elif platform.lower() == "libvirt":
        raise Exception("libvirt is not available")
    else:
        raise Exception("Platform Error: {} is not recognized".format(platform))

    return platform
