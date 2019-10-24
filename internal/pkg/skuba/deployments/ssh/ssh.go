/*
 * Copyright (c) 2019 SUSE LLC.
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
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/deployments"
)

// defKnowHosts is the default `known_hosts` file
const defKnowHosts = "known_hosts"

// trustHostMessage is the message printed when we don't know about a host
// fingerprint.
// (this message is intentionally similar to the message printed by `ssh`)
var trustHostMessage = template.Must(template.New("id-chg").Parse(`
The authenticity of host '{{.Address}}' can't be established.
{{.Algorithm}} key fingerprint is {{.Fingerprint}}.`))

// fingerprintMismatchMessage is the (huge) message printed when the SSH
// fingerprint does not match the value stored in `known_hosts`.
// (this message is intentionally similar to the message printed by `ssh`)
var fingerprintMismatchMessage = template.Must(template.New("id-chg").Parse(`
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now (man-in-the-middle attack)!
It is also possible that the {{.Algorithm }} host key has just been changed.
The fingerprint for the RSA key sent by the remote host is
{{.Fingerprint}}.
Please contact your system administrator.
Add correct host key in {{.Filename}} to get rid of this message, or remove the offending key with: 
$ ssh-keygen -R {{.Address}} -f {{.Filename}}
Host key verification failed.`))

var (
	errSSHAuthErr   = errors.New("authentication error")
	errSSHNoKeysErr = errors.New("no keys loaded in the ssh-agent")
)

var (
	algoToStr = map[string]string{
		ssh.KeyAlgoRSA:      "RSA",
		ssh.KeyAlgoDSA:      "DSA",
		ssh.KeyAlgoECDSA256: "ECDSA",
		ssh.KeyAlgoECDSA384: "ECDSA",
		ssh.KeyAlgoECDSA521: "ECDSA",
	}
)

type Target struct {
	target     *deployments.Target
	user       string
	targetName string
	sudo       bool
	port       int
	client     *ssh.Client
}

// GetFlags adds init flags bound to the config to the specified flagset
func (t *Target) GetFlags() *flag.FlagSet {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagSet.StringVarP(&t.user, "user", "u", "root", "User identity used to connect to target")
	flagSet.BoolVarP(&t.sudo, "sudo", "s", false, "Run remote command via sudo")
	flagSet.IntVarP(&t.port, "port", "p", 22, "Port to connect to using SSH")
	flagSet.StringVarP(&t.targetName, "target", "t", "", "IP or FQDN of the node to connect to using SSH (required)")

	_ = cobra.MarkFlagRequired(flagSet, "target")

	return flagSet
}

func (t Target) String() string {
	return fmt.Sprintf("%s@%s:%d", t.user, t.target.Target, t.port)
}

func (t *Target) GetDeployment(nodename string) *deployments.Target {
	res := deployments.Target{
		Target:   t.targetName,
		Nodename: nodename,
	}
	res.Actionable = &Target{
		target: &res,
		user:   t.user,
		sudo:   t.sudo,
		port:   t.port,
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
		klog.V(1).Infof("running command: %q", finalCommand)
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
			klog.V(1).Infof("%s | %s", description, scanner.Text())
		}
	}
	outputChan <- result.String()
}

// initClient initializes the ssh client to the target
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

	hostKeyCallback, err := t.hostKeyChecker()
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: t.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: hostKeyCallback,
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

// hostKeyChecker checks that the host fingerprint is stored in the known_hosts
// if not present, it warns user (optionally asking for the key to be accepted or not)
// adding the key to the `known_hosts` file. In case the key is found but there is a
// mismatch (or the key has been rejected), it returns an error.
func (t Target) hostKeyChecker() (ssh.HostKeyCallback, error) {
	// make sure the filename exists from the start
	err := os.MkdirAll(path.Dir(defKnowHosts), 0700)
	if err != nil {
		return nil, err
	}

	out, err := os.OpenFile(defKnowHosts, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	hostKeyCallback, err := knownhosts.New(defKnowHosts)
	if err != nil {
		klog.Errorf("could not create callback function for checking hosts fingerprints: %s", err)
		return nil, errSSHNoKeysErr
	}

	return ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := hostKeyCallback(hostname, remote, key)
		if err == nil {
			return nil
		}

		if re, ok := err.(*knownhosts.RevokedError); ok {
			klog.Errorf("remote host identification for %q has been revoked", hostname)
			return re
		} else if ke, ok := err.(*knownhosts.KeyError); ok {
			algoStr := algoToStr[key.Type()]
			keyFingerprintStr := md5String(md5.Sum(key.Marshal()))

			// process one of the error messages as a template, returning the
			// text after replacing some vars...
			replaceMessage := func(tmpl *template.Template) string {
				buf := bytes.Buffer{}
				if err := tmpl.Execute(&buf, struct {
					Algorithm   string
					Address     string
					Fingerprint string
					Filename    string
				}{
					algoStr,
					remote.String(),
					keyFingerprintStr,
					defKnowHosts,
				}); err != nil {
					klog.Fatal("could not perform replacements in template")
				}
				return buf.String()
			}

			// Want holds the accepted host keys. For each key algorithm,
			// there can be one hostkey.  If Want is empty, the host is
			// unknown. If Want is non-empty, there was a mismatch, which
			// can signify a MITM attack.
			if len(ke.Want) == 0 {
				klog.Warning(replaceMessage(trustHostMessage))
				klog.Infof("accepting SSH key for %q", hostname)
				klog.Infof("adding fingerprint for %q to %q", hostname, defKnowHosts)
				line := knownhosts.Line([]string{remote.String()}, key)
				if _, err := out.WriteString(line + "\n"); err != nil {
					return err
				}
				return nil
			}
			// fingerprint mismatch: print a big warning and return an error
			klog.Error(replaceMessage(fingerprintMismatchMessage))
		}
		return err
	}), nil
}

// md5String returns a formatted string representing the given md5Sum in hex
func md5String(md5Sum [16]byte) string {
	md5Str := fmt.Sprintf("% x", md5Sum)
	md5Str = strings.Replace(md5Str, " ", ":", -1)
	return md5Str
}
