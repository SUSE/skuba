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

type TargetCache struct {
	OsRelease map[string]string
}

type Target struct {
	Target     string
	Nodename   string
	Actionable Actionable
	Cache      TargetCache
}

func (t *Target) OSRelease() (map[string]string, error) {
	if len(t.Cache.OsRelease) > 0 {
		return t.Cache.OsRelease, nil
	}
	t.Cache.OsRelease = map[string]string{}
	contents, err := t.DownloadFileContents("/etc/os-release")
	if err != nil {
		return t.Cache.OsRelease, err
	}
	scanner := bufio.NewScanner(strings.NewReader(contents))
	matcher := regexp.MustCompile(`([^=]+)="?([^"]*)`)
	for scanner.Scan() {
		matches := matcher.FindAllStringSubmatch(scanner.Text(), -1)
		for _, match := range matches {
			t.Cache.OsRelease[match[1]] = match[2]
		}
	}
	return t.Cache.OsRelease, nil
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
