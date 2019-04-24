# TODO show also real fqdn hostnames - maybe later over guestinfo or if variable can be filled by remote-exec

output "hostname_hint" {
  value = ["vm[last_two_ip_octets].qa.prv.suse.net"]
}

output "ip_masters" {
  value = ["${vsphere_virtual_machine.master.*.default_ip_address}"]
}

output "ip_workers" {
  value = ["${vsphere_virtual_machine.worker.*.default_ip_address}"]
}

output "ip_load_balancer" {
  value = ["${vsphere_virtual_machine.lb.*.default_ip_address}"]
}
