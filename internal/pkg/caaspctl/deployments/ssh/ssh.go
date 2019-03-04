package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

type Target struct {
	Target *deployments.Target
	User   string
	Sudo   bool
	Port   int
	Client *ssh.Client
}

func NewTarget(target, user string, sudo bool, port int) deployments.Target {
	res := deployments.Target{
		Node: target,
	}
	res.Actionable = &Target{
		Target: &res,
		User: user,
		Sudo: sudo,
		Port: port,
	}
	return res
}

func (t *Target) ssh(command string, args ...string) (stdout string, stderr string, error error) {
	return t.sshWithStdin("", command, args...)
}

func (t *Target) sshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
	if t.Client == nil {
		t.initClient()
	}
	session, err := t.Client.NewSession()
	if err != nil {
		return "", "", err
	}
	if len(stdin) > 0 {
		session.Stdin = bytes.NewBufferString(stdin)
	}
	stdoutReader, err := session.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderrReader, err := session.StderrPipe()
	if err != nil {
		return "", "", err
	}
	finalCommand := strings.Join(append([]string{command}, args...), " ")
	if t.Sudo {
		finalCommand = fmt.Sprintf("sudo sh -c '%s'", finalCommand)
	}
	log.Printf("running command: %s", finalCommand)
	if err := session.Start(finalCommand); err != nil {
		return "", "", err
	}
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	go readerStreamer(stdoutReader, stdoutChan, "stdout")
	go readerStreamer(stderrReader, stderrChan, "stderr")
	if err := session.Wait(); err != nil {
		return "", "", err
	}
	stdout = <-stdoutChan
	stderr = <-stderrChan
	return
}

func readerStreamer(reader io.Reader, outputChan chan<- string, description string) {
	result := bytes.Buffer{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		result.Write([]byte(scanner.Text()))
		log.Printf("%s: %s\n", description, scanner.Text())
	}
	outputChan <- result.String()
}

func (t *Target) initClient() {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: t.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	t.Client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", t.Target.Target(), t.Port), config)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
}
