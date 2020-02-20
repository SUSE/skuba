output "ip_load_balancer" {
  value = zipmap(
    libvirt_domain.lb.*.name,
    libvirt_domain.lb.*.network_interface.0.addresses.0,
  )
}

output "ip_masters" {
  value = zipmap(
    libvirt_domain.master.*.name,
    libvirt_domain.master.*.network_interface.0.addresses.0,
  )
}

output "ip_workers" {
  value = zipmap(
    libvirt_domain.worker.*.name,
    libvirt_domain.worker.*.network_interface.0.addresses.0,
  )
}

