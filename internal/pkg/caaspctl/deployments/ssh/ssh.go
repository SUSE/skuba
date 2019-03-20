/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"suse.com/caaspctl/internal/pkg/caaspctl/deployments"
)

type Target struct {
	target *deployments.Target
	user   string
	sudo   bool
	port   int
	client *ssh.Client
}

func NewTarget(nodename, target, user string, sudo bool, port int, kubeadmArgs map[string]interface{}) *deployments.Target {
	res := deployments.Target{
		Target:      target,
		Nodename:    nodename,
		KubeadmArgs: kubeadmArgs,
	}
	res.Actionable = &Target{
		target: &res,
		user:   user,
		sudo:   sudo,
		port:   port,
	}
	return &res
}

func (t *Target) silentSsh(command string, args ...string) (stdout string, stderr string, error error) {
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)
	return t.ssh(command, args...)
}

func (t *Target) ssh(command string, args ...string) (stdout string, stderr string, error error) {
	return t.sshWithStdin("", command, args...)
}

func (t *Target) silentSshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)
	return t.sshWithStdin(stdin, command, args...)
}

func (t *Target) sshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
	if t.client == nil {
		if err := t.initClient(); err != nil {
			return "", "", errors.Wrap(err, "failed to initialize client")
		}
	}
	session, err := t.client.NewSession()
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
	if t.sudo {
		finalCommand = fmt.Sprintf("sudo sh -c '%s'", finalCommand)
	}
	log.Printf("running command: %q", finalCommand)
	if err := session.Start(finalCommand); err != nil {
		return "", "", err
	}
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	go readerStreamer(stdoutReader, stdoutChan, "stdout | ")
	go readerStreamer(stderrReader, stderrChan, "stderr | ")
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
		log.Printf("%s%s\n", description, scanner.Text())
	}
	outputChan <- result.String()
}

func (t *Target) initClient() error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}
	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: t.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	t.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", t.target.Target, t.port), config)
	if err != nil {
		return err
	}
	return nil
}
