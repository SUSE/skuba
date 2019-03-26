package lib

import (
	"golang.org/x/crypto/ssh"
)

// RunCmd execute command on remote systems
func RunCmd(host, command string) ([]byte, error) {
	client, session, err := connectToHost("root", host)
	if err != nil {
		return nil, err
	}
	out, err := session.CombinedOutput(command)
	if err != nil {
		return out, err
	}
	client.Close()
	return out, nil
}

// We assume that connection  ssh should be running on 22 port.
// By default we use linux, via ENV var this can changed.
func connectToHost(user, host string) (*ssh.Client, *ssh.Session, error) {
	// assume alwasy linux as root pwd.
	sshPassword := "linux"
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(sshPassword)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	client, err := ssh.Dial("tcp", host+":22", sshConfig)
	if err != nil {
		return nil, nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}
