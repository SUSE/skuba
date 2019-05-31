from platforms.openstack import Openstack
from platforms.vmware import VMware
from utils import Format

class Platform:

    @staticmethod    
    def get_platform(conf):
        if conf.platform == "openstack":
            platform = Openstack(conf)
        elif conf.platform == "vmware":
            platform = VMware(conf)
        elif conf.platform == "bare-metal":
            # TODO platform = Bare_metal(conf, utils)
            print("Todo: bare-metal is not ready yet")
            sys.exit(0)
        elif conf.platform == "libvirt":
            # TODO platform = Livbirt(conf, utils)
            print("Todo: libvirt is not ready yet")
            sys.exit(0)
        else:
            raise Exception(Format.alert("Platform Error: {} is not applicable".format(conf.platform)))

        return platform

