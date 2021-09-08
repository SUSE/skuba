# This security group is deliberately left empty,
# it's applied only to worker nodes.
#
# This security group is the only one with the
# `kubernetes.io/cluster/<cluster name>` tag, that makes it discoverable by the
# AWS CPI controller.
# As a result of that, this is going to be the security group the CPI will
# alter to add the rules needed to access the worker nodes from the AWS
# resources dynamically provisioned by the CPI (eg: load balancers).
resource "aws_security_group" "worker" {
  description = "security group rules for worker node"
  name        = "${var.stack_name}-worker"
  vpc_id      = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-worker"
      "Class" = "SecurityGroup"
    },
  )
}

# https://www.terraform.io/docs/providers/aws/r/instance.html
resource "aws_instance" "worker" {

  count             = var.workers
  ami               = data.susepubliccloud_image_ids.sles15sp2_chost_byos.ids[0]
  instance_type     = var.worker_instance_type
  key_name          = aws_key_pair.kube.key_name
  source_dest_check = false

  availability_zone = var.aws_availability_zones[count.index % length(var.aws_availability_zones)]
  # associate_public_ip_address = false
  # subnet_id                 = aws_subnet.private[count.index % length(var.aws_availability_zones)].id
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.public[count.index % length(var.aws_availability_zones)].id
  user_data                   = data.template_cloudinit_config.cfg.rendered
  iam_instance_profile        = length(var.iam_profile_worker) == 0 ? aws_iam_instance_profile.worker.0.name : var.iam_profile_worker
  # ebs_optimized          = true

  depends_on = [
    # aws_route.private_nat_gateway,
    aws_internet_gateway.platform,
    aws_iam_instance_profile.worker,
  ]

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
    volume_size           = var.worker_volume_size
    delete_on_termination = true
  }

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-worker-${count.index}"
      "Class" = "Instance"
    },
  )
}
