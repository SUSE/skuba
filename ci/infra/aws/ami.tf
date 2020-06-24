data "susepubliccloud_image_ids" "sles15sp1_chost_byos" {
  cloud  = "amazon"
  region = data.aws_region.current.name
  state  = "active"

  # USE SLES 15 SP1 Container host AMI - this is needed to avoid issues like bsc#1146774
  name_regex = "suse-sles-15-sp1-chost-byos.*-hvm-ssd-x86_64"
}

data "aws_region" "current" {}
