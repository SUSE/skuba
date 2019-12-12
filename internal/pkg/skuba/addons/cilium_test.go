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

package addons

import (
	"regexp"
	"testing"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	img "github.com/SUSE/skuba/pkg/skuba"
)

func TestGetCiliumInitImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get cilium init image without revision",
			imageTag: "1.5.3",
			want:     img.ImageRepository + "/cilium-init:1.5.3",
		},
		{
			name:     "get cilium init image with revision",
			imageTag: "1.5.3-rev2",
			want:     img.ImageRepository + "/cilium-init:1.5.3-rev2",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCiliumInitImage(tt.imageTag); got != tt.want {
				t.Errorf("GetCiliumInitImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCiliumOperatorImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get cilium operator image without revision",
			imageTag: "1.5.3",
			want:     img.ImageRepository + "/cilium-operator:1.5.3",
		},
		{
			name:     "get cilium operator image with revision",
			imageTag: "1.5.3-rev2",
			want:     img.ImageRepository + "/cilium-operator:1.5.3-rev2",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCiliumOperatorImage(tt.imageTag); got != tt.want {
				t.Errorf("GetCiliumOperatorImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCiliumImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get cilium image without revision",
			imageTag: "1.5.3",
			want:     img.ImageRepository + "/cilium:1.5.3",
		},
		{
			name:     "get cilium image with revision",
			imageTag: "1.5.3-rev2",
			want:     img.ImageRepository + "/cilium:1.5.3-rev2",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCiliumImage(tt.imageTag); got != tt.want {
				t.Errorf("GetCiliumImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderContext_CiliumInitImage(t *testing.T) {
	type test struct {
		name          string
		renderContext renderContext
		want          string
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name: "render Cilium Init Image URL when cluster version is " + ver.String(),
			renderContext: renderContext{
				config: AddonConfiguration{
					ClusterVersion: ver,
					ControlPlane:   "",
					ClusterName:    "",
				},
			},
			want: img.ImageRepository + "/cilium-init:([[:digit:]]{1,}.){2}[[:digit:]]{1,}(-rev[:digit:]{1,})?",
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.renderContext.CiliumInitImage()
			matched, err := regexp.Match(tt.want, []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			if !matched {
				t.Errorf("renderContext.CiliumInitImage() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_renderContext_CiliumOperatorImage(t *testing.T) {
	type test struct {
		name          string
		renderContext renderContext
		want          string
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name: "render Cilium Operator Image URL when cluster version is " + ver.String(),
			renderContext: renderContext{
				config: AddonConfiguration{
					ClusterVersion: ver,
					ControlPlane:   "",
					ClusterName:    "",
				},
			},
			want: img.ImageRepository + "/cilium-operator:([[:digit:]]{1,}.){2}[[:digit:]]{1,}(-rev[:digit:]{1,})?",
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.renderContext.CiliumOperatorImage()
			matched, err := regexp.Match(tt.want, []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			if !matched {
				t.Errorf("renderContext.CiliumOperatorImage() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_renderContext_CiliumImage(t *testing.T) {
	type test struct {
		name          string
		renderContext renderContext
		want          string
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name: "render Cilium Image URL when cluster version is " + ver.String(),
			renderContext: renderContext{
				config: AddonConfiguration{
					ClusterVersion: ver,
					ControlPlane:   "",
					ClusterName:    "",
				},
			},
			want: img.ImageRepository + "/cilium:([[:digit:]]{1,}.){2}[[:digit:]]{1,}(-rev[:digit:]{1,})?",
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.renderContext.CiliumImage()
			matched, err := regexp.Match(tt.want, []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			if !matched {
				t.Errorf("renderContext.CiliumImage() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_ciliumCallbacks_beforeApply(t *testing.T) {
	type test struct {
		name               string
		addonConfiguration AddonConfiguration
		skubaConfiguration *skuba.SkubaConfiguration
		wantErr            bool
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name:               "Test Cilium callbacks beforeApply when cluster version is " + ver.String(),
			addonConfiguration: AddonConfiguration{},
			wantErr:            true,
		}
		t.Run(tt.name, func(t *testing.T) {
			c := ciliumCallbacks{}
			if err := c.beforeApply(tt.addonConfiguration, tt.skubaConfiguration); (err != nil) != tt.wantErr {
				t.Errorf("ciliumCallbacks.beforeApply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
