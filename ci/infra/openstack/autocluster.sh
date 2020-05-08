#!/bin/bash

action=$1
version=$2

folder="cluster-vzm"
state="vzm.tfstate"

function bootstrap() {
    ssh-add /home/vzepedamas/.ssh/id_shared
    source /home/vzepedamas/Downloads/open-cloud.sh
    terraform init
    terraform plan
    terraform apply -auto-approve -state=$state
    outputs=$(cat $state | jq ".outputs")

    loadbalancer=$(echo $outputs | jq -r ".ip_load_balancer.value[]")
    mips=$(echo $outputs | jq -r ".ip_masters.value[]")
    wips=$(echo $outputs | jq -r ".ip_workers.value[]")

    masterips=()
    index=0
    for item in $mips; do
        masterips[$index]=$item
        let "index=$index+1"
    done

    workerips=()
    index=0
    for item in $mips; do
        workerips[$index]=$item
        let "index=$index+1"
    done


    echo "LB $loadbalancer"
    echo "Init"

    if [ -n "$version" ]; then
        skuba cluster init $folder --control-plane $loadbalancer --kubernetes-version $version
    else
        skuba cluster init $folder --control-plane $loadbalancer
    fi

    cd "${folder}"

    echo "master0 ${masterips[0]}"
    echo "Bootstrap"

    echo "skuba node bootstrap master-1 -s -u sles -t ${masterips[0]} -v 5"
    skuba node bootstrap master-1 -s -u sles -t ${masterips[0]} -v 5

    echo "Joining masters"
    index=2
    for item in ${masterips[@]:1}; do
        echo "skuba node join master-$index -r master -s -u sles -t $item -v 5"
        skuba node join master-$index -r master -s -u sles -t $item -v 5
        let "index=$index+1"
    done

    echo "Joining workers"
    index=1
    for item in ${workerips[@]}; do
        echo "skuba node join worker-$index -r worker -s -u sles -t $item -v 5"
        skuba node join worker-$index -r worker -s -u sles -t $item -v 5
        let "index=$index+1"
    done

    skuba cluster status
}

function cleanup() {
    rm -rf $folder
    terraform destroy -auto-approve -state=$state
    exit
}

if [ "$action" == "cleanup" ]; then {
    echo "CLEANUP"
    cleanup
}
elif [ "$action" == "bootstrap" ]; then {
    echo "BOOTSTRAP"
    bootstrap
}
fi