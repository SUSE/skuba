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

package cluster

import (
	"github.com/SUSE/skuba/pkg/skuba"
)

var (
	scaffoldFiles = []struct {
		Location    string
		Content     string
		DoNotRender bool
	}{
		{
			Location: skuba.KubeadmInitConfFile(),
			Content:  kubeadmInitConf,
		},
		{
			Location: skuba.MasterConfTemplateFile(),
			Content:  masterConfTemplate,
		},
		{
			Location: skuba.WorkerConfTemplateFile(),
			Content:  workerConfTemplate,
		},
		{
			Location: skuba.CiliumManifestFile(),
			Content:  ciliumManifest,
			// cilium.yaml need delayed rendering as some
			// fields remain unknown while `cluster init`
			DoNotRender: true,
		},
		{
			Location: skuba.PspUnprivManifestFile(),
			Content:  pspUnprivManifest,
		},
		{
			Location: skuba.PspPrivManifestFile(),
			Content:  pspPrivManifest,
		},
	}
)
