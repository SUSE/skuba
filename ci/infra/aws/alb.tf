resource "aws_alb" "lb" {
  name                       = "${var.stack_name}-kube-lb"
  internal                   = false
  load_balancer_type         = "network"
  enable_deletion_protection = false
  subnets                    = ["${aws_subnet.public.*.id}"]

  tags = {
    Name = "${var.stack_name}"
  }
}

resource "aws_alb_target_group" "masters" {
  name     = "${var.stack_name}-target-group-masters"
  port     = 6443
  protocol = "TCP"
  vpc_id   = "${aws_vpc.main.id}"
}

resource "aws_alb_listener" "api_server" {
  load_balancer_arn = "${aws_alb.lb.arn}"
  port              = "6443"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_alb_target_group.masters.arn}"
  }
}

resource "aws_alb_target_group_attachment" "master" {
  count            = "${var.masters}"
  target_group_arn = "${aws_alb_target_group.masters.arn}"
  target_id        = "${element(aws_instance.master.*.id, count.index)}"
  port             = 6443
}
