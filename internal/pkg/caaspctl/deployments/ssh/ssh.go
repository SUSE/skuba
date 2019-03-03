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
	"suse.com/caaspctl/internal/pkg/caaspctl/deployments/ssh/assets"
)

var (
	States = map[string]deployments.Runner{
		"kubelet.configure": kubeletConfigure(),
		"kubelet.enable": kubeletEnable(),
		"kubeadm.init": kubeadmInit(),
		"cni.deploy": cniDeploy(),
	}
)

type Target struct {
	Node   string
	User   string
	Sudo   bool
	Client *ssh.Client
}

func NewTarget(target, user string, sudo bool) deployments.Target {
	return &Target{
		Node: target,
		User: user,
		Sudo: sudo,
	}
}

func (t *Target) Apply(states ...string) error {
	for _, state := range states {
		if state, stateExists := States[state]; stateExists {
			state.Run(t)
		} else {
			log.Fatalf("State does not exist: %s", state)
		}
	}
	return nil
}

func (t *Target) Target() string {
	return t.Node
}

func (t *Target) UploadFileContents(contents, target string) error {
	return nil
}

func (t *Target) DownloadFileContents(source string) (string, error) {
	return "", nil
}

func (t *Target) ssh(command string, args []string) (stdout string, stderr string, error error) {
	return t.sshWithStdin(command, args, "")
}

func (t *Target) sshWithStdin(command string, args []string, stdin string) (stdout string, stderr string, error error) {
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
		fmt.Printf("%s: %s\n", description, scanner.Text())
	}
	outputChan <- result.String()
}

func (t *Target) initClient() {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("net.Dial: %v", err)
	}
	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: t.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	t.Client, err = ssh.Dial("tcp", t.Node, config)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
}

func sshTarget(target deployments.Target) *Target {
	if target, ok := target.(*Target); ok {
		return target
	}
	log.Fatal("Target is of the wrong type")
	return nil
}

func kubeletConfigure() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		if target := sshTarget(t); target != nil {
			target.sshWithStdin("cat", []string{"> /lib/systemd/system/kubelet.service"}, assets.KubeletService)
			target.sshWithStdin("cat", []string{"> /etc/systemd/system/kubelet.service.d/10-kubeadm.conf"}, assets.KubeadmService)
			target.sshWithStdin("cat", []string{"> /etc/sysconfig/kubelet"}, assets.KubeletSysconfig)
			target.ssh("systemctl", []string{"daemon-reload"})
		}
		return nil
	}
	return runner
}

func kubeletEnable() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		if target := sshTarget(t); target != nil {
			target.ssh("systemctl", []string{"enable", "kubelet"})
		}
		return nil
	}
	return runner
}

func kubeadmInit() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		if target := sshTarget(t); target != nil {
			target.sshWithStdin("cat > /tmp/kubeadm.conf", []string{}, "")
			target.ssh("systemctl", []string{"enable", "--now", "docker"})
			target.ssh("systemctl", []string{"stop", "kubelet"})
			target.ssh("kubeadm", []string{"init", "--config", "/tmp/kubeadm.conf", "--skip-token-print"})
			target.ssh("rm", []string{"/tmp/kubeadm.conf"})
		}
		return nil
	}
	return runner
}

func cniDeploy() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		// Deploy locally
		return nil
	}
	return runner
}
