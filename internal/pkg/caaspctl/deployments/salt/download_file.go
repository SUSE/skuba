package salt

import (
	"fmt"
	"strings"
)

func DownloadFile(target Target, masterConfig MasterConfig, file string) (string, error) {
	stdout, stderr, err := Ssh(target, masterConfig, "--no-color", "cmd.run", fmt.Sprintf("'cat %s'", file))
	if err != nil {
		return "",
			fmt.Errorf(
				"error while downloading file %s from %s: %s\n%v",
				file,
				target.Node,
				stderr,
				err)
	}
	return strings.Join(strings.Split(stdout, "\n    ")[1:], "\n"), nil
}
