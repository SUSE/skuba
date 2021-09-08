resource "aws_security_group" "master" {
  description = "security rules for master nodes"
  name        = "${var.stack_name}-master"
  vpc_id      = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-master"
      "Class" = "SecurityGroup"
    },
  )

  # etcd - internal
  ingress {
    from_port   = 2379
    to_port     = 2380
    protocol    = "tcp"
    cidr_blocks = [var.cidr_block]
    description = "etcd"
  }

  # api-server - everywhere
  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "kubernetes api-server"
  }
}

# https://www.terraform.io/docs/providers/aws/r/instance.html
resource "aws_instance" "master" {

  count             = var.masters
  ami               = data.susepubliccloud_image_ids.sles15sp2_chost_byos.ids[0]
  instance_type     = var.master_instance_type
  key_name          = aws_key_pair.kube.key_name
  source_dest_check = false

  availability_zone = var.aws_availability_zones[count.index % length(var.aws_availability_zones)]
  # associate_public_ip_address = false
  # subnet_id                 = aws_subnet.private[count.index % length(var.aws_availability_zones)].id
  associate_public_ip_address = true
  subnet_id                   = aws_subnet.public[count.index % length(var.aws_availability_zones)].id
  user_data                   = data.template_cloudinit_config.cfg.rendered
  iam_instance_profile        = length(var.iam_profile_master) == 0 ? aws_iam_instance_profile.master.0.name : var.iam_profile_master
  # ebs_optimized          = true

  depends_on = [
    aws_internet_gateway.platform,
    aws_iam_instance_profile.master,
  ]

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
    volume_size           = var.master_volume_size
    delete_on_termination = true
  }

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-master-${count.index}"
      "Class" = "Instance"
    },
  )
}
