package salt

import (
	"fmt"
	"log"
	"strings"
)

func DownloadFile(target string, file string) (string, error) {
	stdout, _, err := Ssh(target, "--no-color", "cmd.run", fmt.Sprintf("'cat %s'", file))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Join(strings.Split(stdout, "\n    ")[1:], "\n"), nil
}
