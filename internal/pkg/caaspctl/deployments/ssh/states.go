package ssh

import (
	"log"
)

var (
	stateMap = map[string]Runner{}
)

func (t *Target) Apply(data interface{}, states ...string) error {
	for _, state := range states {
		log.Printf("target %v: about to apply state %v\n", t.Target.Target(), state)
		if state, stateExists := stateMap[state]; stateExists {
			state.Run(t, data)
		} else {
			log.Fatalf("state does not exist: %s", state)
		}
	}
	return nil
}
