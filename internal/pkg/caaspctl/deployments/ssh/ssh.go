package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

func (t *Target) ssh(command string, args []string, stdin string) (stdout string, stderr string, error error) {
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
	if err := session.Run(finalCommand); err != nil {
		return "", "", err
	}
	stdoutBytes, error := ioutil.ReadAll(stdoutReader)
	stdout = string(stdoutBytes)
	stderrBytes, error := ioutil.ReadAll(stderrReader)
	stderr = string(stderrBytes)
	return
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
			target.ssh("cat > /lib/systemd/system/kubelet.service", []string{}, assets.KubeletService)
			target.ssh("cat > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf", []string{}, assets.KubeadmService)
			target.ssh("cat > /etc/sysconfig/kubelet", []string{}, assets.KubeletSysconfig)
			target.ssh("systemctl", []string{"daemon-reload"}, "")
		}
		return nil
	}
	return runner
}

func kubeletEnable() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		return nil
	}
	return runner
}

func kubeadmInit() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		return nil
	}
	return runner
}

func cniDeploy() deployments.Runner {
	runner := struct{ deployments.State }{}
	runner.DoRun = func(t deployments.Target) error {
		return nil
	}
	return runner
}
