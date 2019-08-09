resource "aws_instance" "nodes" {
  ami                         = "${data.susepubliccloud_image_ids.sles15sp1_byos.ids[0]}"
  associate_public_ip_address = true
  count                       = "${var.workers}"
  instance_type               = "${var.worker_size}"
  key_name                    = "${aws_key_pair.kube.key_name}"
  source_dest_check           = false

  # TODO: remove from the public network once we have bastion hosts
  subnet_id = "${aws_subnet.public.0.id}"
  user_data = "${data.template_cloudinit_config.cfg.rendered}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-node-${count.index}",
    "Class", "Instance"))}"

  security_groups = [
    "${aws_security_group.ssh.id}",
    "${aws_security_group.icmp.id}",
    "${aws_security_group.egress.id}",
    "${aws_security_group.allow_workers_traffic.id}",
    "${aws_security_group.lbports.id}",
  ]

  lifecycle {
    create_before_destroy = true

    # ignore_changes = ["associate_public_ip_address"]
  }

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 20
    delete_on_termination = true
  }
}

resource "null_resource" "nodes_wait_cloudinit" {
  depends_on = ["aws_instance.nodes"]
  count      = "${var.workers}"

  connection {
    host  = "${element(aws_instance.nodes.*.public_ip, count.index)}"
    user  = "ec2-user"
    type  = "ssh"
    agent = true
  }

  provisioner "remote-exec" {
    inline = [
      "cloud-init status --wait > /dev/null || /bin/true",
    ]
  }
}
