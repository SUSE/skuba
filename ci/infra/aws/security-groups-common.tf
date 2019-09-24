resource "aws_security_group" "common" {
  description = "common security group rules for master and worker nodes"
  name        = "${var.stack_name}-common"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.basic_tags, map(
    "Name", "${var.stack_name}-common",
    "Class", "SecurityGroup"))}"

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
    cidr_blocks     = ["${var.vpc_cidr_block}"]
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
    cidr_blocks = ["${var.vpc_cidr_block}"]
    description = "cilium - health check - internal"
  }

  # cilium - VXLAN traffic - internal
  ingress {
    from_port   = 8472
    to_port     = 8472
    protocol    = "udp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
    description = "cilium - VXLAN traffic - internal"
  }

  # master -> worker kubelet communication - internal
  ingress {
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
    description = "master to worker kubelet communication - internal"
  }

  # kubeproxy health check - internal only
  ingress {
    from_port   = 10256
    to_port     = 10256
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
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
