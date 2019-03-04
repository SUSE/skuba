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
		log.Printf("=== about to apply state %s ===\n", stateName)
		if state, stateExists := stateMap[stateName]; stateExists {
			if err := state(t, data); err != nil {
				log.Printf("=== failed to apply state %s: %v ===\n", stateName, err)
			} else {
				log.Printf("=== state %s applied successfully ===\n", stateName)
			}
		} else {
			log.Fatalf("state does not exist: %s", stateName)
		}
	}
	return nil
}
