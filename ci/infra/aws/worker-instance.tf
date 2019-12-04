resource "aws_instance" "nodes" {
  ami                         = data.susepubliccloud_image_ids.sles15sp1_chost_byos.ids[0]
  associate_public_ip_address = false
  count                       = var.workers
  instance_type               = var.worker_size
  key_name                    = aws_key_pair.kube.key_name
  source_dest_check           = false
  user_data                   = data.template_cloudinit_config.cfg.rendered
  iam_instance_profile        = length(var.iam_profile_worker) == 0 ? local.aws_iam_policy_worker_terraform : var.iam_profile_worker
  subnet_id                   = aws_subnet.private.id

  depends_on = [
    aws_route.private_nat_gateway,
    aws_iam_policy.worker,
  ]

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-node-${count.index}"
      "Class" = "Instance"
    },
  )

  vpc_security_group_ids = [
    aws_security_group.egress.id,
    aws_security_group.common.id,
    aws_security_group.worker.id,
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

