output "ip_masters" {
  value = "${zipmap(openstack_compute_instance_v2.master.*.name, openstack_networking_floatingip_v2.master_ext.*.address)}"
}

output "ip_workers" {
  value = "${zipmap(openstack_compute_instance_v2.worker.*.name, openstack_networking_floatingip_v2.worker_ext.*.address)}"
}

output "ip_internal_load_balancer" {
  value = "${openstack_lb_loadbalancer_v2.lb.vip_address}"
}

output "ip_load_balancer" {
  value = "${zipmap(list(openstack_lb_loadbalancer_v2.lb.name), list(openstack_networking_floatingip_v2.lb_ext.address))}"
}
