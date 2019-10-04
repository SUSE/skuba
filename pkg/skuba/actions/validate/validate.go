/*
 * Copyright (c) 2019 SUSE LLC.
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

package validate

import (
	"fmt"
	"regexp"

	"github.com/SUSE/skuba/pkg/skuba"
)

var nodeNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// NodeName checks whether the node name is valid and accpetable for kubelet.
func NodeName(nodename string) error {
	if len(nodename) > skuba.MaxNodeNameLength {
		return fmt.Errorf(
			"invalid node name \"%s\": must be no more than %d characters",
			nodename, skuba.MaxNodeNameLength)
	}

	if !nodeNameRegex.MatchString(nodename) {
		return fmt.Errorf(
			"invalid node name \"%s\": must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
			nodename)
	}
	return nil
}
