echo "solver.onlyRequires = true" >> /etc/zypp/zypp.conf
zypper -n rm docker containerd docker-runc docker-libnetwork
zypper -n install ${packages}

# remove docker0 interface and all its iptables rules
ip link delete docker0
iptables -L | grep DOCKER | awk {'print $2'} | xargs -d "\n" -i iptables -X {}
iptables-save | awk '/^[*]/ { print $1 "\nCOMMIT" }' | iptables-restore
lsmod | egrep ^iptable_ | awk '{print $1}' | xargs -rd\\n modprobe -r
