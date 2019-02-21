package definitions

import (
	"path"
)

func CurrentDefinition() string {
	return "3-masters-3-workers-vagrant"
}

func CurrentDefinitionPrefix() string {
	return path.Join("samples", CurrentDefinition())
}
