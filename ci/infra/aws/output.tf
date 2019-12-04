output "control_plane_public_ip" {
  value = "${zipmap(aws_instance.control_plane.*.id, aws_instance.control_plane.*.public_ip)}"
}

output "control_plane_private_dns" {
  value = "${zipmap(aws_instance.control_plane.*.id, aws_instance.control_plane.*.private_dns)}"
}

output "nodes_private_dns" {
  value = "${zipmap(aws_instance.nodes.*.id, aws_instance.nodes.*.private_dns)}"
}
