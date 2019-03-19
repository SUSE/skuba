data "aws_ami" "latest_ami" {
  filter {
    name   = "name"
    values = ["openSUSE-Leap-15-*"]
  }

  most_recent = true
  owners      = ["679593333241"] # openSUSE
}
