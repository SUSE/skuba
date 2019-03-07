package deployments

import (
	"bufio"
	"regexp"
	"strings"
)

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
