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
	"net"
	"os"
	"strings"

	"k8s.io/klog"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/SUSE/caaspctl/internal/pkg/caaspctl/deployments"
)

var (
	errSSHAuthErr   = errors.New("authentication error")
	errSSHNoKeysErr = errors.New("no keys loaded in the ssh-agent")
)

type Target struct {
	target *deployments.Target
	user   string
	sudo   bool
	port   int
	client *ssh.Client
}

func NewTarget(nodename, target, user string, sudo bool, port int) *deployments.Target {
	res := deployments.Target{
		Target:   target,
		Nodename: nodename,
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
	return t.internalSshWithStdin(true, "", command, args...)
}

func (t *Target) ssh(command string, args ...string) (stdout string, stderr string, error error) {
	return t.internalSshWithStdin(false, "", command, args...)
}

func (t *Target) silentSshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
	return t.internalSshWithStdin(true, stdin, command, args...)
}

func (t *Target) sshWithStdin(stdin string, command string, args ...string) (stdout string, stderr string, error error) {
	return t.internalSshWithStdin(false, stdin, command, args...)
}

func (t *Target) internalSshWithStdin(silent bool, stdin string, command string, args ...string) (stdout string, stderr string, error error) {
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
	if !silent {
		klog.Infof("running command: %q", finalCommand)
	}
	if err := session.Start(finalCommand); err != nil {
		return "", "", err
	}
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	go readerStreamer(stdoutReader, stdoutChan, "stdout", silent)
	go readerStreamer(stderrReader, stderrChan, "stderr", silent)
	if err := session.Wait(); err != nil {
		return "", "", err
	}
	stdout = <-stdoutChan
	stderr = <-stderrChan
	return
}

func readerStreamer(reader io.Reader, outputChan chan<- string, description string, silent bool) {
	result := bytes.Buffer{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		result.Write([]byte(scanner.Text()))
		if !silent {
			klog.Infof("%s | %s\n", description, scanner.Text())
		}
	}
	outputChan <- result.String()
}

func (t *Target) initClient() error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if len(socket) == 0 {
		return errors.Errorf("SSH_AUTH_SOCK is undefined. Make sure ssh-agent is running")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}
	agentClient := agent.NewClient(conn)

	// check a precondition: there must be some SSH keys loaded in the ssh agent
	keys, err := agentClient.List()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		klog.Errorf("no keys have been loaded in the ssh-agent.")
		return errSSHNoKeysErr
	}

	config := &ssh.ClientConfig{
		User: t.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	t.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", t.target.Target, t.port), config)
	if err != nil {
		// crypto/ssh does not provide constants for some common errors, so we
		// must "pattern match" the error strings in order to guess what failed
		if strings.Contains(err.Error(), "unable to authenticate") {
			klog.Errorf("ssh authentication error: please make sure you have added to "+
				"your ssh-agent a ssh key that is authorized in %q.", t.target.Target)
			return errSSHAuthErr
		}
		return err
	}
	return nil
}
