package ssh

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
)

func (t *Target) UploadFile(sourcePath, targetPath string) error {
	if contents, err := ioutil.ReadFile(sourcePath); err == nil {
		return t.UploadFileContents(targetPath, string(contents))
	}
	return nil
}

func (t *Target) UploadFileContents(targetPath, contents string) error {
	if target := sshTarget(t); target != nil {
		dir, _ := path.Split(targetPath)
		encodedContents := base64.StdEncoding.EncodeToString([]byte(contents))
		target.ssh("mkdir", "-p", dir)
		target.sshWithStdin(encodedContents, "base64", "-d", "-w0", fmt.Sprintf("> %s", targetPath))
	}
	return errors.New("cannot access SSH target")
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	if target := sshTarget(t); target != nil {
		if stdout, _, err := target.ssh("base64", "-w0", sourcePath); err == nil {
			decodedStdout, err := base64.StdEncoding.DecodeString(stdout)
			if err != nil {
				return "", err
			}
			return string(decodedStdout), nil
		} else {
			return "", err
		}
	}
	return "", errors.New("cannot access SSH target")
}
