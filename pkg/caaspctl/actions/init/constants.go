package init

import (
	"suse.com/caaspctl/pkg/caaspctl"
)

var (
	scaffoldFiles = []struct {
		Location string
		Content  string
	}{
		{
			Location: caaspctl.KubeadmInitConfFile(),
			Content:  kubeadmInitConf,
		},
		{
			Location: caaspctl.MasterConfTemplateFile(),
			Content:  masterConfTemplate,
		},
		{
			Location: caaspctl.WorkerConfTemplateFile(),
			Content:  workerConfTemplate,
		},
		{
			Location: caaspctl.FlannelManifestFile(),
			Content:  flannelManifest,
		},
	}
)
