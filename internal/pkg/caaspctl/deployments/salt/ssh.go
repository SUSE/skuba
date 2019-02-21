package salt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"suse.com/caaspctl/internal/pkg/caaspctl/constants"
)

func Ssh(target string, command string, args ...string) (string, string, error) {
	saltArgs := []string{"-c", ".", "-i", "--roster=scan", "--key-deploy", "--user=vagrant", "--sudo", target, command}
	saltArgs = append(saltArgs, args...)
	cmd := exec.Command("salt-ssh", saltArgs...)
	cmd.Dir = constants.DefinitionPath
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
