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

package retry

import "k8s.io/apimachinery/pkg/util/wait"

// OnAnyError retries the fn function with backoff strategy in the
// light of any error
func OnAnyError(backoff wait.Backoff, fn func() error) error {
	return OnError(
		backoff,
		func(error) bool {
			return true
		},
		fn,
	)
}

// OnError executes the provided function repeatedly, retrying if the server returns a specified
// error. Callers should preserve previous executions if they wish to retry changes. It performs an
// exponential backoff.
//
//     var pod *api.Pod
//     err := retry.OnError(DefaultBackoff, errors.IsConflict, func() (err error) {
//       pod, err = c.Pods("mynamespace").UpdateStatus(podStatus)
//       return
//     })
//     if err != nil {
//       // may be conflict if max retries were hit
//       return err
//     }
//     ...
//
// Extracted from k8s.io/kubernetes/staging/src/k8s.io/client-go/util/retry package,
// available in 1.16.0 or newer.
func OnError(backoff wait.Backoff, errorFunc func(error) bool, fn func() error) error {
	var lastConflictErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case errorFunc(err):
			lastConflictErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastConflictErr
	}
	return err
}
