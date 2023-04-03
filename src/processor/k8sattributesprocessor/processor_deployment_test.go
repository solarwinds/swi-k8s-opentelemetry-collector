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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
)

func withDeploymentName(name string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr("k8s.deployment.name", name)
	}
}

func TestProcessorNoDeployments(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	var i int
	resourceFunc := ([]generateResourceFunc{
		withDeploymentName("deployment1"),
	})
	ctx := client.NewContext(context.Background(), client.Info{})
	m.testConsume(
		ctx,
		generateTraces(resourceFunc...),
		generateMetrics(resourceFunc...),
		generateLogs(resourceFunc...),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(i + 1)
	m.assertResourceObjectLen(i)
	m.assertResourceAttributesLen(0, 1)
}

func TestProcessorAddDeploymentLabels(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.deployment.rules.DeploymentUID = true
		kp.deployment.associations = []kube.Association{
			{
				Sources: []kube.AssociationSource{
					{
						From: "resource_attribute",
						Name: "k8s.deployment.name",
					},
				},
			},
		}
	})

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		pi := kube.DeploymentIdentifier{
			kube.PodIdentifierAttributeFromResourceAttribute("k8s.deployment.name", "deployment1"),
		}
		kp.kc.(*fakeClient).Deployments[pi] = &kube.Deployment{
			Attributes: map[string]string{
				"deployment":  "test-2323",
				"ns":          "default",
				"another tag": "value",
			},
		}
	})

	var i int
	resourceFunc := ([]generateResourceFunc{
		withDeploymentName("deployment1"),
	})
	ctx := client.NewContext(context.Background(), client.Info{})
	m.testConsume(
		ctx,
		generateTraces(resourceFunc...),
		generateMetrics(resourceFunc...),
		generateLogs(resourceFunc...),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(i + 1)
	m.assertResourceObjectLen(i)
	m.assertResourceAttributesLen(0, 4)
	m.assertResource(i, func(res pcommon.Resource) {
		require.Greater(t, res.Attributes().Len(), 0)
		assertResourceHasStringAttribute(t, res, "deployment", "test-2323")
	})
}
