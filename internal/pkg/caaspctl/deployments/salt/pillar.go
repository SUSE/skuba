package salt

type Pillar struct {
	Bootstrap *Bootstrap `json:"bootstrap,omitempty"`
	Join      *Join      `json:"join,omitempty"`
}

type Bootstrap struct {
	Kubeadm Kubeadm `json:"kubeadm"`
	Cni     Cni     `json:"cni"`
}

type Join struct {
	Kubeadm    Kubeadm     `json:"kubeadm"`
	Kubernetes *Kubernetes `json:"kubernetes,omitempty"`
}

type Kubeadm struct {
	ConfigPath string `json:"config_path"`
}

type Cni struct {
	ConfigPath string `json:"config_path"`
}

type Kubernetes struct {
	SecretsPath string `json:"secrets_path"`
}
