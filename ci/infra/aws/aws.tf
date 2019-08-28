locals {
  tags = "${merge(
    map("Name", var.stack_name,
        "Environment", var.stack_name),
    var.tags)}"
}

provider "aws" {
  region     = "${var.aws_region}"
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  profile    = "default"
}

resource "aws_key_pair" "kube" {
  key_name   = "${var.stack_name}-keypair"
  public_key = "${element(var.authorized_keys, 0)}"
}
