package deployments

import (
	"bufio"
	"io/ioutil"
	"regexp"
	"strings"
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

func (t *Target) OSRelease() (map[string]string, error) {
	result := map[string]string{}
	contents, err := t.DownloadFileContents("/etc/os-release")
	if err != nil {
		return result, err
	}
	scanner := bufio.NewScanner(strings.NewReader(contents))
	matcher := regexp.MustCompile(`([^=]+)="?([^"]*)`)
	for scanner.Scan() {
		matches := matcher.FindAllStringSubmatch(scanner.Text(), -1)
		for _, match := range matches {
			result[match[1]] = match[2]
		}
	}
	return result, nil
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
