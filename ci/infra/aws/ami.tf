data "susepubliccloud_image_ids" "sles15sp1_byos" {
  cloud      = "amazon"
  region     = "${var.aws_region}"
  state      = "active"
  name_regex = "suse-sles-15-sp1-byos.*-hvm-ssd-x86_64"
}
