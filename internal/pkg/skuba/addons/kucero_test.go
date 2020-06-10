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

package addons

import (
	"testing"

	img "github.com/SUSE/skuba/pkg/skuba"
)

func TestGetKuceroImage(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     string
	}{
		{
			name:     "get kucero image without revision",
			imageTag: "1.1.1",
			want:     img.ImageRepository + "/kucero:1.1.1",
		},
		{
			name:     "get kucero image with revision",
			imageTag: "1.1.1-rev1",
			want:     img.ImageRepository + "/kucero:1.1.1-rev1",
		},
	}
	for _, tt := range tests {
		tt := tt // Parallel testing
		t.Run(tt.name, func(t *testing.T) {
			if got := GetKuceroImage(tt.imageTag); got != tt.want {
				t.Errorf("GetKuceroImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
