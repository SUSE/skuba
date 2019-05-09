variable "stack_name" {
  default     = "caasp-test"
  description = "identifier to make all your resources unique and avoid clashes with other users of this terraform project"
}

variable "region" {
  default     = "eu-west-3"
  description = "Name of the region to be used - London by default"
}

variable "ami_name_pattern" {
  default     = "openSUSE-Leap-15-*"
  description = "Pattern for choosing the AMI image"
}

variable "ami_owner" {
  default     = "679593333241"
  description = "AMI owner id - SUSE by default"
}

variable "authorized_keys" {
  type        = "list"
  default     = []
  description = "ssh keys to inject into all the nodes. First key will be used for creating a keypair."
}

variable "subnet_cidr" {
  type        = "string"
  default     = "10.0.0.0/16"
  description = "Subnet CIDR"
}

variable "access_key" {
  default     = ""
  description = "AWS access key"
}

variable "secret_key" {
  default     = ""
  description = "AWS secret key"
}

variable "master_size" {
  default     = "t2.micro"
  description = "Size of the master nodes"
}

variable "masters" {
  default     = 1
  description = "Number of master nodes"
}

variable "worker_size" {
  default     = "t2.micro"
  description = "Size of the worker nodes"
}

variable "workers" {
  default     = 1
  description = "Number of worker nodes"
}

variable "public_worker" {
  description = "Weither or not the workers should have a public IP"
  default     = true
}

variable "repositories" {
  type = "list"

  default = [{
    kubic = "https://download.opensuse.org/repositories/devel:/kubic/openSUSE_Leap_15.1"
  }]

  description = "List of extra repositories (as maps with '<name>'='<url>') to add via cloud-init"
}

variable "packages" {
  type = "list"

  default = [
    "kubernetes-kubeadm",
    "kubernetes-kubelet",
    "kubernetes-client",
    "cri-o",
    "cni-plugins",
    "kmod",
    "-docker",
    "-containerd",
    "-docker-runc",
    "-docker-libnetwork",
  ]

  description = "list of additional packages to install"
}

variable "username" {
  default     = "opensuse"
  description = "Username for the cluster nodes"
}

variable "password" {
  default     = "linux"
  description = "Password for the cluster nodes"
}

variable "tags" {
  type        = "map"
  default     = {}
  description = "Extra tags used for the AWS resources created"
}
