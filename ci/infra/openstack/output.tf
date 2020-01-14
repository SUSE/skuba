output "hostnames_masters" {
  value = openstack_dns_recordset_v2.master.*.name
}

output "ip_masters" {
  value = [openstack_networking_floatingip_v2.master_ext.*.address]
}

output "ip_workers" {
  value = [openstack_networking_floatingip_v2.worker_ext.*.address]
}

output "ip_internal_load_balancer" {
  value = openstack_lb_loadbalancer_v2.lb.vip_address
}

output "ip_load_balancer" {
  value = openstack_networking_floatingip_v2.lb_ext.address
}

