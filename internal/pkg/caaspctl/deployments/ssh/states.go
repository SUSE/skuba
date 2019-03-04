package ssh

import (
	"log"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

var (
	stateMap = map[string]deployments.Runner{}
)

func (t *Target) Apply(data interface{}, states ...string) error {
	for _, state := range states {
		if state, stateExists := stateMap[state]; stateExists {
			state.Run(t, data)
		} else {
			log.Fatalf("State does not exist: %s", state)
		}
	}
	return nil
}
