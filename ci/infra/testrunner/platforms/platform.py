import sys

from platforms.openstack import Openstack
from platforms.vmware import VMware
from utils import Format


class Platform:

    @staticmethod    
    def get_platform(conf, platform):
        if platform.lower() == "openstack":
            platform = Openstack(conf)
        elif platform.lower() == "vmware":
            platform = VMware(conf)
        elif platform.lower() == "bare-metal":
            # TODO platform = Bare_metal(conf, utils)
            print("Todo: bare-metal is not ready yet")
            sys.exit(0)
        elif platform.lower() == "libvirt":
            # TODO platform = Livbirt(conf, utils)
            print("Todo: libvirt is not ready yet")
            sys.exit(0)
        else:
            raise Exception(Format.alert("Platform Error: {} is not applicable".format(platform)))

        return platform

