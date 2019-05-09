variable "region" {
  default     = "eu-west-1"
  description = "Name of the region to be used - Ireland by default"
}

variable "availability_zones" {
  type        = "list"
  default     = ["eu-west-1a", "eu-west-1b", "eu-west-1c"]
  description = "ZoneName(s) of the availability zones to be used"
}

variable "access_key" {
  default     = ""
  description = "Access key to use"
}

variable "secret_key" {
  default     = ""
  description = "Secret key to use"
}

variable "master_size" {
  default     = "t2.medium"
  description = "Size of the master nodes"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "worker_size" {
  default     = "t2.medium"
  description = "Size of the worker nodes"
}

variable "workers" {
  default     = 1
  description = "Number of worker nodes"
}

variable "stack_name" {
  default     = "testing"
  description = "identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes"
}

variable "repo_baseurl" {
  default     = "https://download.opensuse.org/repositories/devel:/CaaSP:/Head:/ControllerNode/openSUSE_Leap_15.0"
  description = "Url of the repository to mount via cloud-init"
}
