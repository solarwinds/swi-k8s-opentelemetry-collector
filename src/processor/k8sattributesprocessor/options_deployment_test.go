// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sattributesprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
)

func TestWithExtractDeploymentAssociation(t *testing.T) {
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
				deployment: &kubernetesProcessorDeployment{},
			}
			opt := withExtractDeploymentAssociations(tt.args...)
			assert.NoError(t, opt(p))
			assert.Equal(t, tt.want, p.deployment.associations)
		})
	}
}
