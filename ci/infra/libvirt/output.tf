output "ip_lb" {
  value = "${libvirt_domain.lb.network_interface.0.addresses.0}"
}

output "masters" {
  value = ["${libvirt_domain.master.*.network_interface.0.addresses.0}"]
}

output "workers" {
  value = ["${libvirt_domain.worker.*.network_interface.0.addresses.0}"]
}
