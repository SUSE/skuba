#####################
# Cluster variables #
#####################

variable "img" {
  type        = "string"
  default     = "opensuse-caasp"
  description = "image name"
}

variable "force_img" {
  type        = "string"
  default     = ""
  description = "force the image re-creation"
}

variable "master_count" {
  default     = 1
  description = "Number of masters to be created"
}

variable "worker_count" {
  default     = 2
  description = "Number of workers to be created"
}

variable "name_prefix" {
  type        = "string"
  default     = "ag-"
  description = "Optional prefix to be able to have multiple clusters on one host"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "ssh_user" {
  type        = "string"
  default     = "root"
  description = "The SSH user"
}

variable "ssh_pass" {
  type        = "string"
  default     = "linux"
  description = "The SSH password"
}

variable "domain_name" {
  type        = "string"
  default     = "test.net"
  description = "The domain name"
}
