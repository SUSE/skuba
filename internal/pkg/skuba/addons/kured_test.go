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
	img "github.com/SUSE/skuba/pkg/skuba"
)

func TestGetKuredImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get kured image without revision",
			imageTag: "1.2.0",
			want:     img.ImageRepository + "/kured:1.2.0",
		},
		{
			name:     "get kured image with revision",
			imageTag: "1.2.0-rev4",
			want:     img.ImageRepository + "/kured:1.2.0-rev4",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetKuredImage(tt.imageTag); got != tt.want {
				t.Errorf("GetKuredImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_renderContext_KuredImage(t *testing.T) {
	type test struct {
		name          string
		renderContext renderContext
		want          string
	}
	for _, ver := range kubernetes.AvailableVersions() {
		tt := test{
			name: "render kured image URL when cluster version is " + ver.String(),
			renderContext: renderContext{
				config: AddonConfiguration{
					ClusterVersion: ver,
					ControlPlane:   "",
					ClusterName:    "",
				},
			},
			want: img.ImageRepository + "/kured:([[:digit:]]{1,}.){2}[[:digit:]]{1,}(-rev[:digit:]{1,})?",
		}
		t.Run(tt.name, func(t *testing.T) {
			got := tt.renderContext.KuredImage()
			matched, err := regexp.Match(tt.want, []byte(got))
			if err != nil {
				t.Error(err)
				return
			}
			if !matched {
				t.Errorf("renderContext.KuredImage() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
