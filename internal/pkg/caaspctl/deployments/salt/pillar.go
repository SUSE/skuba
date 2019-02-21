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
	Kubeadm Kubeadm `json:"kubeadm"`
}

type Kubeadm struct {
	ConfigPath string `json:"config_path"`
}

type Cni struct {
	ConfigPath string `json:"config_path"`
}
