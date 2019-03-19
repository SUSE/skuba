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

# flavio, rafa, salt-ssh keys
ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDJ8uCu3Uzy7phfkow2WbAGVH4pDpbfO8JWRnnHftGyRoNzYBtqxWthvtRsgcZpQ/Qxts1pWd7Zoy0fs5915du4qxhXJ5unuGrgZi29fWVbwUwP7baMF9q//HUh5oymP5YrxpPZaEdm2t9zXP/X6jTBFG2VOY2us6A/9+iff1sRg34YOFP4HoMIqXov4BK4m73SVU72VqrdeGLYy7gughK2ZB3l9QuI3tEY1Pz+uwErJO7rlR+Id4LBS2JCWuZKpvKVcocY2+0a1q829xTsLqEnHckTgbs6zcYl77CwculONfrFwEdWSXzbL+YwLig4JjR8uIVZ1IJaNBPBAssDOZPt cardno:000604160827
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAEAQDJz9rVLqUHt9ZFjep4RsN3B5xr9s6MtHSz4PbJHACj3bA3pP7UZwePzzDMofOZLhOIKzMJ+s9H0E28ruEN8xhAv9qPYN6DI15vvPoaMu4VbzyFOGAz4UXoMQpSkr3p9E8C3psJIMpgxOAGelp7PqODlCQS/6DVMqz3DqtkOJYPssAtivH1AfitA2NVPvI9bgswAhF0jArKJmnFPSy6DAc0G2q5DyVEZVfD943kOprd7GkVWdD9FpaHqjmLGd77RfHmmqrj9Tg5+ajYa+VrASJfTBkDJ/lZcFLH9DdfcUFcQzu2pzi/cX94e+FnTtog8TOGwCrWGDAZVPP5YEHmGU3QX0NQBl4vNprWe0oJaSjSzvIT2ZrixUhOWKTjW44To/+7UwIlCc4KWL/LJ+kbwWCpiOhhWqRs380cqUmuMRaq59uTIWCRImBTkqjTqBIxaj7060GV2ZWGzbYKhUsxPchx8KJVWyGxYoox+T/zQjF1KtwvnPVghIt4hiIifYCclxoeY1yAIU5T8LvZXaqBlSYPi715mJg7533IM6NhHMg09ANgkKt6fQmQUNtaYBpHfaIaKI68oSCJOFTiP3e1RYmKaz36GQPWqEBNKT5zaIYsSOMCyLhoecH6pF9Nqvust5iIpYgNSDlRh1qnOd1AUCimyJQiswsiEQTuCClbZHg152x33/6y8CZrpHRSzDh8cBApanvtQ5pmzD4IP9mZ3eGvWaSrVx6EtpYWkr4LSoPkh2dRWHdVu+a27TLVkl7V+2dE5WAIZzRsfpAfQB3JIVD5WmTVlbU1zgIIBSXr7SfGJo0bMQ59JptE9+ffoyGWk8fnbFww2re3QTphXau9Hy+88pUqvXkiYUxsSpHzXlpRAWbfR9wqCS3adKRaz+3vZYvJGP6d66ay9NRkTGeIKxEeYjdBSNues59UGsWiJVOaR9bxfvL5+F+WyIjv9a9yOJln9NXcADp32zUlAMY97+Kw0NRQeBnpX2fF6HjNLj+onlOt50EVNYGE3fS5CW8L9nSuIY4jAycQ8xF2GYG8lGgDfaCrGTVm0cFab6ytvLRFBaKFWqcIh2rYOgKV0p7qzoadQYH5hIH/V5LGt3yRgPdwqGHNd6662n5FKlio/omE1CUpAdedA1l+geMnaIdQpDG5ghjSb8jJnoUsPYVjTLQmg7g2HAnC2ofURbKAxEVfDDIuXmLp+plyb7DGGIhj6wprnoy7mDd/YJBzf9zmRjOz1mKhrgdbSHiDvfpbs0BW5HtodYHY7R6oEU8OtXYOR2bJfoqhspz5M/vmYBbzo5P7cbpBc5b6PW/xFnt2Sabuwrem0YTWh++eDmeDgSOK5F9k4NGQZriJYg5JqICqslht
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDCcQ4vGwNYRd1Pyx4I1tbLZe+wNdJ6qFXDrd1eJIM4uY528S/sljpkrht4W+yRW5gNYvJ05Dg5qj9JN9HRtP8uzXgMP5iafsOOUz4QhpI28RVcwdamQzJu6z7EiF/8A6gcUFsBqf5cB/N6zif91oJ1cfWtG8fJwUHVxUGkfQpOW3/tNNnmnrNJwO9c4aKXhwm/+uZDzk9KDAGgKgOS2vhqB2FV+nQ5bACUeDvK1+gTVFuHNQL+x8ha2H8ak7Huz3IBXnrN+mQOioJVAOcnzjlxKUvvfhc8zn0AxxcpDntoWfYvuyUyAHIuPjI5LfjKuILiLGNrHDbA8rjCMGGxquH1 root@freedom

# need to disable gpg checks because the cloud image has an untrusted repo
zypper:
  repos:
    - id: caasp
      name: caasp
      baseurl: ${repo_baseurl}
      enabled: 1
      autorefresh: 1
      gpgcheck: 0
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
  - kubectl
  - cni-plugins
  - "-docker"
  - "-containerd"
  - "-docker-runc"
  - "-docker-libnetwork"

bootcmd:
  - ip link set dev eth0 mtu 1400

final_message: "The system is finally up, after $UPTIME seconds"
