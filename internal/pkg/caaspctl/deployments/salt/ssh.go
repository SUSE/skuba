package salt

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Target struct {
	Node string
	User string
	Sudo bool
}

func Ssh(target Target, masterConfig MasterConfig, command string, args ...string) (string, string, error) {
	saltArgs := []string{
		"-c",
		masterConfig.GetTempDir(),
		"-i",
		"--roster=scan",
		"--key-deploy",
		fmt.Sprintf("--user=%s", target.User),
	}

	defer os.RemoveAll(masterConfig.GetTempDir())

	if target.Sudo {
		saltArgs = append(saltArgs, "--sudo")
	}

	saltArgs = append(
		saltArgs,
		target.Node,
		command,
	)

	saltArgs = append(saltArgs, args...)

	cmd := exec.Command("salt-ssh", saltArgs...)
	fmt.Printf("Command is %v\n", cmd)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	stdout := stdOut.String()
	stderr := stdErr.String()
	fmt.Printf("stdout is %s\n", stdout)
	fmt.Printf("stderr is %s\n", stderr)
	if err != nil {
		log.Fatal(err)
	}

	return stdout, stderr, nil
}
