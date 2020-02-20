data "template_file" "register_rmt" {
  template = file("cloud-init/register-rmt.tpl")
  count    = var.rmt_server_name == "" ? 0 : 1

  vars = {
    rmt_server_name = var.rmt_server_name
  }
}

data "template_file" "register_scc" {
  # register with SCC iff an RMT has not been provided
  count    = var.caasp_registry_code != "" && var.rmt_server_name == "" ? 1 : 0
  template = file("cloud-init/register-scc.tpl")

  vars = {
    caasp_registry_code = var.caasp_registry_code
  }
}

data "template_file" "register_suma" {
  template = file("cloud-init/register-suma.tpl")
  count    = var.suma_server_name == "" ? 0 : 1

  vars = {
    suma_server_name = var.suma_server_name
  }
}

data "template_file" "repositories" {
  count    = length(var.repositories)
  template = file("cloud-init/repository.tpl")

  vars = {
    repository_url  = var.repositories[count.index]
    repository_name = var.repositories[count.index]
  }
}

data "template_file" "commands" {
  template = file("cloud-init/commands.tpl")
  count    = length(var.packages) == 0 ? 0 : 1

  vars = {
    packages = join(", ", var.packages)
  }
}

data "template_file" "cloud-init" {
  template = file("cloud-init/cloud-init.yaml.tpl")

  vars = {
    authorized_keys = join("\n", formatlist("  - %s", var.authorized_keys))
    commands        = join("\n", data.template_file.commands.*.rendered)
    repositories    = length(var.repositories) == 0 ? "\n" : join("\n", data.template_file.repositories.*.rendered)
    register_scc    = var.caasp_registry_code != "" && var.rmt_server_name == "" ? join("\n", data.template_file.register_scc.*.rendered) : ""
    register_rmt    = var.rmt_server_name != "" ? join("\n", data.template_file.register_rmt.*.rendered) : ""
    register_suma   = var.suma_server_name != "" ? join("\n", data.template_file.register_suma.*.rendered) : ""
  }
}

data "template_cloudinit_config" "cfg" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/cloud-config"
    content      = data.template_file.cloud-init.rendered
  }
}

