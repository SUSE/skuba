data "template_file" "master-cloud-init" {
  template = "${file("cloud-init/common.tpl")}"

  vars {
    authorized_keys = "${join("\n", formatlist("  - %s", var.authorized_keys))}"
    repo_baseurl    = "${var.repo_baseurl}"
  }
}

resource "aws_instance" "master" {
  count                  = "${var.masters}"
  ami                    = "${data.aws_ami.latest_ami.id}"
  instance_type          = "${var.master_size}"
  subnet_id              = "${element(aws_subnet.public.*.id, count.index)}"
  user_data              = "${data.template_file.master-cloud-init.rendered}"
  vpc_security_group_ids = ["${aws_security_group.kubernetes.id}"]

  tags = {
    Name = "ag-master-${var.stack_name}-${count.index}"
  }
}
