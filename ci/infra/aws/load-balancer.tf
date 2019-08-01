resource "aws_elb" "kube_api" {
  connection_draining       = false
  cross_zone_load_balancing = true
  idle_timeout              = 400
  instances                 = ["${aws_instance.control_plane.id}"]
  name                      = "${var.stack_name}-elb"
  security_groups           = ["${aws_security_group.elb.id}"]
  subnets                   = ["${aws_subnet.public.0.id}"]

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  health_check {
    healthy_threshold   = 2
    interval            = 30
    target              = "TCP:6443"
    timeout             = 3
    unhealthy_threshold = 6
  }
}

output "elb_address" {
  value = "${aws_elb.kube_api.dns_name}"
}
