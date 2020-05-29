/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/config/strict"
)

// validateSupportedVersion checks if the supplied GroupVersion is not on the lists of old unsupported or deprecated GVs.
// If it is, an error is returned.
func validateSupportedVersion(gv schema.GroupVersion, allowDeprecated bool) error {
	// The support matrix will look something like this now and in the future:
	// v1.10 and earlier: v1alpha1
	// v1.11: v1alpha1 read-only, writes only v1alpha2 config
	// v1.12: v1alpha2 read-only, writes only v1alpha3 config. Errors if the user tries to use v1alpha1
	// v1.13: v1alpha3 read-only, writes only v1beta1 config. Errors if the user tries to use v1alpha1 or v1alpha2
	// v1.14: v1alpha3 convert only, writes only v1beta1 config. Errors if the user tries to use v1alpha1 or v1alpha2
	// v1.15: v1beta1 read-only, writes only v1beta2 config. Errors if the user tries to use v1alpha1, v1alpha2 or v1alpha3
	oldKnownAPIVersions := map[string]string{
		"kubeadm.k8s.io/v1alpha1": "v1.11",
		"kubeadm.k8s.io/v1alpha2": "v1.12",
		"kubeadm.k8s.io/v1alpha3": "v1.14",
	}

	// Deprecated API versions are supported by us, but can only be used for migration.
	deprecatedAPIVersions := map[string]struct{}{
		"kubeadm.k8s.io/v1beta1": {},
	}

	gvString := gv.String()

	if useKubeadmVersion := oldKnownAPIVersions[gvString]; useKubeadmVersion != "" {
		return errors.Errorf("your configuration file uses an old API spec: %q. Please use kubeadm %s instead and run 'kubeadm config migrate --old-config old.yaml --new-config new.yaml', which will write the new, similar spec using a newer API version.", gv.String(), useKubeadmVersion)
	}

	if _, present := deprecatedAPIVersions[gvString]; present && !allowDeprecated {
		klog.Warningf("your configuration file uses a deprecated API spec: %q. Please use 'kubeadm config migrate --old-config old.yaml --new-config new.yaml', which will write the new, similar spec using a newer API version.", gv)
	}

	return nil
}

// documentMapToJoinConfiguration takes a map between GVKs and YAML documents (as returned by SplitYAMLDocuments),
// finds a JoinConfiguration, decodes it, dynamically defaults it and then validates it prior to return.
func documentMapToJoinConfiguration(gvkmap kubeadmapi.DocumentMap, allowDeprecated bool) (*kubeadmapi.JoinConfiguration, error) {
	joinBytes := []byte{}
	for gvk, bytes := range gvkmap {
		// not interested in anything other than JoinConfiguration
		if gvk.Kind != constants.JoinConfigurationKind {
			continue
		}

		// check if this version is supported and possibly not deprecated
		if err := validateSupportedVersion(gvk.GroupVersion(), allowDeprecated); err != nil {
			return nil, err
		}

		// verify the validity of the YAML
		err := strict.VerifyUnmarshalStrict(bytes, gvk)
		if err != nil {
			return nil, err
		}

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
