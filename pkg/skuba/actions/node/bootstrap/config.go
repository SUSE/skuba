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

package bootstrap

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
	"k8s.io/kubernetes/cmd/kubeadm/app/componentconfigs"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"
)

// LoadInitConfigurationFromFile loads a supported versioned InitConfiguration from a file, converts it into internal config, defaults it and verifies it.
func LoadInitConfigurationFromFile(cfgPath string) (*kubeadmapi.InitConfiguration, error) {
	klog.V(1).Infof("loading configuration from %q", cfgPath)

	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read config from %q ", cfgPath)
	}

	return BytesToInitConfiguration(b)
}

// BytesToInitConfiguration converts a byte slice to an internal, defaulted and validated InitConfiguration object.
// The map may contain many different YAML documents. These YAML documents are parsed one-by-one
// and well-known ComponentConfig GroupVersionKinds are stored inside of the internal InitConfiguration struct.
// The resulting InitConfiguration is then dynamically defaulted and validated prior to return.
func BytesToInitConfiguration(b []byte) (*kubeadmapi.InitConfiguration, error) {
	gvkmap, err := kubeadmutil.SplitYAMLDocuments(b)
	if err != nil {
		return nil, err
	}

	return documentMapToInitConfiguration(gvkmap, false)
}

// documentMapToInitConfiguration converts a map of GVKs and YAML documents to defaulted and validated configuration object.
func documentMapToInitConfiguration(gvkmap map[schema.GroupVersionKind][]byte, allowDeprecated bool) (*kubeadmapi.InitConfiguration, error) {
	var initcfg *kubeadmapi.InitConfiguration
	var clustercfg *kubeadmapi.ClusterConfiguration
	decodedComponentConfigObjects := map[componentconfigs.RegistrationKind]runtime.Object{}

	for gvk, fileContent := range gvkmap {
		// verify the validity of the YAML
		strict.VerifyUnmarshalStrict(fileContent, gvk)

		// Try to get the registration for the ComponentConfig based on the kind
		regKind := componentconfigs.RegistrationKind(gvk.Kind)
		if registration, found := componentconfigs.Known[regKind]; found {
			// Unmarshal the bytes from the YAML document into a runtime.Object containing the ComponentConfiguration struct
			obj, err := registration.Unmarshal(fileContent)
			if err != nil {
				return nil, err
			}
			decodedComponentConfigObjects[regKind] = obj
			continue
		}

		if kubeadmutil.GroupVersionKindsHasInitConfiguration(gvk) {
			// Set initcfg to an empty struct value the deserializer will populate
			initcfg = &kubeadmapi.InitConfiguration{}
			// Decode the bytes into the internal struct. Under the hood, the bytes will be unmarshalled into the
			// right external version, defaulted, and converted into the internal version.
			if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), fileContent, initcfg); err != nil {
				return nil, err
			}
			continue
		}
		if kubeadmutil.GroupVersionKindsHasClusterConfiguration(gvk) {
			// Set clustercfg to an empty struct value the deserializer will populate
			clustercfg = &kubeadmapi.ClusterConfiguration{}
			// Decode the bytes into the internal struct. Under the hood, the bytes will be unmarshalled into the
			// right external version, defaulted, and converted into the internal version.
			if err := runtime.DecodeInto(kubeadmscheme.Codecs.UniversalDecoder(), fileContent, clustercfg); err != nil {
				return nil, err
			}
			continue
		}

		fmt.Printf("[config] WARNING: Ignored YAML document with GroupVersionKind %v\n", gvk)
	}

	// Enforce that InitConfiguration and/or ClusterConfiguration has to exist among the YAML documents
	if initcfg == nil && clustercfg == nil {
		return nil, errors.New("no InitConfiguration or ClusterConfiguration kind was found in the YAML file")
	}

	// If InitConfiguration wasn't given, default it by creating an external struct instance, default it and convert into the internal type
	if initcfg == nil {
		extinitcfg := &kubeadmapiv1beta1.InitConfiguration{}
		kubeadmscheme.Scheme.Default(extinitcfg)
		// Set initcfg to an empty struct value the deserializer will populate
		initcfg = &kubeadmapi.InitConfiguration{}
		kubeadmscheme.Scheme.Convert(extinitcfg, initcfg, nil)
	}
	// If ClusterConfiguration was given, populate it in the InitConfiguration struct
	if clustercfg != nil {
		initcfg.ClusterConfiguration = *clustercfg
	}

	// Save the loaded ComponentConfig objects in the initcfg object
	for kind, obj := range decodedComponentConfigObjects {
		if registration, found := componentconfigs.Known[kind]; found {
			if ok := registration.SetToInternalConfig(obj, &initcfg.ClusterConfiguration); !ok {
				return nil, errors.Errorf("couldn't save componentconfig value for kind %q", string(kind))
			}
		} else {
			// This should never happen in practice
			fmt.Printf("[config] WARNING: Decoded a kind that couldn't be saved to the internal configuration: %q\n", string(kind))
		}
	}

	return initcfg, nil
}
