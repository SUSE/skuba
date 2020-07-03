# /etc/sysconfig/network/config
IFS=' ' read -r -a servers <<< "${name_servers}"
for server in $${servers[*]} 
do
    echo nameserver $${server} | sudo tee -a /etc/resolv.conf
done
