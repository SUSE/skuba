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

package dex

import (
	"crypto/rand"
	"fmt"
)

const (
	// CertCommonName is the dex server certificate CN
	CertCommonName = "oidc-dex"
	// PodLabelName is the dex pod label name
	PodLabelName = "app=oidc-dex"
	// CertSecretName is the dex certificate secret name
	CertSecretName = "oidc-dex-cert"
)

// GenerateClientSecret returns client secret which is used by
// auth client (gangway) to authenticate to auth server (dex)
//
// Due to this issue https://github.com/dexidp/dex/issues/1099
// client secret is not configurable through environment variable
// so, replace client secret in configmap by rendering
func GenerateClientSecret() string {
	b := make([]byte, 12)
	_, err := rand.Read(b)
	if err != nil {
		// TODO: handle the error correctly
		fmt.Println("error generating random bytes:", err)
	}
	return fmt.Sprintf("%x", b)
}
