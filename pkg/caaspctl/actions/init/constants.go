package init

var (
	ScaffoldFiles = []struct{
		Location string
		Content  string
	}{
		{
			Location: "kubeadm-init.conf",
			Content: kubeadmInitConf,
		},
		{
			Location: "kubeadm-join-conf.d/master.conf.template",
			Content: masterConfTemplate,
		},
		{
			Location: "kubeadm-join-conf.d/worker.conf.template",
			Content: workerConfTemplate,
		},
		{
			Location: "addons/cni/flannel.yaml",
			Content: flannelManifests,
		},
	}
)
