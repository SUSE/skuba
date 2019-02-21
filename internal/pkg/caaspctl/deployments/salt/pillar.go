package salt

type Pillar struct {
	Kubeadm *Kubeadm `json:"kubeadm,omitempty"`
	Cni     *Cni     `json:"cni,omitempty"`
}

type Kubeadm struct {
	ConfigPath string `json:"config_path"`
}

type Cni struct {
	ConfigPath string `json:"config_path"`
}
