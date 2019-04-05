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

package node

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"
)

// LoadJoinConfigurationFromFile loads versioned JoinConfiguration from file, converts it to internal, defaults and validates it
func LoadJoinConfigurationFromFile(cfgPath string) (*kubeadmapi.JoinConfiguration, error) {
	klog.V(1).Infof("loading configuration from %q", cfgPath)

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config from %q ", cfgPath)
	}

	gvkmap, err := kubeadmutil.SplitYAMLDocuments(b)
	if err != nil {
		return nil, err
	}

	return documentMapToJoinConfiguration(gvkmap, false)
}

// documentMapToJoinConfiguration takes a map between GVKs and YAML documents (as returned by SplitYAMLDocuments),
// finds a JoinConfiguration, decodes it, dynamically defaults it and then validates it prior to return.
func documentMapToJoinConfiguration(gvkmap map[schema.GroupVersionKind][]byte, allowDeprecated bool) (*kubeadmapi.JoinConfiguration, error) {
	joinBytes := []byte{}
	for gvk, bytes := range gvkmap {
		// not interested in anything other than JoinConfiguration
		if gvk.Kind != constants.JoinConfigurationKind {
			continue
		}

		// verify the validity of the YAML
		strict.VerifyUnmarshalStrict(bytes, gvk)

		joinBytes = bytes
	}

	if len(joinBytes) == 0 {
		return nil, errors.Errorf("no %s found in the supplied config", constants.JoinConfigurationKind)
	}

	internalcfg := &kubeadmapi.JoinConfiguration{}
	if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), joinBytes, internalcfg); err != nil {
		return nil, err
	}

	return internalcfg, nil
}
