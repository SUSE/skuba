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

	"k8s.io/client-go/kubernetes/fake"

	"github.com/SUSE/skuba/internal/pkg/skuba/kubernetes"
	"github.com/SUSE/skuba/internal/pkg/skuba/skuba"
	img "github.com/SUSE/skuba/pkg/skuba"
)

func TestGetGangwayImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get gangway init image without revision",
			imageTag: "3.1.0",
			want:     img.ImageRepository + "/gangway:3.1.0",
		},
		{
			name:     "get gangway init image with revision",
			imageTag: "3.1.0-rev2",
			want:     img.ImageRepository + "/gangway:3.1.0-rev2",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGangwayImage(tt.imageTag); got != tt.want {
				t.Errorf("GetGangwayImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderContext_GangwayImage(t *testing.T) {
	type test struct {
		name          string
		renderContext renderContext
		want          string
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name: "render gangway image url when cluster version is " + ver.String(),
			renderContext: renderContext{
				config: AddonConfiguration{
					ClusterVersion: ver,
					ControlPlane:   "",
					ClusterName:    "",
				},
			},
			want: img.ImageRepository + "/gangway:([[:digit:]]{1,}.){2}[[:digit:]]{1,}(-rev[:digit:]{1,})?",
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.renderContext.GangwayImage()
			matched, err := regexp.Match(tt.want, []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			if !matched {
				t.Errorf("renderContext.GangwayImage() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_renderContext_GangwayClientSecret(t *testing.T) {
	for _, ver := range kubernetes.AvailableVersions() {
		t.Run("get gangway client secret when cluster version is "+ver.String(), func(t *testing.T) {
			fake.NewSimpleClientset()

			got := renderContext{}.GangwayClientSecret()
			if got != "" {
				t.Errorf("expect got client secret empty")
				return
			}
		})
	}
}

func Test_gangwayCallbacks_beforeApply(t *testing.T) {
	type test struct {
		name               string
		addonConfiguration AddonConfiguration
		skubaConfiguration *skuba.SkubaConfiguration
		wantErr            bool
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name:               "test gangway callbacks beforeApply when cluster version is " + ver.String(),
			addonConfiguration: AddonConfiguration{},
			wantErr:            true,
		}
		t.Run(tt.name, func(t *testing.T) {
			g := gangwayCallbacks{}
			if err := g.beforeApply(tt.addonConfiguration, tt.skubaConfiguration); (err != nil) != tt.wantErr {
				t.Errorf("gangwayCallbacks.beforeApply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
