locals {
  # Do not add the special `kubernetes.io/cluster<cluster-name>` here,
  # this tag cannot be added to all our resources otherwise the CPI
  # will get confused when dealing with security rules objects.
  basic_tags = merge(
    {
      "Name"        = var.stack_name
      "Environment" = var.stack_name
    },
    var.tags,
  )

  # tags = local.basic_tags
  tags = merge(
    local.basic_tags,
    {
      format("kubernetes.io/cluster/%v", var.stack_name) = "SUSE-terraform"
    },
  )
}

# https://www.terraform.io/docs/providers/aws/index.html
provider "aws" {
  profile = "default"
  region  = var.aws_region
}

data "susepubliccloud_image_ids" "sles15sp2_chost_byos" {
  cloud  = "amazon"
  region = var.aws_region
  state  = "active"

  # USE SLES 15 SP2 Container host AMI - this is needed to avoid issues like bsc#1146774
  name_regex = "suse-sles-15-sp2-chost-byos.*-hvm-ssd-x86_64"
}

resource "aws_key_pair" "kube" {
  key_name   = "${var.stack_name}-keypair"
  public_key = element(var.authorized_keys, 0)

  tags = merge(
    local.basic_tags,
    {
      "Name"  = "${var.stack_name}-keypair"
      "Class" = "KeyPair"
    },
  )
}

# list of availability_zones which can be access from the current region
data "aws_availability_zones" "availability_zones" {
  state = "available"
}
