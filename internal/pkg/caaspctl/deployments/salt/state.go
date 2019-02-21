package salt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func Apply(target Target, pillar *Pillar, mods ...string) error {
	args := []string{strings.Join(mods, ",")}
	if pillar != nil {
		jsonPillar, err := json.Marshal(*pillar)
		if err != nil {
			log.Fatal(err)
		}
		args = append(args, fmt.Sprintf("pillar=%s", jsonPillar))
	}
	_, _, err := Ssh(target, "state.sls", args...)
	return err
}
