
# Create environment.json file - see environment.json.example

from collections import Counter
import json
import logging
import os

log = logging.getLogger(__name__)

SSH_KEY_RELPATH = "automation/misc-files/id_shared"


def create_environment_json(available_hosts):
    """
    """
    ws = os.environ.get('WORKSPACE', os.path.expanduser("~"))
    ssh_key_path = os.path.join(ws, SSH_KEY_RELPATH)

    d = {
        "sshUser" : "root",
        "sshKey" : ssh_key_path,
        "minions": []
    }
    # FIXME: this is picking a macaddr    master_ipaddr = available_hosts[1][2]
    indexes = Counter()
    for cnt, minion in enumerate(available_hosts):
        name, hw_serial, macaddr, ipaddr, machine_id = minion
        if cnt == 0:
            role = "master"
        elif cnt == 1:
            role = "worker"
            # FIXME: hardcoded 1 master 1 worker

        d["minions"].append({
           "minionId" : machine_id,
           "index" : str(indexes[role]),
           "fqdn" : hw_serial,
           "addresses" : {
              "privateIpv4" : ipaddr,
              "publicIpv4" : ipaddr,
           },
           "status" : "unused",
           "role": role,
        })
        # count up index for each role
        indexes.update((role,))


    fn = os.path.abspath('environment.json')
    with open(fn, 'w') as f:
        json.dump(d, f, indent=4, sort_keys=True)
    log.info('{} written'.format(fn))
