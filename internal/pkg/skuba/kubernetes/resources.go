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

package kubernetes

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// DoesResourceExistWithError check the given error from a client-go API call
// and returns whether the resource exists. If no error was given, it will
// return true, if the given error was a `IsNotFound` it will return false;
// otherwise it will return an error
func DoesResourceExistWithError(err error) (bool, error) {
	if err == nil {
		return true, nil
	} else if apierrors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}
