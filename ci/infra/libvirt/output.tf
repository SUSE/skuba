output "ip_load_balancer" {
  value = var.create_lb ? "{ \"${libvirt_domain.lb.0.network_interface.0.hostname}\" = \"${libvirt_domain.lb.0.network_interface.0.addresses.0}\" }" : "not created"
}

output "ip_masters" {
  value = zipmap(
    libvirt_domain.master.*.network_interface.0.hostname,
    libvirt_domain.master.*.network_interface.0.addresses.0,
  )
}

output "ip_workers" {
  value = zipmap(
    libvirt_domain.worker.*.network_interface.0.hostname,
    libvirt_domain.worker.*.network_interface.0.addresses.0,
  )
}
