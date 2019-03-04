package ssh

import (
	"encoding/base64"
	"fmt"
	"path"
)

func (t *Target) UploadFileContents(targetPath, contents string) error {
	dir, _ := path.Split(targetPath)
	encodedContents := base64.StdEncoding.EncodeToString([]byte(contents))
	t.ssh("mkdir", "-p", dir)
	_, _, err := t.sshWithStdin(encodedContents, "base64", "-d", "-w0", fmt.Sprintf("> %s", targetPath))
	return err
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	if stdout, _, err := t.ssh("base64", "-w0", sourcePath); err == nil {
		decodedStdout, err := base64.StdEncoding.DecodeString(stdout)
		if err != nil {
			return "", err
		}
		return string(decodedStdout), nil
	} else {
		return "", err
	}
}
