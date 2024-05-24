// Copyright 2022 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swk8sattributesprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"
)

func TestWithExtractResourceAssociation(t *testing.T) {
	tests := []struct {
		name string
		args []AssociationConfig
		want []kube.Association
	}{
		{
			"empty",
			[]AssociationConfig{},
			[]kube.Association{},
		},
		{
			"basic",
			[]AssociationConfig{
				{
					Sources: []AssociationSourceConfig{
						{
							From: "label",
							Name: "ip",
						},
					},
				},
			},
			[]kube.Association{
				{
					Sources: []kube.AssociationSource{
						{
							From: "label",
							Name: "ip",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{
				resources: map[string]*kubernetesProcessorResource{
					kube.MetadataFromDeployment: {},
				},
			}
			opt := withExtractResourceAssociations(kube.MetadataFromDeployment, tt.args...)
			assert.NoError(t, opt(p))
			assert.Equal(t, tt.want, p.resources[kube.MetadataFromDeployment].associations)
		})
	}
}
