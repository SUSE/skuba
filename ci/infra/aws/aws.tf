locals {
  tags = "${merge(
    map("Name", var.stack_name,
        "Environment", var.stack_name,
        format("kubernetes.io/cluster/%v", var.stack_name), "owned"),
    var.tags)}"
}

provider "aws" {
  region     = "${var.aws_region}"
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  profile    = "default"
}

###########################################
# images
###########################################

data "aws_ami" "latest_ami" {
  filter {
    name   = "name"
    values = ["${var.ami_name_pattern}"]
  }

  most_recent = true
  owners      = ["${var.ami_owner}"]

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
