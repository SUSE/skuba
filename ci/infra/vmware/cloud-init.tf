data "local_file" "cloud-init-metadata" {
    filename = "cloud-init/meta-data.ccfile"
}

data "local_file" "cloud-init-netconfig" {
    filename = "cloud-init/network-config.ccfile"
}

locals {
  cloud-init-metadata = "${data.local_file.cloud-init-metadata.content}"
  cloud-init-netconfig = "${data.local_file.cloud-init-netconfig.content}"
}
