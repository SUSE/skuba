output "control_plane.public_ip" {
  value = "${aws_instance.control_plane.*.public_ip}"
}

output "control_plane.private_dns" {
  value = "${aws_instance.control_plane.*.private_dns}"
}

output "nodes.public_ip" {
  value = "${aws_instance.nodes.*.public_ip}"
}

output "nodes.private_dns" {
  value = "${aws_instance.nodes.*.private_dns}"
}
