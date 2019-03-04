package ssh

import (
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

func init() {
	stateMap["kubernetes.upload-secrets"] = kubernetesUploadSecrets()
}

func kubernetesUploadSecrets() Runner {
	runner := struct{ State }{}
	runner.DoRun = func(t *Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			t.UploadFile(file, path.Join("/etc/kubernetes", file))
		}
		return nil
	}
	return runner
}
