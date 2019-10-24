#!/bin/bash
#shellcheck disable=SC2145,SC2016
log()   { (>&1 echo -e "$@") ; }
cmd()   { log "[ CMD ] $@" ; }
info()  { log "[ INFO ] $@" ; }
error() { (>&2 echo -e "[ ERROR ] $@") ;}

if [ -z "${OS_AUTH_URL}" ] || [ -z "${OS_USERNAME}" ] || \
   [ -z "${OS_PASSWORD}" ] || [ -z "${OS_PROJECT_ID}" ] || \
   [ -z "${OS_PRIVATE_SUBNET_ID}" ] || [ -z "${OS_PUBLIC_NET_ID}" ]; then
    error '$OS_AUTH_URL $OS_USERNAME $OS_PASSWORD $OS_PROJECT_ID'
    error '$OS_PRIVATE_SUBNET_ID $OS_PUBLIC_NET_ID must be specified'
    error 'Please download and source your OpenStack RC file'
    exit 1
fi

OPENSTACK_CONF="openstack.conf"

umask 077

cat << EOF > "${OPENSTACK_CONF}"
[Global]
auth-url="${OS_AUTH_URL}"
username="${OS_USERNAME}"
password="${OS_PASSWORD}"
tenant-id="${OS_PROJECT_ID}"
tenant-name="${OS_PROJECT_NAME}"
domain-id="${OS_USER_DOMAIN_ID}"
domain-name="${OS_USER_DOMAIN_NAME}"
region="${OS_REGION_NAME}"
ca-file="${CA_FILE}"
[LoadBalancer]
lb-version=v2
subnet-id="${OS_PRIVATE_SUBNET_ID}"
floating-network-id="${OS_PUBLIC_NET_ID}"
create-monitor=yes
monitor-delay=1m
monitor-timeout=30s
monitor-max-retries=3
[BlockStorage]
trust-device-path=false
bs-version=v2
ignore-volume-az=true
EOF

umask 022

[ -z "$OS_PROJECT_NAME" ] && sed -i '/^tenant-name=/d' "${OPENSTACK_CONF}"
[ -z "$OS_USER_DOMAIN_ID" ] &&  sed -i '/^domain-id=/d' "${OPENSTACK_CONF}"
[ -z "$OS_USER_DOMAIN_NAME" ] && sed -i '/^domain-name=/d' "${OPENSTACK_CONF}"
[ -z "$CA_FILE" ] && sed -i '/^ca-file=/d' "${OPENSTACK_CONF}"

if [ -z "${TR_STACK}" ] || [ -z "${TR_LB_IP}" ] || \
   [ -z "$TR_MASTER_IPS" ] || [ -z "$TR_WORKER_IPS" ] || \
   [ -z "${TR_USERNAME}" ]; then
    error '$TR_STACK $TR_LB_IP $TR_MASTER_IPS $TR_WORKER_IPS must be specified'
    exit 1
fi

info "### Run following commands to bootstrap skuba cluster:\\n"
cmd " skuba cluster init --control-plane ${TR_LB_IP} --cloud-provider openstack ${TR_STACK}-cluster"
cmd " mv openstack.conf ${TR_STACK}-cluster/cloud/openstack/openstack.conf"
cmd " cd ${TR_STACK}-cluster"

i=0
for MASTER in $TR_MASTER_IPS; do
    if [ $i -eq "0" ]; then
        cmd " skuba node bootstrap --target ${MASTER} --sudo --user ${TR_USERNAME} caasp-master-${TR_STACK}-0"
    else
        cmd " skuba node join --role master --target ${MASTER} --sudo --user ${TR_USERNAME} caasp-master-${TR_STACK}-${i}"
    fi
    ((++i))
done

i=0
for WORKER in $TR_WORKER_IPS; do
    cmd " skuba node join --role worker --target ${WORKER} --sudo --user ${TR_USERNAME} caasp-worker-${TR_STACK}-${i}"
    ((++i))
done
