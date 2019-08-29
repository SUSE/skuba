resource "aws_security_group" "egress" {
  description = "egress traffic"
  name        = "${var.stack_name}-egress"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.tags, map(
    "Name", "${var.stack_name}-egress",
    "Class", "SecurityGroup"))}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
