data "template_file" "register_rmt" {
  template = file("${path.module}/cloud-init/register-rmt.tpl")
  count    = var.rmt_server_name == "" ? 0 : 1

  vars = {
    rmt_server_name = var.rmt_server_name
  }
}

data "template_file" "register_scc" {
  # register with SCC if an RMT has not been provided
  count    = var.caasp_registry_code != "" && var.rmt_server_name == "" ? 1 : 0
  template = file("${path.module}/cloud-init/register-scc.tpl")

  vars = {
    caasp_registry_code = var.caasp_registry_code
  }
}

data "template_file" "register_suma" {
  template = file("${path.module}/cloud-init/register-suma.tpl")
  count    = var.suma_server_name == "" ? 0 : 1

  vars = {
    suma_server_name = var.suma_server_name
  }
}

data "template_file" "repositories" {
  count    = length(var.repositories)
  template = file("${path.module}/cloud-init/repository.tpl")

  vars = {
    repository_url  = element(values(var.repositories), count.index)
    repository_name = element(keys(var.repositories), count.index)
  }
}

data "template_file" "ntp_servers" {
  count    = length(var.ntp_servers) == 0 ? 0 : 1
  template = file("${path.module}/cloud-init/ntp.tpl")

  vars = {
    ntp_servers = join(" ", var.ntp_servers)
  }
}

data "template_file" "dns_nameservers" {
  count    = length(var.dns_nameservers) == 0 ? 0 : 1
  template = file("${path.module}/cloud-init/nameserver.tpl")

  vars = {
    name_servers = join(" ", var.dns_nameservers)
  }
}

data "template_file" "commands" {
  count    = length(var.packages) == 0 ? 0 : 1
  template = file("${path.module}/cloud-init/commands.tpl")

  vars = {
    packages = join(", ", var.packages)
  }
}

data "template_file" "cloud-init" {
  template = file("${path.module}/cloud-init/init.sh.tpl")

  vars = {
    commands      = join("\n", data.template_file.commands.*.rendered)
    ntp_servers   = join("\n", data.template_file.ntp_servers.*.rendered)
    name_servers  = join("\n", data.template_file.dns_nameservers.*.rendered)
    repositories  = length(var.repositories) == 0 ? "\n" : join("\n", data.template_file.repositories.*.rendered)
    register_scc  = var.caasp_registry_code != "" && var.rmt_server_name == "" ? join("\n", data.template_file.register_scc.*.rendered) : ""
    register_rmt  = var.rmt_server_name != "" ? join("\n", data.template_file.register_rmt.*.rendered) : ""
    register_suma = var.suma_server_name != "" ? join("\n", data.template_file.register_suma.*.rendered) : ""
  }
}
