output "ip_masters" {
  value = "${zipmap(vsphere_virtual_machine.master.*.name, vsphere_virtual_machine.master.*.default_ip_address)}"
}

output "ip_workers" {
  value = "${zipmap(vsphere_virtual_machine.worker.*.name, vsphere_virtual_machine.worker.*.default_ip_address)}"
}

output "ip_load_balancer" {
  value = "${zipmap(vsphere_virtual_machine.lb.*.name, vsphere_virtual_machine.lb.*.default_ip_address)}"
}
