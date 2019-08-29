resource "aws_security_group" "master" {
  description = "security rules for master nodes"
  name        = "${var.stack_name}-master"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-control-plane",
    "Class", "SecurityGroup"))}"

  # etcd - internal
  ingress {
    from_port   = 2379
    to_port     = 2380
    protocol    = "tcp"
    cidr_blocks = ["${var.vpc_cidr_block}"]
  }

  # api-server - everywhere
  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
