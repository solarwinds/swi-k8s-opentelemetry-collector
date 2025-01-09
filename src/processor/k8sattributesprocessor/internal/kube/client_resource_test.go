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

package kube

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

func deploymentAddAndUpdateTest(t *testing.T, c *WatchClient, handler func(obj any)) {
	assert.Empty(t, c.DeploymentClient.Resources)

	deployment := &appsv1.Deployment{}
	deployment.Name = "deploymentA"
	deployment.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	handler(deployment)
	assert.Len(t, c.DeploymentClient.Resources, 1)
	got := c.DeploymentClient.Resources[newResourceIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")]
	assert.Equal(t, "deploymentA", got.GetName())
	assert.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", got.GetUID())
}

func TestDeploymentAdd(t *testing.T) {
	c, _ := newTestClient(t)
	deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleResourceAdd)
}

// TestDeploymentCreate tests that a new deployment, created after otel-collector starts, has its attributes set
// correctly
func TestDeploymentCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.DeploymentClient.Resources)

	// deployment is created in Pending phase. At this point it has a UID but no start time
	deployment := &appsv1.Deployment{}
	deployment.Name = "deployment1"
	deployment.UID = "11111111-2222-3333-4444-555555555555"
	c.DeploymentClient.handleResourceAdd(deployment)
	assert.Len(t, c.DeploymentClient.Resources, 1)
	got := c.DeploymentClient.Resources[newResourceIdentifier("resource_attribute", "k8s.deployment.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "deployment1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	deployment.CreationTimestamp = startTime
	c.DeploymentClient.handleResourceUpdate(&appsv1.Deployment{}, deployment)
	assert.Len(t, c.DeploymentClient.Resources, 1)
	got = c.DeploymentClient.Resources[newResourceIdentifier("resource_attribute", "k8s.deployment.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "deployment1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestDeploymentUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	deploymentAddAndUpdateTest(t, c, func(obj any) {
		// first argument (old deployment) is not used right now
		c.DeploymentClient.handleResourceUpdate(&appsv1.Deployment{}, obj)
	})
}

func TestDeploymentDelete(t *testing.T) {
	tests := []struct {
		name        string
		objToDelete any
	}{
		{
			name: "Deployment should be deleted",
			objToDelete: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deploymentA",
					UID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				},
			},
		},
		{
			name: "Deployment wrapped in DeletedFinalStateUnknown should be deleted",
			objToDelete: cache.DeletedFinalStateUnknown{
				Obj: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deploymentA",
						UID:  "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestClient(t)
			deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleResourceAdd)
			assert.Len(t, c.DeploymentClient.Resources, 1)

			c.DeploymentClient.handleResourceDelete(tt.objToDelete)

			assert.Len(t, c.DeploymentClient.Resources, 1)
			assert.Len(t, c.DeploymentClient.deleteQueue, 1)
			deleteRequest := c.DeploymentClient.deleteQueue[0]
			assert.Equal(t, newResourceIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"), deleteRequest.id)
			assert.Equal(t, "deploymentA", deleteRequest.resourceName)
			assert.False(t, deleteRequest.ts.After(time.Now()))
		})
	}
}

func TestDeploymentExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth-service-deployment",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
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
			rules: ExtractionRulesResource{
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
			rules: ExtractionRulesResource{
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
			c.DeploymentClient.handleResourceAdd(deployment)
			p, ok := c.GetResource(MetadataFromDeployment, newResourceIdentifier("resource_attribute", "k8s.deployment.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
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
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "True ",
				},
			},
		},
	}, {
		ignore: true,
		deployment: appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "true",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "false",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"opentelemetry.io/k8s-processor/ignore": "",
				},
			},
		},
	}, {
		ignore: false,
		deployment: appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-deployment-name",
			},
		},
	},
	}

	c, _ := newTestClient(t)
	for _, tc := range testCases {
		assert.Equal(t, tc.ignore, c.DeploymentClient.shouldIgnoreResource(&tc.deployment))
	}
}

// StatefulSet tests
func TestStatefulSetAdd(t *testing.T) {
	c, _ := newTestClient(t)
	statefulSet := &appsv1.StatefulSet{}
	resourceAddAndUpdateTest(t, MetadataFromStatefulSet, statefulSet, c.StatefulSetClient, c.StatefulSetClient.handleResourceAdd)
}

// TestStatefulSetCreate tests that a new statefulset, created after otel-collector starts, has its attributes set
// correctly
func TestStatefulSetCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.StatefulSetClient.Resources)

	// statefulset is created in Pending phase. At this point it has a UID but no start time
	statefulSet := &appsv1.StatefulSet{}
	statefulSet.Name = "statefulSet1"
	statefulSet.UID = "11111111-2222-3333-4444-555555555555"
	c.StatefulSetClient.handleResourceAdd(statefulSet)
	assert.Len(t, c.StatefulSetClient.Resources, 1)
	got := c.StatefulSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.statefulset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "statefulSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	statefulSet.CreationTimestamp = startTime
	c.StatefulSetClient.handleResourceUpdate(&appsv1.StatefulSet{}, statefulSet)
	assert.Len(t, c.StatefulSetClient.Resources, 1)
	got = c.StatefulSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.statefulset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "statefulSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestStatefulSetUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	statefulSet := &appsv1.StatefulSet{}
	resourceAddAndUpdateTest(t, MetadataFromStatefulSet, statefulSet, c.StatefulSetClient, func(obj any) {
		c.StatefulSetClient.handleResourceUpdate(&appsv1.StatefulSet{}, obj)
	})
}

func TestStatefulSetExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth-service-statefulset",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromStatefulSet,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromStatefulSet,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromStatefulSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.statefulset.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromStatefulSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.statefulset.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.StatefulSetClient.Rules = tc.rules
			c.StatefulSetClient.handleResourceAdd(statefulSet)
			p, ok := c.GetResource(MetadataFromStatefulSet, newResourceIdentifier("resource_attribute", "k8s.statefulset.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// ReplicaSet tests
func TestReplicaSetAdd(t *testing.T) {
	c, _ := newTestClient(t)
	replicaSet := &appsv1.ReplicaSet{}
	resourceAddAndUpdateTest(t, MetadataFromReplicaSet, replicaSet, c.ReplicaSetClient, c.ReplicaSetClient.handleResourceAdd)
}

func TestReplicaSetCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.ReplicaSetClient.Resources)

	// ReplicaSet is created in Pending phase. At this point it has a UID but no start time
	replicaSet := &appsv1.ReplicaSet{}
	replicaSet.Name = "replicaSet1"
	replicaSet.UID = "11111111-2222-3333-4444-555555555555"
	c.ReplicaSetClient.handleResourceAdd(replicaSet)
	assert.Len(t, c.ReplicaSetClient.Resources, 1)
	got := c.ReplicaSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.replicaset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "replicaSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	replicaSet.CreationTimestamp = startTime
	c.ReplicaSetClient.handleResourceUpdate(&appsv1.ReplicaSet{}, replicaSet)
	assert.Len(t, c.ReplicaSetClient.Resources, 1)
	got = c.ReplicaSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.replicaset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "replicaSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestReplicaSetUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	replicaSet := &appsv1.ReplicaSet{}
	resourceAddAndUpdateTest(t, MetadataFromReplicaSet, replicaSet, c.ReplicaSetClient, func(obj any) {
		c.ReplicaSetClient.handleResourceUpdate(&appsv1.ReplicaSet{}, obj)
	})
}

func TestReplicaSetExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	replicaSet := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth-service-replicaset",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromReplicaSet,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromReplicaSet,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromReplicaSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.replicaset.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromReplicaSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.replicaset.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.ReplicaSetClient.Rules = tc.rules
			c.ReplicaSetClient.handleResourceAdd(replicaSet)
			p, ok := c.GetResource(MetadataFromReplicaSet, newResourceIdentifier("resource_attribute", "k8s.replicaset.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// DaemonSet tests
func TestDaemonSetAdd(t *testing.T) {
	c, _ := newTestClient(t)
	daemonSet := &appsv1.DaemonSet{}
	resourceAddAndUpdateTest(t, MetadataFromDaemonSet, daemonSet, c.DaemonSetClient, c.DaemonSetClient.handleResourceAdd)
}

// TestDaemonSetCreate tests that a new DaemonSet, created after otel-collector starts, has its attributes set
// correctly
func TestDaemonSetCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.DaemonSetClient.Resources)

	// DaemonSet is created in Pending phase. At this point it has a UID but no start time
	daemonSet := &appsv1.DaemonSet{}
	daemonSet.Name = "daemonSet1"
	daemonSet.UID = "11111111-2222-3333-4444-555555555555"
	c.DaemonSetClient.handleResourceAdd(daemonSet)
	assert.Len(t, c.DaemonSetClient.Resources, 1)
	got := c.DaemonSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.daemonset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "daemonSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	daemonSet.CreationTimestamp = startTime
	c.DaemonSetClient.handleResourceUpdate(&appsv1.DaemonSet{}, daemonSet)
	assert.Len(t, c.DaemonSetClient.Resources, 1)
	got = c.DaemonSetClient.Resources[newResourceIdentifier("resource_attribute", "k8s.daemonset.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "daemonSet1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestDaemonSetUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	daemonSet := &appsv1.DaemonSet{}
	resourceAddAndUpdateTest(t, MetadataFromDaemonSet, daemonSet, c.DaemonSetClient, func(obj any) {
		c.DaemonSetClient.handleResourceUpdate(&appsv1.DaemonSet{}, obj)
	})
}

func TestDaemonSetExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth-service-daemonset",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromDaemonSet,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromDaemonSet,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromDaemonSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.daemonset.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromDaemonSet,
				},
				},
			},
			attributes: map[string]string{
				"k8s.daemonset.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.DaemonSetClient.Rules = tc.rules
			c.DaemonSetClient.handleResourceAdd(daemonSet)
			p, ok := c.GetResource(MetadataFromDaemonSet, newResourceIdentifier("resource_attribute", "k8s.daemonset.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// Job tests
func TestJobAdd(t *testing.T) {
	c, _ := newTestClient(t)
	job := &batchv1.Job{}
	resourceAddAndUpdateTest(t, MetadataFromJob, job, c.JobClient, c.JobClient.handleResourceAdd)
}

// TestJobCreate tests that a new Job, created after otel-collector starts, has its attributes set
// correctly
func TestJobCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, (c.JobClient.Resources))

	// Job is created in Pending phase. At this point it has a UID but no start time
	job := &batchv1.Job{}
	job.Name = "job1"
	job.UID = "11111111-2222-3333-4444-555555555555"
	c.JobClient.handleResourceAdd(job)
	assert.Len(t, c.JobClient.Resources, 1)
	got := c.JobClient.Resources[newResourceIdentifier("resource_attribute", "k8s.job.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "job1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	job.CreationTimestamp = startTime
	c.JobClient.handleResourceUpdate(&batchv1.Job{}, job)
	assert.Len(t, c.JobClient.Resources, 1)
	got = c.JobClient.Resources[newResourceIdentifier("resource_attribute", "k8s.job.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "job1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestJobUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	job := &batchv1.Job{}
	resourceAddAndUpdateTest(t, MetadataFromJob, job, c.JobClient, func(obj any) {
		c.JobClient.handleResourceUpdate(&batchv1.Job{}, obj)
	})
}

func TestJobExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "job",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromJob,
			},
			},
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromJob,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromJob,
				},
				},
			},
			attributes: map[string]string{
				"k8s.job.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromJob,
				},
				},
			},
			attributes: map[string]string{
				"k8s.job.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.JobClient.Rules = tc.rules
			c.JobClient.handleResourceAdd(job)
			p, ok := c.GetResource(MetadataFromJob, newResourceIdentifier("resource_attribute", "k8s.job.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)
			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// CronJob tests
func TestCronJobAdd(t *testing.T) {
	c, _ := newTestClient(t)
	cronJob := &batchv1.CronJob{}
	resourceAddAndUpdateTest(t, MetadataFromCronJob, cronJob, c.CronJobClient, c.CronJobClient.handleResourceAdd)
}

// TestCronJobCreate tests that a new CronJob, created after otel-collector starts, has its attributes set
// correctly
func TestCronJobCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.CronJobClient.Resources)

	// CronJob is created in Pending phase. At this point it has a UID but no start time
	cronJob := &batchv1.CronJob{}
	cronJob.Name = "cronJob1"
	cronJob.UID = "11111111-2222-3333-4444-555555555555"
	c.CronJobClient.handleResourceAdd(cronJob)
	assert.Len(t, c.CronJobClient.Resources, 1)
	got := c.CronJobClient.Resources[newResourceIdentifier("resource_attribute", "k8s.cronjob.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "cronJob1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	cronJob.CreationTimestamp = startTime
	c.CronJobClient.handleResourceUpdate(&batchv1.CronJob{}, cronJob)
	assert.Len(t, c.CronJobClient.Resources, 1)
	got = c.CronJobClient.Resources[newResourceIdentifier("resource_attribute", "k8s.cronjob.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "cronJob1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestCronJobUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	cronJob := &batchv1.CronJob{}
	resourceAddAndUpdateTest(t, MetadataFromCronJob, cronJob, c.CronJobClient, func(obj any) {
		c.CronJobClient.handleResourceUpdate(&batchv1.CronJob{}, obj)
	})
}

func TestCronJobExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cronjob",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromCronJob,
			},
			},
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromCronJob,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromCronJob,
				},
				},
			},
			attributes: map[string]string{
				"k8s.cronjob.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromCronJob,
				},
				},
			},
			attributes: map[string]string{
				"k8s.cronjob.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.CronJobClient.Rules = tc.rules
			c.CronJobClient.handleResourceAdd(cronJob)
			p, ok := c.GetResource(MetadataFromCronJob, newResourceIdentifier("resource_attribute", "k8s.cronjob.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// Node tests
func TestNodeAdd(t *testing.T) {
	c, _ := newTestClient(t)
	node := &corev1.Node{}
	resourceAddAndUpdateTest(t, MetadataFromNode, node, c.NodeClient, c.NodeClient.handleResourceAdd)
}

// TestNodeCreate tests that a new Node, created after otel-collector starts, has its attributes set
// correctly
func TestNodeCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.NodeClient.Resources)

	// Node is created in Pending phase. At this point it has a UID but no start time
	node := &corev1.Node{}
	node.Name = "node1"
	node.UID = "11111111-2222-3333-4444-555555555555"
	c.NodeClient.handleResourceAdd(node)
	assert.Len(t, c.NodeClient.Resources, 1)
	got := c.NodeClient.Resources[newResourceIdentifier("resource_attribute", "k8s.node.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "node1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	node.CreationTimestamp = startTime
	c.NodeClient.handleResourceUpdate(&corev1.Node{}, node)
	assert.Len(t, c.NodeClient.Resources, 1)
	got = c.NodeClient.Resources[newResourceIdentifier("resource_attribute", "k8s.node.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "node1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestNodeUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	node := &corev1.Node{}
	resourceAddAndUpdateTest(t, MetadataFromNode, node, c.NodeClient, func(obj any) {
		c.NodeClient.handleResourceUpdate(&corev1.Node{}, obj)
	})
}

func TestNodeExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node1",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromNode,
			},
			},
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromNode,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromNode,
				},
				},
			},
			attributes: map[string]string{
				"k8s.node.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromNode,
				},
				},
			},
			attributes: map[string]string{
				"k8s.node.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.NodeClient.Rules = tc.rules
			c.NodeClient.handleResourceAdd(node)
			p, ok := c.GetResource(MetadataFromNode, newResourceIdentifier("resource_attribute", "k8s.node.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// Persistent volume tests
func TestPersistentVolumeAdd(t *testing.T) {
	c, _ := newTestClient(t)
	persistentVolume := &corev1.PersistentVolume{}
	resourceAddAndUpdateTest(t, MetadataFromPersistentVolume, persistentVolume, c.PersistentVolumeClient, c.PersistentVolumeClient.handleResourceAdd)
}

// TestPersistentVolumeCreate tests that a new PersistentVolume
func TestPersistentVolumeCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.PersistentVolumeClient.Resources)

	// PersistentVolume is created in Pending phase. At this point it has a UID but no start time
	persistentVolume := &corev1.PersistentVolume{}
	persistentVolume.Name = "persistentVolume1"
	persistentVolume.UID = "11111111-2222-3333-4444-555555555555"
	c.PersistentVolumeClient.handleResourceAdd(persistentVolume)
	assert.Len(t, c.PersistentVolumeClient.Resources, 1)
	got := c.PersistentVolumeClient.Resources[newResourceIdentifier("resource_attribute", "k8s.persistentvolume.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "persistentVolume1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	persistentVolume.CreationTimestamp = startTime
	c.PersistentVolumeClient.handleResourceUpdate(&corev1.PersistentVolume{}, persistentVolume)
	assert.Len(t, c.PersistentVolumeClient.Resources, 1)
	got = c.PersistentVolumeClient.Resources[newResourceIdentifier("resource_attribute", "k8s.persistentvolume.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "persistentVolume1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestPersistentVolumeUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	persistentVolume := &corev1.PersistentVolume{}
	resourceAddAndUpdateTest(t, MetadataFromPersistentVolume, persistentVolume, c.PersistentVolumeClient, func(obj any) {
		c.PersistentVolumeClient.handleResourceUpdate(&corev1.PersistentVolume{}, obj)
	})
}

func TestPersistentVolumeExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	persistentVolume := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "persistentVolume1",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromPersistentVolume,
			},
			},
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromPersistentVolume,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromPersistentVolume,
				},
				},
			},
			attributes: map[string]string{
				"k8s.persistentvolume.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromPersistentVolume,
				},
				},
			},
			attributes: map[string]string{
				"k8s.persistentvolume.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.PersistentVolumeClient.Rules = tc.rules
			c.PersistentVolumeClient.handleResourceAdd(persistentVolume)
			p, ok := c.GetResource(MetadataFromPersistentVolume, newResourceIdentifier("resource_attribute", "k8s.persistentvolume.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// Persistent volume claim tests
func TestPersistentVolumeClaimAdd(t *testing.T) {
	c, _ := newTestClient(t)
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	resourceAddAndUpdateTest(t, MetadataFromPersistentVolumeClaim, persistentVolumeClaim, c.PersistentVolumeClaimClient, c.PersistentVolumeClaimClient.handleResourceAdd)
}

// TestPersistentVolumeClaimCreate tests that a new PersistentVolumeClaim
func TestPersistentVolumeClaimCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.PersistentVolumeClaimClient.Resources)

	// PersistentVolumeClaim is created in Pending phase. At this point it has a UID but no start time
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	persistentVolumeClaim.Name = "persistentVolumeClaim1"
	persistentVolumeClaim.UID = "11111111-2222-3333-4444-555555555555"
	c.PersistentVolumeClaimClient.handleResourceAdd(persistentVolumeClaim)
	assert.Len(t, c.PersistentVolumeClaimClient.Resources, 1)
	got := c.PersistentVolumeClaimClient.Resources[newResourceIdentifier("resource_attribute", "k8s.persistentvolumeclaim.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "persistentVolumeClaim1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	persistentVolumeClaim.CreationTimestamp = startTime
	c.PersistentVolumeClaimClient.handleResourceUpdate(&corev1.PersistentVolumeClaim{}, persistentVolumeClaim)
	assert.Len(t, c.PersistentVolumeClaimClient.Resources, 1)
	got = c.PersistentVolumeClaimClient.Resources[newResourceIdentifier("resource_attribute", "k8s.persistentvolumeclaim.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "persistentVolumeClaim1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestPersistentVolumeClaimUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	persistentVolumeClaim := &corev1.PersistentVolumeClaim{}
	resourceAddAndUpdateTest(t, MetadataFromPersistentVolumeClaim, persistentVolumeClaim, c.PersistentVolumeClaimClient, func(obj any) {
		c.PersistentVolumeClaimClient.handleResourceUpdate(&corev1.PersistentVolumeClaim{}, obj)
	})
}

func TestPersistentVolumeClaimExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	persistentVolumeClaim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "persistentVolumeClaim1",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromPersistentVolumeClaim,
			},
			},
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromPersistentVolumeClaim,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromPersistentVolumeClaim,
				},
				},
			},
			attributes: map[string]string{
				"k8s.persistentvolumeclaim.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromPersistentVolumeClaim,
				},
				},
			},
			attributes: map[string]string{
				"k8s.persistentvolumeclaim.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.PersistentVolumeClaimClient.Rules = tc.rules
			c.PersistentVolumeClaimClient.handleResourceAdd(persistentVolumeClaim)
			p, ok := c.GetResource(MetadataFromPersistentVolumeClaim, newResourceIdentifier("resource_attribute", "k8s.persistentvolumeclaim.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

// Service tests
func TestServiceAdd(t *testing.T) {
	c, _ := newTestClient(t)
	service := &corev1.Service{}
	resourceAddAndUpdateTest(t, MetadataFromService, service, c.ServiceClient, c.ServiceClient.handleResourceAdd)
}

// TestServiceCreate tests that a new Service
func TestServiceCreate(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Empty(t, c.ServiceClient.Resources)

	// Service is created in Pending phase. At this point it has a UID but no start time
	service := &corev1.Service{}
	service.Name = "service1"
	service.UID = "11111111-2222-3333-4444-555555555555"
	c.ServiceClient.handleResourceAdd(service)
	assert.Len(t, c.ServiceClient.Resources, 1)
	got := c.ServiceClient.Resources[newResourceIdentifier("resource_attribute", "k8s.service.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "service1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())

	startTime := metav1.NewTime(time.Now())
	service.CreationTimestamp = startTime
	c.ServiceClient.handleResourceUpdate(&corev1.Service{}, service)
	assert.Len(t, c.ServiceClient.Resources, 1)
	got = c.ServiceClient.Resources[newResourceIdentifier("resource_attribute", "k8s.service.uid", "11111111-2222-3333-4444-555555555555")]
	assert.Equal(t, "service1", got.GetName())
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", got.GetUID())
}

func TestServiceUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	service := &corev1.Service{}
	resourceAddAndUpdateTest(t, MetadataFromService, service, c.ServiceClient, func(obj any) {
		c.ServiceClient.handleResourceUpdate(&corev1.Service{}, obj)
	})
}

func TestServiceExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "service1",
			UID:               "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			CreationTimestamp: metav1.Now(),
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
		rules      ExtractionRulesResource
		attributes map[string]string
	}{{
		name:       "no-rules",
		rules:      ExtractionRulesResource{},
		attributes: nil,
	}, {
		name: "labels",
		rules: ExtractionRulesResource{
			Labels: []FieldExtractionRule{{
				Name: "l1",
				Key:  "label1",
				From: MetadataFromService,
			},
			},
			Annotations: []FieldExtractionRule{{
				Name: "a1",
				Key:  "annotation1",
				From: MetadataFromService,
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
			rules: ExtractionRulesResource{
				Labels: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:la.*)$"),
					From:     MetadataFromService,
				},
				},
			},
			attributes: map[string]string{
				"k8s.service.labels.label1": "lv1",
			},
		},
		{
			name: "all-annotations",
			rules: ExtractionRulesResource{
				Annotations: []FieldExtractionRule{{
					KeyRegex: regexp.MustCompile("^(?:an.*)$"),
					From:     MetadataFromService,
				},
				},
			},
			attributes: map[string]string{
				"k8s.service.annotations.annotation1": "av1",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c.ServiceClient.Rules = tc.rules
			c.ServiceClient.handleResourceAdd(service)
			p, ok := c.GetResource(MetadataFromService, newResourceIdentifier("resource_attribute", "k8s.service.uid", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(p.GetAttributes()))
			for k, v := range tc.attributes {
				got, ok := p.GetAttributes()[k]
				assert.True(t, ok)
				assert.Equal(t, v, got)
			}
		})
	}
}

func TestResourceDeleteLoop(t *testing.T) {
	c, _ := newTestClient(t)

	deploymentAddAndUpdateTest(t, c, c.DeploymentClient.handleResourceAdd)
	assert.Len(t, c.DeploymentClient.Resources, 1)

	deployment := &appsv1.Deployment{}
	deployment.Name = "deploymentA"
	deployment.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	c.DeploymentClient.handleResourceDelete(deployment)

	gracePeriod := time.Millisecond * 500
	go c.DeploymentClient.deleteLoop(time.Millisecond, gracePeriod)
	go func() {
		time.Sleep(time.Millisecond * 50)
		c.m.Lock()
		assert.Len(t, c.DeploymentClient.Resources, 1)
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Len(t, c.DeploymentClient.deleteQueue, 1)
		c.deleteMut.Unlock()

		time.Sleep(gracePeriod + (time.Millisecond * 50))
		c.m.Lock()
		assert.Empty(t, c.DeploymentClient.Resources)
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Empty(t, c.DeploymentClient.deleteQueue)
		c.deleteMut.Unlock()
		close(c.stopCh)
	}()
	<-c.stopCh
}

func TestExtractResourceLabelsAnnotations(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, Filters{})
	testCases := []struct {
		name                    string
		shouldExtractDeployment bool
		rules                   ExtractionRulesResource
	}{{
		name:                    "empty-rules",
		shouldExtractDeployment: false,
		rules:                   ExtractionRulesResource{},
	}, {
		name:                    "pod-rules",
		shouldExtractDeployment: false,
		rules: ExtractionRulesResource{
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
		rules: ExtractionRulesResource{
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
		rules: ExtractionRulesResource{
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
			assert.Equal(t, tc.shouldExtractDeployment, c.DeploymentClient.extractResourceLabelsAnnotations(MetadataFromDeployment))
		})
	}
}

// Utility function for add and update tests
func resourceAddAndUpdateTest(t *testing.T, resourceType string, resource metav1.Object, client *WatchResourceClient[KubernetesResource], handler func(obj any)) {
	assert.Empty(t, client.Resources)

	resourceUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	resource.SetName("resourceA")
	resource.SetUID(types.UID(resourceUID))
	handler(resource)
	assert.Len(t, client.Resources, 1)
	got := client.Resources[newResourceIdentifier("resource_attribute", "k8s."+resourceType+".uid", resourceUID)]
	assert.Equal(t, "resourceA", got.GetName())
	assert.Equal(t, resourceUID, string(got.GetUID()))
}

func newResourceIdentifier(from string, name string, value string) ResourceIdentifier {
	return ResourceIdentifier{
		{
			Source: AssociationSource{
				From: from,
				Name: name,
			},
			Value: value,
		},
	}
}
