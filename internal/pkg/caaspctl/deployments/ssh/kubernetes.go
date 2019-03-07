package ssh

import (
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

const (
	KubernetesUploadSecretsFailOnError     = iota
	KubernetesUploadSecretsContinueOnError = iota
)

func init() {
	stateMap["kubernetes.bootstrap.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsContinueOnError)
	stateMap["kubernetes.join.upload-secrets"] = kubernetesUploadSecrets(KubernetesUploadSecretsFailOnError)
}

type KubernetesUploadSecretsErrorBehavior uint

func kubernetesUploadSecrets(errorHandling KubernetesUploadSecretsErrorBehavior) Runner {
	return func(t *Target, data interface{}) error {
		for _, file := range deployments.Secrets {
			if err := t.target.UploadFile(file, path.Join("/etc/kubernetes", file)); err != nil {
				if errorHandling == KubernetesUploadSecretsFailOnError {
					return err
				}
			}
		}
		return nil
	}
}
