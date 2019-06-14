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

package deployments

import (
	"strings"

	"k8s.io/klog"
)

type Role uint

const (
	MasterRole Role = iota
	WorkerRole Role = iota
)

func MustGetRoleFromString(s string) (role Role) {
	switch strings.ToLower(s) {
	case "master":
		role = MasterRole
	case "worker":
		role = WorkerRole
	default:
		klog.Fatalf("[join] invalid role provided: %q, 'master' or 'worker' are the only accepted roles", s)
	}
	return
}

type JoinConfiguration struct {
	Role             Role
	KubeadmExtraArgs map[string]string
}
