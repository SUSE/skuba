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

package skuba

import (
	"fmt"
	"runtime"
)

var (
	Version   string
	BuildDate string
	Commit    string
)

type SkubaVersion struct {
	Version   string
	BuildType string
	BuildDate string
	Commit    string
	GoVersion string
}

func CurrentVersion() SkubaVersion {
	return SkubaVersion{
		Version:   Version,
		BuildType: BuildType,
		BuildDate: BuildDate,
		Commit:    Commit,
		GoVersion: runtime.Version(),
	}
}

func (s SkubaVersion) String() string {
	return fmt.Sprintf("skuba version: %s (%s) %s %s %s", s.Version, s.BuildType, s.Commit, s.BuildDate, s.GoVersion)
}
