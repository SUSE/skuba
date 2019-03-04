package ssh

import (
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func init() {
	stateMap["kubernetes.upload-secrets"] = kubernetesUploadSecrets()
}

func kubernetesUploadSecrets() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			t.UploadFile(file, path.Join("/etc/kubernetes", file))
		}
		return nil
	}
	return runner
}
