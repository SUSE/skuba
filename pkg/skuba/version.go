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
	Version    string
	BuildDate  string
	Tag        string
	ClosestTag string
)

type SkubaVersion struct {
	Version   string
	BuildType string
	BuildDate string
	Tag       string
	GoVersion string
}

func CurrentVersion() SkubaVersion {
	skubaVersion := SkubaVersion{
		Version:   Version,
		BuildType: BuildType,
		BuildDate: BuildDate,
		Tag:       Tag,
		GoVersion: runtime.Version(),
	}
	if skubaVersion.Tag == "" {
		skubaVersion.Version = fmt.Sprintf("untagged (%s)", ClosestTag)
	}
	return skubaVersion
}

func (s SkubaVersion) String() string {
	if s.Tag == "" {
		return fmt.Sprintf("skuba version: %s (%s) %s %s", s.Version, s.BuildType, s.BuildDate, s.GoVersion)
	}
	return fmt.Sprintf("skuba version: %s (%s) (tagged as %q) %s %s", s.Version, s.BuildType, s.Tag, s.BuildDate, s.GoVersion)
}
