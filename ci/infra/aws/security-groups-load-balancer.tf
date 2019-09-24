# A security group for the ELB so it is accessible via the web
resource "aws_security_group" "elb" {
  name        = "${var.stack_name}-elb"
  description = "give access to kube api server"
  vpc_id      = "${aws_vpc.platform.id}"

  tags = "${merge(local.basic_tags, map(
    "Name", "${var.stack_name}-elb",
    "Class", "SecurityGroup"))}"

  # HTTP access from anywhere
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS access from anywhere
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "kubernetes API server"
  }

  # Allow access to dex (32000) and gangway (32001)
  ingress {
    from_port   = 32000
    to_port     = 32001
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "dex and gangway"
  }
}
