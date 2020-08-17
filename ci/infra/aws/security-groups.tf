resource "aws_security_group" "common" {
  description = "common security group rules for master and worker nodes"
  name        = "${var.stack_name}-common"
  vpc_id      = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-common"
      "Class" = "SecurityGroup"
    },
  )

  # Allow ICMP
  ingress {
    from_port       = -1
    to_port         = -1
    protocol        = "icmp"
    security_groups = []
    self            = true
    description     = "allow ICPM traffic ingress"
  }

  egress {
    from_port       = -1
    to_port         = -1
    protocol        = "icmp"
    security_groups = []
    cidr_blocks     = [var.cidr_block]
    description     = "allow ICPM traffic egress"
  }

  # Allow ssh from anywhere
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "allow ssh from everywhere"
  }

  # cilium - health check - internal
  ingress {
    from_port   = 4240
    to_port     = 4240
    protocol    = "tcp"
    cidr_blocks = [var.cidr_block]
    description = "cilium - health check - internal"
  }

  # cilium - VXLAN traffic - internal
  ingress {
    from_port   = 8472
    to_port     = 8472
    protocol    = "udp"
    cidr_blocks = [var.cidr_block]
    description = "cilium - VXLAN traffic - internal"
  }

  # master -> worker kubelet communication - internal
  ingress {
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    cidr_blocks = [var.cidr_block]
    description = "master to worker kubelet communication - internal"
  }

  # kubeproxy health check - internal only
  ingress {
    from_port   = 10256
    to_port     = 10256
    protocol    = "tcp"
    cidr_blocks = [var.cidr_block]
    description = "kubeproxy health check - internal only"
  }

  # range of ports used by kubernetes when allocating services
  # of type `NodePort` - internal
  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "kubernetes NodePort services"
  }

  # range of ports used by kubernetes when allocating services
  # of type `NodePort` - internal
  ingress {
    from_port   = 30000
    to_port     = 32767
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "kubernetes NodePort services"
  }
}

resource "aws_security_group" "egress" {
  description = "egress traffic"
  name        = "${var.stack_name}-egress"
  vpc_id      = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-egress"
      "Class" = "SecurityGroup"
    },
  )

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# resource "aws_security_group" "default" {
#   count       = local.security_group_count
#   name        = module.label.id
#   vpc_id      = var.vpc_id
#   description = "Instance default security group (only egress access is allowed)"
#   tags        = module.label.tags

#   lifecycle {
#     create_before_destroy = true
#   }
# }

# resource "aws_security_group_rule" "egress" {
#   count             = var.create_default_security_group ? 1 : 0
#   type              = "egress"
#   from_port         = 0
#   to_port           = 65535
#   protocol          = "-1"
#   cidr_blocks       = ["0.0.0.0/0"]
#   security_group_id = aws_security_group.default[0].id
# }

# resource "aws_security_group_rule" "ingress" {
#   count             = var.create_default_security_group ? length(compact(var.allowed_ports)) : 0
#   type              = "ingress"
#   from_port         = var.allowed_ports[count.index]
#   to_port           = var.allowed_ports[count.index]
#   protocol          = "tcp"
#   cidr_blocks       = ["0.0.0.0/0"]
#   security_group_id = aws_security_group.default[0].id
# }
