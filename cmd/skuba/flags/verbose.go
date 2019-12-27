/*
 * Copyright (c) 2020 SUSE LLC.
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

package flags

import (
	"flag"

	"github.com/spf13/pflag"
)

const (
	verboseLevelFlagShort = "v"
	verboseLevelFlagLong  = "verbosity"
	verboseLevelFlagUsage = "log level [0-5]. 0 (Only Error and Warning) to 5 (Maximum detail)."
)

// GetVerboseFlagLevel returns verbose flag level.
func GetVerboseFlagLevel() string {
	if f := flag.CommandLine.Lookup(verboseLevelFlagShort); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		return pflagFlag.Value.String()
	}

	return "0"
}

// RegisterVerboseFlag register verbose flag.
func RegisterVerboseFlag(local *pflag.FlagSet) {
	if f := flag.CommandLine.Lookup(verboseLevelFlagShort); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		pflagFlag.Name = verboseLevelFlagLong
		pflagFlag.Usage = verboseLevelFlagUsage
		local.AddFlag(pflagFlag)
	}
}
