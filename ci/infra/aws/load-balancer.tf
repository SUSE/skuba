# A security group for the ELB so it is accessible via the web
resource "aws_security_group" "elb" {
  name        = "${var.stack_name}-elb"
  description = "give access to kube api server"
  vpc_id      = aws_vpc.platform.id

  tags = merge(
    local.tags,
    {
      "Name"  = "${var.stack_name}-elb"
      "Class" = "SecurityGroup"
    },
  )

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

}

# https://www.terraform.io/docs/providers/aws/r/elb.html
resource "aws_elb" "elb" {
  name = "${var.stack_name}-elb"
  # https://www.terraform.io/docs/providers/aws/r/elb.html
  # Exactly one of availability_zones or subnets must be specified: this determines if the ELB exists in a VPC or in EC2-classic.
  # availability_zones        = tolist(aws_subnet.public.*.availability_zone)
  subnets                   = tolist(aws_subnet.public.*.id)
  instances                 = aws_instance.master.*.id
  cross_zone_load_balancing = true
  idle_timeout              = 400
  connection_draining       = false
  # connection_draining_timeout = 400
  security_groups = [
    aws_security_group.elb.id,
    aws_security_group.egress.id,
  ]

  # kube
  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  # dex - protocol is set to tcp instead of https. Otherwise
  # we would have to create the SSL certificate right now
  listener {
    instance_port     = 32000
    instance_protocol = "tcp"
    lb_port           = 32000
    lb_protocol       = "tcp"
  }

  # gangway - protocol is set to tcp instead of https. Otherwise
  # we would have to create the SSL certificate right now
  listener {
    instance_port     = 32001
    instance_protocol = "tcp"
    lb_port           = 32001
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    interval            = 30
    target              = "TCP:6443"
    timeout             = 3
    unhealthy_threshold = 6
  }

  tags = merge(
    local.basic_tags,
    {
      Name    = "${var.stack_name}-elb"
      "Class" = "ElasticLoadBalancer"
    },
  )
}
