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

  tags = merge(
    local.basic_tags,
    {
      format("kubernetes.io/cluster/%v", var.stack_name) = "SUSE-terraform"
    },
  )
}

provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key
  profile    = "default"
}

resource "aws_key_pair" "kube" {
  key_name   = "${var.stack_name}-keypair"
  public_key = element(var.authorized_keys, 0)
}

