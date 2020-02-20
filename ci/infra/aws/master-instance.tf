resource "aws_instance" "control_plane" {
  ami                         = data.susepubliccloud_image_ids.sles15sp1_chost_byos.ids[0]
  associate_public_ip_address = true
  count                       = var.masters
  instance_type               = var.master_size
  key_name                    = aws_key_pair.kube.key_name
  source_dest_check           = false
  subnet_id                   = aws_subnet.public.id
  user_data                   = data.template_cloudinit_config.cfg.rendered
  iam_instance_profile        = length(var.iam_profile_master) == 0 ? local.aws_iam_policy_master_terraform : var.iam_profile_master

  depends_on = [
    aws_internet_gateway.platform,
    aws_iam_policy.master,
  ]

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-master-${count.index}"
      "Class" = "Instance"
    },
  )

  vpc_security_group_ids = [
    aws_security_group.egress.id,
    aws_security_group.common.id,
    aws_security_group.master.id,
  ]

  lifecycle {
    create_before_destroy = true

    ignore_changes = [ami]
  }

  root_block_device {
    volume_type           = "gp2"
    volume_size           = 20
    delete_on_termination = true
  }
}

