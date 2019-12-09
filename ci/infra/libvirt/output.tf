output "hostnames_masters" {
  value = ["${libvirt_domain.master.*.network_interface.0.hostname}"]
}

output "hostnames_workers" {
  value = ["${libvirt_domain.worker.*.network_interface.0.hostname}"]
}

output "ip_load_balancer" {
  value = "${libvirt_domain.lb.network_interface.0.addresses.0}"
}

output "ip_masters" {
  value = ["${libvirt_domain.master.*.network_interface.0.addresses.0}"]
}

output "ip_workers" {
  value = ["${libvirt_domain.worker.*.network_interface.0.addresses.0}"]
}
