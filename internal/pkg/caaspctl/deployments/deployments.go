package deployments

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
)

type Actionable interface {
	Apply(data interface{}, states ...string) error
	UploadFileContents(targetPath, contents string) error
	DownloadFileContents(sourcePath string) (string, error)
}

type TargetCache struct {
	OsRelease map[string]string
}

type Target struct {
	Target     string
	Nodename   string
	Actionable Actionable
	Cache      TargetCache
}

func (t *Target) Apply(data interface{}, states ...string) error {
	return t.Actionable.Apply(data, states...)
}

func (t *Target) UploadFile(sourcePath, targetPath string) error {
	log.Printf("uploading file %q to %q", sourcePath, targetPath)
	if contents, err := ioutil.ReadFile(sourcePath); err == nil {
		return t.UploadFileContents(targetPath, string(contents))
	}
	return errors.New(fmt.Sprintf("could not find file %s", sourcePath))
}

func (t *Target) UploadFileContents(targetPath, contents string) error {
	return t.Actionable.UploadFileContents(targetPath, contents)
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	return t.Actionable.DownloadFileContents(sourcePath)
}
