package salt

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

func Ssh(target string, command string, args ...string) error {
	saltArgs := []string{"-c", ".", "-i", "--roster=scan", "--key-deploy", "--user=vagrant", "--sudo", target, command}
	saltArgs = append(saltArgs, args...)
	cmd := exec.Command("salt-ssh", saltArgs...)
	cmd.Dir = "deployments/salt"
	fmt.Printf("Command is %v\n", cmd)
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	err := cmd.Run()
	fmt.Printf("stdout is %s\n", stdOut.String())
	fmt.Printf("stderr is %s\n", stdErr.String())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
