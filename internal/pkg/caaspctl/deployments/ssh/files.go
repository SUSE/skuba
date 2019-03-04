package ssh

import (
	"encoding/base64"
	"fmt"
	"log"
	"path"
)

func (t *Target) UploadFileContents(targetPath, contents string) error {
	log.Printf("uploading file %s", targetPath)
	dir, _ := path.Split(targetPath)
	encodedContents := base64.StdEncoding.EncodeToString([]byte(contents))
	if _, _, err := t.silentSsh("mkdir", "-p", dir); err != nil {
		return err
	}
	_, _, err := t.silentSshWithStdin(encodedContents, "base64", "-d", "-w0", fmt.Sprintf("> %s", targetPath))
	return err
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	log.Printf("downloading file %s", sourcePath)
	if stdout, _, err := t.silentSsh("base64", "-w0", sourcePath); err == nil {
		decodedStdout, err := base64.StdEncoding.DecodeString(stdout)
		if err != nil {
			return "", err
		}
		return string(decodedStdout), nil
	} else {
		return "", err
	}
}
