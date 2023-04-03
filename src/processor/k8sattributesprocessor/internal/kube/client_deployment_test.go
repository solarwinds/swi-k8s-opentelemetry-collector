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

package kube

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDeploymentIdentifier(from string, name string, value string) DeploymentIdentifier {
	return DeploymentIdentifier{
		{
			Source: AssociationSource{
				From: from,
				Name: name,
			},
			Value: value,
		},
	}
}

func deploymentAddAndUpdateTest(t *testing.T, c *WatchClient, handler func(obj interface{})) {
	assert.Equal(t, 0, len(c.DeploymentClient.Deployments))

	deployment := &appsv1.Deployment{}
	deployment.Name = "deploymentA"
	deployment.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	handler(deployment)
	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))
	got := c.DeploymentClient.Deployments[newDeploymentIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")]
	assert.Equal(t, "deploymentA", got.Name)
	assert.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", got.DeploymentUID)
}

func TestDeploymentAdd(t *testing.T) {
	c, _ := newTestClient(t)
	deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleDeploymentAdd)
}

// TestDeploymentCreate tests that a new deployment, created after otel-collector starts, has its attributes set
// correctly
func TestDeploymentCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Equal(t, 0, len(c.DeploymentClient.Deployments))

	// deployment is created in Pending phase. At this point it has a UID but no start time
	deployment := &appsv1.Deployment{}
	deployment.Name = "deployment1"
	deployment.UID = "11111111-2222-3333-4444-555555555555"
	c.DeploymentClient.handleDeploymentAdd(deployment)
	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))
	got := c.DeploymentClient.Deployments[newDeploymentIdentifier("resource_attribute", "k8s.deployment.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "deployment1", got.Name)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.DeploymentUID)

	startTime := meta_v1.NewTime(time.Now())
	deployment.CreationTimestamp = startTime
	c.DeploymentClient.handleDeploymentUpdate(&appsv1.Deployment{}, deployment)
	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))
	got = c.DeploymentClient.Deployments[newDeploymentIdentifier("resource_attribute", "k8s.deployment.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "deployment1", got.Name)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.DeploymentUID)
}

func TestDeploymentUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	deploymentAddAndUpdateTest(t, c, func(obj interface{}) {
		// first argument (old deployment) is not used right now
		c.DeploymentClient.handleDeploymentUpdate(&appsv1.Deployment{}, obj)
	})
}

func TestDeploymentDelete(t *testing.T) {
	c, _ := newTestClient(t)
	deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleDeploymentAdd)
	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))

	deployment := &appsv1.Deployment{}
	deployment.Name = "deploymentA"
	deployment.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	c.DeploymentClient.handleDeploymentDelete(deployment)

	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))
	assert.Equal(t, 1, len(c.DeploymentClient.deleteQueue))
	deleteRequest := c.DeploymentClient.deleteQueue[0]
	assert.Equal(t, newDeploymentIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"), deleteRequest.id)
	assert.Equal(t, "deploymentA", deleteRequest.deploymentName)
	assert.False(t, deleteRequest.ts.After(time.Now()))
}

func TestDeploymentDeleteLoop(t *testing.T) {
	// go c.deleteLoop(time.Second * 1)
	c, _ := newTestClient(t)

	deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleDeploymentAdd)
	assert.Equal(t, 1, len(c.DeploymentClient.Deployments))

	deployment := &appsv1.Deployment{}
	deployment.Name = "deploymentA"
	deployment.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	c.DeploymentClient.handleDeploymentDelete(deployment)

	gracePeriod := time.Millisecond * 500
	go c.DeploymentClient.deleteLoop(time.Millisecond, gracePeriod)
	go func() {
		time.Sleep(time.Millisecond * 50)
		c.m.Lock()
		assert.Equal(t, 1, len(c.DeploymentClient.Deployments))
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Equal(t, 1, len(c.DeploymentClient.deleteQueue))
		c.deleteMut.Unlock()

		time.Sleep(gracePeriod + (time.Millisecond * 50))
		c.m.Lock()
		assert.Equal(t, 0, len(c.DeploymentClient.Deployments))
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Equal(t, 0, len(c.DeploymentClient.deleteQueue))
		c.deleteMut.Unlock()
		close(c.stopCh)
	}()
	<-c.stopCh
}

func TestDeploymentExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{}, Filters{})

	deployment := &appsv1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:              "auth-service-deployment",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: meta_v1.Now(),
			Labels: map[string]string{
				"label1": "lv1",
			},
			Annotations: map[string]string{
				"annotation1": "av1",
			},
		},
	}

	testCases := []struct {
		name       string
		rules      ExtractionRulesDeployment
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesDeployment{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesDeployment{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromDeployment,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromDeployment,
			},
			},
		},
		attributes: map[string]string{
			"l1": "lv1",
			"a1": "av1",
		},
	},
		{
			name: "all-labels",
			rules: ExtractionRulesDeployment{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromDeployment,
				},
				},
			},
			attributes: map[string]string{
				"k8s.deployment.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesDeployment{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromDeployment,
				},
				},
			},
			attributes: map[string]string{
				"k8s.deployment.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.DeploymentClient.Rules = tc.rules
			c.DeploymentClient.handleDeploymentAdd(deployment)
			p, ok := c.GetDeployment(newDeploymentIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.Attributes))
			for k, v := range tc.attributes {
				got, ok := p.Attributes[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

func TestDeploymentIgnorePatterns(t *testing.T) {
	testCases := []struct {
		ignore     bool
		deployment appsv1.Deployment
	}{{
		ignore:     false,
		deployment: appsv1.Deployment{},
	}, {
		ignore: true,
		deployment: appsv1.Deployment{
			ObjectMeta: meta_v1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "True ",
				},
			},
		},
	}, {
		ignore: true,
		deployment: appsv1.Deployment{
			ObjectMeta: meta_v1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "true",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: meta_v1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "false",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: meta_v1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "test-deployment-name",
			},
		},
	},
	}

	c, _ := newTestClient(t)
	for _, tc := range testCases {
		assert.Equal(t, tc.ignore, c.DeploymentClient.shouldIgnoreDeployment(&tc.deployment))
	}
}

func TestExtractDeploymentLabelsAnnotations(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{}, Filters{})
	testCases := []struct {
		name                    string
		shouldExtractDeployment bool
		rules                   ExtractionRulesDeployment
	}{{
		name:                    "empty-rules",
		shouldExtractDeployment: false,
		rules:                   ExtractionRulesDeployment{},
	}, {
		name:                    "pod-rules",
		shouldExtractDeployment: false,
		rules: ExtractionRulesDeployment{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromPod,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromPod,
			},
			},
		},
	}, {
		name:                    "deployment-rules-only-annotations",
		shouldExtractDeployment: true,
		rules: ExtractionRulesDeployment{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromDeployment,
			},
			},
		},
	}, {
		name:                    "deployment-rules-only-labels",
		shouldExtractDeployment: true,
		rules: ExtractionRulesDeployment{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromDeployment,
			},
			},
		},
	},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.DeploymentClient.Rules = tc.rules
			assert.Equal(t, tc.shouldExtractDeployment, c.DeploymentClient.extractDeploymentLabelsAnnotations())
		})
	}
}
