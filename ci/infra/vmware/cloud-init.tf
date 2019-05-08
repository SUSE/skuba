data "local_file" "cloud_init_metadata" {
    filename = "cloud-init/meta-data.ccfile"
}

data "local_file" "cloud_init_netconfig" {
    filename = "cloud-init/network-config.ccfile"
}

locals {
  cloud_init_metadata  = "${data.local_file.cloud_init_metadata.content}"
  cloud_init_netconfig = "${data.local_file.cloud_init_netconfig.content}"
}

resource "null_resource" "local_clean_up_isos" {
  provisioner "local-exec" {
    when    = "destroy"
    command = "rm -f $PWD/cc-*.iso"
  }
}

