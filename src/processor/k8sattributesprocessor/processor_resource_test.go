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

func withResourceAttributeName(key, name string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr(key, name)
	}
}

func testProcessorAddResourceLabels(t *testing.T, resourceType, resourceName, resourceAttributeName string, kubernetesResource kube.KubernetesResource) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.resources[resourceType].rules.UID = true
		kp.resources[resourceType].associations = []kube.Association{
			{
				Sources: []kube.AssociationSource{
					{
						From: "resource_attribute",
						Name: resourceAttributeName,
					},
				},
			},
		}
	})

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		id := kube.ResourceIdentifier{
			kube.PodIdentifierAttributeFromResourceAttribute(resourceAttributeName, resourceName),
		}
		kp.kc.(*fakeClient).Resources[resourceType][id] = kubernetesResource
	})

	var i int
	resourceFunc := ([]generateResourceFunc{
		withResourceAttributeName(resourceAttributeName, resourceName),
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
		len := res.Attributes().Len()
		require.Greater(t, len, 0)
		assertResourceHasStringAttribute(t, res, resourceName, "test-2323")
	})
}

func TestProcessorNoResources(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	var i int
	resourceFunc := ([]generateResourceFunc{
		withResourceAttributeName("k8s.deployment.name", "deployment1"),
		withResourceAttributeName("k8s.statefulset.name", "statefulset1"),
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
	m.assertResourceAttributesLen(0, 2)
}

func TestProcessorAddDeploymentLabels(t *testing.T) {
	testProcessorAddResourceLabels(t, kube.MetadataFromDeployment, "deployment", "k8s.deployment.name", &kube.Resource{
		Attributes: map[string]string{
			"deployment":  "test-2323",
			"ns":          "default",
			"another tag": "value",
		},
	})
}

func TestProcessorAddStatefulSetLabels(t *testing.T) {
	testProcessorAddResourceLabels(t, kube.MetadataFromStatefulSet, "statefulset", "k8s.statefulset.name", &kube.Resource{
		Attributes: map[string]string{
			"statefulset": "test-2323",
			"ns":          "default",
			"another tag": "value",
		},
	})
}
