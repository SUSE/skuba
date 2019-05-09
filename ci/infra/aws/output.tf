output "ip_masters" {
  value = ["${aws_instance.master.*.public_ip}"]
}

output "ip_workers" {
  value = ["${aws_instance.worker.*.public_ip}"]
}

output "lb_dns_name" {
  value = "${aws_alb.lb.dns_name}"
}

#output "lb_ips" {
#  value = ["${aws_eip.lb.*.public_ip}"]
#}

