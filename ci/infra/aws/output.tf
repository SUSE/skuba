output "control_plane.public_ip" {
  value = "${zipmap(aws_instance.control_plane.*.id, aws_instance.control_plane.*.public_ip)}"
}

output "control_plane.private_dns" {
  value = "${zipmap(aws_instance.control_plane.*.id, aws_instance.control_plane.*.private_dns)}"
}

output "nodes.private_dns" {
  value = "${zipmap(aws_instance.nodes.*.id, aws_instance.nodes.*.private_dns)}"
}
