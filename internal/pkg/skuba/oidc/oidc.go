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

package oidc

const (
	// dex certificate
	DexCertCN         = "oidc-dex"
	DexCertSecretName = "oidc-dex-cert"
	// gangway certificate
	GangwayCertCN         = "oidc-gangway"
	GangwayCertSecretName = "oidc-gangway-cert"

	// oidc client secret
	ClientSecretName        = "oidc-client-secret"
	ClientSecretKey_Gangway = "gangway"

	// gangway session key
	GangwaySecretName        = "oidc-gangway-secret"
	GangwaySecret_SessionKey = "session-key"
)
