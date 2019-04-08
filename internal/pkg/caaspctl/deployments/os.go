/*
 * Copyright (c) 2019 SUSE LLC. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package deployments

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	SUSEOSID = "suse"
)

func (t *Target) osRelease() (map[string]string, error) {
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

func (t *Target) hasOS(os string) (bool, error) {
	osRelease, err := t.osRelease()
	if err != nil {
		return false, errors.Wrap(err, "could not retrieve OS release information")
	}

	if strings.Contains(osRelease["ID_LIKE"], os) {
		return true, nil
	}

	return false, nil
}

func (t *Target) IsSUSEOS() (bool, error) {
	return t.hasOS(SUSEOSID)
}
