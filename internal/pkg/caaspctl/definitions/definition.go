package definitions

import (
	"path"

	"suse.com/caaspctl/internal/pkg/caaspctl/constants"
)

func PKIPath() string {
	return path.Join(
		constants.DefinitionPath,
		"states",
		"samples",
		CurrentDefinition(),
		"pki",
	)
}

func CurrentDefinition() string {
	return "3-masters-3-workers-vagrant"
}

func CurrentDefinitionPrefix() string {
	return path.Join("samples", CurrentDefinition())
}
