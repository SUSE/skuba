from timeout_decorator import timeout
from terraform import Terraform


class Openstack(Terraform):
    def __init__(self, conf):
        self.osconf = conf.openstack
        Terraform.__init__(self, conf)

    def _env_setup_cmd(self):
        return "source {openrc}".format(openrc=self.conf.openstack.openrc)

    @timeout(600)
    def _cleanup_platform(self):
        # TODO: check why (and if) the following two commands are needed
        cmd = 'mkdir -p {}/logs'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        cmd = 'chmod a+x {}'.format(self.conf.workspace)
        self.utils.runshellcommand(cmd)

        cmd = ("source {openrc};"
               " terraform destroy -auto-approve"
               " -var internal_net=net-{run}"
               " -var stack_name={run}".format(
                   openrc=self.conf.openstack.openrc,
                   run=self.conf.jenkins.run_name))

        self.runshellcommandterraform(cmd)
