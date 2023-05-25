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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

func Test_fakeResourceInformer(t *testing.T) {
	// nothing real to test here. just to make coverage happy
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	i := NewFakeResourceInformer(c, "ns", nil, nil)
	_, err = i.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{}, time.Second)
	assert.NoError(t, err)
	i.HasSynced()
	i.LastSyncResourceVersion()
	store := i.GetStore()
	assert.NoError(t, store.Add(appsv1.Deployment{}))
}

func Test_newDeploymentSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newDeploymentSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_deploymentInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := deploymentInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_deploymentInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := deploymentInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newStatefulSetSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newStatefulSetSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_StatefulSetInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := statefulSetInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_StatefulSetInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := statefulSetInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newReplicaSetSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newReplicaSetSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_ReplicaSetInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := replicaSetInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_ReplicaSetInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := replicaSetInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newDaemonSetSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newDaemonSetSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_DaemonSetInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := daemonSetInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_DaemonSetInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := daemonSetInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newJobSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newJobSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_JobInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := jobInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_JobInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := jobInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newCronJobSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newCronJobSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_CronJobInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := cronJobInformerListFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_CronJobInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := cronJobInformerWatchFuncWithSelectors(c, "test-ns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newNodeSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newNodeSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_NodeInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := nodeInformerListFuncWithSelectors(c, ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_NodeInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := nodeInformerWatchFuncWithSelectors(c, ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newPersistentVolumeSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newPersistentVolumeSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_PersistentVolumeInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := persistentVolumeInformerListFuncWithSelectors(c, ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_PersistentVolumeInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := persistentVolumeInformerWatchFuncWithSelectors(c, ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_newPersistentVolumeClaimSharedInformer(t *testing.T) {
	labelSelector, fieldSelector, err := selectorsFromFilters(Filters{})
	require.NoError(t, err)
	client, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)
	informer := newPersistentVolumeClaimSharedInformer(client, "testns", labelSelector, fieldSelector)
	assert.NotNil(t, informer)
}

func Test_PersistentVolumeClaimInformerListFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	listFunc := persistentVolumeClaimInformerListFuncWithSelectors(c, "testns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := listFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func Test_PersistentVolumeClaimInformerWatchFuncWithSelectors(t *testing.T) {
	ls, fs, err := newTestSelectors()
	assert.NoError(t, err)
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	assert.NoError(t, err)
	watchFunc := persistentVolumeClaimInformerWatchFuncWithSelectors(c, "testns", ls, fs)
	opts := metav1.ListOptions{}
	obj, err := watchFunc(opts)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
}

func newTestSelectors() (labels.Selector, fields.Selector, error) {
	ls, fs, err := selectorsFromFilters(Filters{
		Fields: []FieldFilter{
			{
				Key:   "kk1",
				Value: "kv1",
				Op:    selection.Equals,
			},
		},
		Labels: []FieldFilter{
			{
				Key:   "lk1",
				Value: "lv1",
				Op:    selection.NotEquals,
			},
		},
	})
	return ls, fs, err
}
