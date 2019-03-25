variable "stack_name" {
  default = "caasp"
}

variable "repositories" {
  type = "list"

  default = [
    {
      caasp_devel_leap15 = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
    },
  ]

  description = "Urls of the repositories to mount via cloud-init"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "repo_baseurl" {
  type    = "string"
  default = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
}

variable "worker_count" {
  default     = 1
  description = "number of worker nodes"
}

variable "master_count" {
  default     = 1
  description = "number of master nodes"
}
