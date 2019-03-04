package ssh

import (
	"log"
)

var (
	stateMap = map[string]Runner{}
)

type Runner func(t *Target, data interface{}) (error)

func (t *Target) Apply(data interface{}, states ...string) error {
	for _, state := range states {
		log.Printf("target %v: about to apply state %v\n", t.Node(), state)
		if state, stateExists := stateMap[state]; stateExists {
			state(t, data)
		} else {
			log.Fatalf("state does not exist: %s", state)
		}
	}
	return nil
}
