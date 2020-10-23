# https://docs.microsoft.com/en-us/azure/virtual-machines/linux/using-cloud-init

data "template_file" "repositories" {
  count    = length(var.repositories)
  template = file("${path.module}/cloud-init/repository.tpl")

  vars = {
    repository_url  = element(
      values(var.repositories), 
      count.index
    )
    repository_name = element(
      keys(var.repositories), 
      count.index
    )
  }
}

data "template_file" "register_scc" {
  count    = var.caasp_registry_code != "" && var.rmt_server_name == "" ? 1 : 0
  template = file("${path.module}/cloud-init/register-scc.tpl")

  vars = {
    caasp_registry_code = var.caasp_registry_code
    rmt_server_name     = var.rmt_server_name
  }
}

data "template_file" "register_rmt" {
  count    = var.rmt_server_name == "" ? 0 : 1
  template = file("${path.module}/cloud-init/register-rmt.tpl")

  vars = {
    rmt_server_name = var.rmt_server_name
  }
}

data "template_file" "register_suma" {
  count    = var.suma_server_name == "" ? 0 : 1
  template = file("${path.module}/cloud-init/register-suma.tpl")

  vars = {
    suma_server_name = var.suma_server_name
  }
}

data "template_file" "commands" {
  count    = join("", var.packages) == "" ? 0 : 1
  template = file("${path.module}/cloud-init/commands.tpl")

  vars = {
    packages = join(" ", var.packages)
  }
}

data "template_file" "cloud-init" {
  template = file("${path.module}/cloud-init/cloud-init.yaml.tpl")

  vars = {
    authorized_keys = join("\n", formatlist("  - %s", var.authorized_keys))
    register_scc    = var.caasp_registry_code != "" && var.rmt_server_name == "" ? join("\n", data.template_file.register_scc.*.rendered) : ""
    register_rmt    = var.rmt_server_name != "" ? join("\n", data.template_file.register_rmt.*.rendered) : ""
    register_suma   = var.suma_server_name != "" ? join("\n", data.template_file.register_suma.*.rendered) : ""
    repositories    = length(var.repositories) == 0 ? "\n" : join("\n", data.template_file.repositories.*.rendered)
    commands        = join("\n", data.template_file.commands.*.rendered)
  }
}

data "template_cloudinit_config" "cfg" {
  gzip          = false
  base64_encode = true

  part {
    content_type = "text/cloud-config"
    content      = data.template_file.cloud-init.rendered
  }
}
