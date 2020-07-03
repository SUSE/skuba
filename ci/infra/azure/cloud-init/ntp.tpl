sudo rm -f /etc/chrony.d/ntp_customized.conf
IFS=' ' read -r -a servers <<< "${ntp_servers}"
for server in $${servers[*]} 
do
    echo "server $${server} iburst" | sudo tee -a /etc/chrony.d/ntp_customized.conf
done
