package deployments

import (
	"io/ioutil"
)

type Actionable interface {
	Apply(data interface{}, states ...string) error
	UploadFileContents(targetPath, contents string) error
	DownloadFileContents(sourcePath string) (string, error)
}

type Target struct {
	Node       string
	Actionable Actionable
}

func (t *Target) Target() string {
	return t.Node
}

func (t *Target) Apply(data interface{}, states ...string) error {
	return t.Actionable.Apply(data, states...)
}

func (t *Target) UploadFile(sourcePath, targetPath string) error {
	if contents, err := ioutil.ReadFile(sourcePath); err == nil {
		return t.UploadFileContents(targetPath, string(contents))
	}
	return nil
}

func (t *Target) UploadFileContents(targetPath, contents string) error {
	return t.Actionable.UploadFileContents(targetPath, contents)
}

func (t *Target) DownloadFileContents(sourcePath string) (string, error) {
	return t.Actionable.DownloadFileContents(sourcePath)
}
