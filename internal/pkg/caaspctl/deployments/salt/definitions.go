package salt

import (
	"fmt"

	"suse.com/caaspctl/internal/pkg/caaspctl/definitions"
)

func CurrentDefinitionPrefix() string {
	return fmt.Sprintf("salt://samples/%s", definitions.CurrentDefinition())
}
