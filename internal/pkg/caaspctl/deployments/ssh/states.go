package ssh

import (
	"log"
)

var (
	stateMap = map[string]Runner{}
)

type Runner func(t *Target, data interface{}) (error)

func (t *Target) Apply(data interface{}, states ...string) error {
	for _, stateName := range states {
		log.Printf("=== %s: about to apply state %s ===\n", t.Node(), stateName)
		if state, stateExists := stateMap[stateName]; stateExists {
			if err := state(t, data); err != nil {
				log.Printf("=== %s: failed to apply state %s: %v ===\n", t.Node(), stateName, err)
			} else {
				log.Printf("=== %s: state %s applied successfully ===\n", t.Node(), stateName)
			}
		} else {
			log.Fatalf("state does not exist: %s", stateName)
		}
	}
	return nil
}
