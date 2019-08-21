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
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

const validConfig = `---
apiVersion: kubeadm.k8s.io/v1beta1
caCertPath: /etc/kubernetes/pki/ca.crt
discovery:
  bootstrapToken:
    apiServerEndpoint: 10.84.152.198:6443
    token: j2sxf2.s3q02ll7nvbkjjr0
    unsafeSkipCAVerification: true
  timeout: 5m0s
  tlsBootstrapToken: ""
kind: JoinConfiguration
nodeRegistration:
  criSocket: /var/run/crio/crio.sock
  kubeletExtraArgs:
    cni-bin-dir: /usr/lib/cni
    hostname-override: worker-0
    pod-infra-container-image: registry.suse.de/devel/caasp/4.0/containers/containers/caasp/v4/pause:3.1
  name: worker-0
`

func Test_LoadInitConfigurationFromFile(t *testing.T) {
	tests := []struct {
		name          string
		cfgPath       string
		expectedError bool
	}{
		{
			name:    "normal",
			cfgPath: "testdata/init.conf",
		},
		{
			name:    "cluster configuration only",
			cfgPath: "testdata/cluster.conf",
		},
		{
			name:          "config path not exist",
			cfgPath:       "testdata/not-exist.conf",
			expectedError: true,
		},
		{
			name:          "invalid api version",
			cfgPath:       "testdata/invalid.conf",
			expectedError: true,
		},
		{
			name:          "not init or cluster configuration",
			cfgPath:       "testdata/join.conf",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadInitConfigurationFromFile(tt.cfgPath)
			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}

func Test_LoadJoinConfigurationFromFile(t *testing.T) {
	tests := []struct {
		name          string
		cfgPath       string
		expectedError bool
	}{
		{
			name:    "normal",
			cfgPath: "testdata/join.conf",
		},
		{
			name:          "invalid yaml file",
			cfgPath:       "testdata/invalid.conf",
			expectedError: true,
		},
		{
			name:          "config path not exist",
			cfgPath:       "testdata/not-exist.conf",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadJoinConfigurationFromFile(tt.cfgPath)
			if tt.expectedError && err == nil {
				t.Errorf("error expected on %s, but no error reported", tt.name)
				return
			} else if !tt.expectedError && err != nil {
				t.Errorf("error not expected on %s, but an error was reported (%v)", tt.name, err)
				return
			}
		})
	}
}

func Test_documentMapToJoinConfiguration(t *testing.T) {
	type args struct {
		gvkmap          map[schema.GroupVersionKind][]byte
		allowDeprecated bool
	}
	tests := []struct {
		name    string
		args    args
		want    *kubeadmapi.JoinConfiguration
		wantErr bool
	}{
		{
			name: "Map a JoinConfiguration",
			args: args{
				gvkmap: map[schema.GroupVersionKind][]byte{
					schema.GroupVersionKind{
						Group:   "",
						Version: "kubeadm.k8s.io/v1beta1",
						Kind:    "JoinConfiguration",
					}: []byte(validConfig),
				},
				allowDeprecated: false,
			},
			want: &kubeadmapi.JoinConfiguration{
				TypeMeta: metav1.TypeMeta{},
				NodeRegistration: kubeadmapi.NodeRegistrationOptions{
					Name:      "worker-0",
					CRISocket: "/var/run/crio/crio.sock",
					Taints:    nil,
					KubeletExtraArgs: map[string]string{
						"cni-bin-dir":               "/usr/lib/cni",
						"hostname-override":         "worker-0",
						"pod-infra-container-image": "registry.suse.de/devel/caasp/4.0/containers/containers/caasp/v4/pause:3.1",
					},
				},
				CACertPath: "/etc/kubernetes/pki/ca.crt",
				Discovery: kubeadmapi.Discovery{
					BootstrapToken: &kubeadmapi.BootstrapTokenDiscovery{
						Token:                    "j2sxf2.s3q02ll7nvbkjjr0",
						APIServerEndpoint:        "10.84.152.198:6443",
						CACertHashes:             nil,
						UnsafeSkipCAVerification: true,
					},
					File:              nil,
					TLSBootstrapToken: "j2sxf2.s3q02ll7nvbkjjr0",
					Timeout: &metav1.Duration{
						Duration: 300000000000,
					},
				},
				ControlPlane: nil,
			},
			wantErr: false,
		},
		{
			name: "Attempt to map another configuration",
			args: args{
				gvkmap: map[schema.GroupVersionKind][]byte{
					schema.GroupVersionKind{
						Group:   "",
						Version: "kubeadm.k8s.io/v1beta1",
						Kind:    "Something else",
					}: make([]byte, 0),
				},
				allowDeprecated: false,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			got, err := documentMapToJoinConfiguration(tt.args.gvkmap, tt.args.allowDeprecated)
			if (err != nil) != tt.wantErr {
				t.Errorf("documentMapToJoinConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("documentMapToJoinConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}
