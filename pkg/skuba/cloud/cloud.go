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

package cloud

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/pkg/skuba"
	"github.com/SUSE/skuba/pkg/skuba/actions"
)

func HasCloudIntegration(action actions.Action) bool {
	switch action {
	case actions.Bootstrap:
		if _, err := os.Stat(skuba.OpenstackCloudConfFile()); err == nil {
			return true
		} else if _, err := os.Stat(skuba.OpenstackCloudConfTemplateFile()); err == nil {
			klog.Fatalf("%q file exists, but %q file does not. Please create this file with the expected contents to enable cloud integration", skuba.OpenstackCloudConfTemplateFile(), skuba.OpenstackCloudConfFile())
		}
	case actions.Join:
		client, err := kubernetes.GetAdminClientSet()
		if err != nil {
			return false
		}
		_, err = client.CoreV1().Secrets(metav1.NamespaceSystem).Get(kubernetes.OpenstackCloudControllerManagerSecretName, metav1.GetOptions{})
		return err == nil
	}
	return false
}
