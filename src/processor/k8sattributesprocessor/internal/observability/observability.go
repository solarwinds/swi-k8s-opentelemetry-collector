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

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib
// Changes customizing the original source code: see CHANGELOG.md in deploy/helm directory

package observability // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/observability"

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// TODO: re-think if processor should register it's own telemetry views or if some other
// mechanism should be used by the collector to discover views from all components

func init() {
	_ = view.Register(
		viewPodsUpdated,
		viewPodsAdded,
		viewPodsDeleted,
		viewIPLookupMiss,
		viewPodTableSize,
		viewNamespacesAdded,
		viewNamespacesUpdated,
		viewNamespacesDeleted,
		viewDeploymentsUpdated,
		viewDeploymentsAdded,
		viewDeploymentsDeleted,
		viewDeploymentTableSize,
		viewStatefulSetsUpdated,
		viewStatefulSetsAdded,
		viewStatefulSetsDeleted,
		viewStatefulSetTableSize,

		viewReplicaSetsUpdated,
		viewReplicaSetsAdded,
		viewReplicaSetsDeleted,
		viewReplicaSetTableSize,

		viewDaemonSetsUpdated,
		viewDaemonSetsAdded,
		viewDaemonSetsDeleted,
		viewDaemonSetTableSize,

		viewJobsUpdated,
		viewJobsAdded,
		viewJobsDeleted,
		viewJobTableSize,

		viewCronJobsUpdated,
		viewCronJobsAdded,
		viewCronJobsDeleted,
		viewCronJobTableSize,

		viewNodesUpdated,
		viewNodesAdded,
		viewNodesDeleted,
		viewNodeTableSize,

		viewPersistentVolumesUpdated,
		viewPersistentVolumesAdded,
		viewPersistentVolumesDeleted,
		viewPersistentVolumeTableSize,

		viewPersistentVolumeClaimsUpdated,
		viewPersistentVolumeClaimsAdded,
		viewPersistentVolumeClaimsDeleted,
		viewPersistentVolumeClaimTableSize,
	)
}

var (
	mPodsUpdated       = stats.Int64("otelsvc/k8s/pod_updated", "Number of pod update events received", "1")
	mPodsAdded         = stats.Int64("otelsvc/k8s/pod_added", "Number of pod add events received", "1")
	mPodsDeleted       = stats.Int64("otelsvc/k8s/pod_deleted", "Number of pod delete events received", "1")
	mPodTableSize      = stats.Int64("otelsvc/k8s/pod_table_size", "Size of table containing pod info", "1")
	mIPLookupMiss      = stats.Int64("otelsvc/k8s/ip_lookup_miss", "Number of times pod by IP lookup failed.", "1")
	mNamespacesUpdated = stats.Int64("otelsvc/k8s/namespace_updated", "Number of namespace update events received", "1")
	mNamespacesAdded   = stats.Int64("otelsvc/k8s/namespace_added", "Number of namespace add events received", "1")
	mNamespacesDeleted = stats.Int64("otelsvc/k8s/namespace_deleted", "Number of namespace delete events received", "1")
)

var viewPodsUpdated = &view.View{
	Name:        mPodsUpdated.Name(),
	Description: mPodsUpdated.Description(),
	Measure:     mPodsUpdated,
	Aggregation: view.Sum(),
}

var viewPodsAdded = &view.View{
	Name:        mPodsAdded.Name(),
	Description: mPodsAdded.Description(),
	Measure:     mPodsAdded,
	Aggregation: view.Sum(),
}

var viewPodsDeleted = &view.View{
	Name:        mPodsDeleted.Name(),
	Description: mPodsDeleted.Description(),
	Measure:     mPodsDeleted,
	Aggregation: view.Sum(),
}

var viewIPLookupMiss = &view.View{
	Name:        mIPLookupMiss.Name(),
	Description: mIPLookupMiss.Description(),
	Measure:     mIPLookupMiss,
	Aggregation: view.Sum(),
}

var viewPodTableSize = &view.View{
	Name:        mPodTableSize.Name(),
	Description: mPodTableSize.Description(),
	Measure:     mPodTableSize,
	Aggregation: view.LastValue(),
}

var viewNamespacesUpdated = &view.View{
	Name:        mNamespacesUpdated.Name(),
	Description: mNamespacesUpdated.Description(),
	Measure:     mNamespacesUpdated,
	Aggregation: view.Sum(),
}

var viewNamespacesAdded = &view.View{
	Name:        mNamespacesAdded.Name(),
	Description: mNamespacesAdded.Description(),
	Measure:     mNamespacesAdded,
	Aggregation: view.Sum(),
}

var viewNamespacesDeleted = &view.View{
	Name:        mNamespacesDeleted.Name(),
	Description: mNamespacesDeleted.Description(),
	Measure:     mNamespacesDeleted,
	Aggregation: view.Sum(),
}

// RecordPodUpdated increments the metric that records pod update events received.
func RecordPodUpdated() {
	stats.Record(context.Background(), mPodsUpdated.M(int64(1)))
}

// RecordPodAdded increments the metric that records pod add events receiver.
func RecordPodAdded() {
	stats.Record(context.Background(), mPodsAdded.M(int64(1)))
}

// RecordPodDeleted increments the metric that records pod events deleted.
func RecordPodDeleted() {
	stats.Record(context.Background(), mPodsDeleted.M(int64(1)))
}

// RecordIPLookupMiss increments the metric that records Pod lookup by IP misses.
func RecordIPLookupMiss() {
	stats.Record(context.Background(), mIPLookupMiss.M(int64(1)))
}

// RecordPodTableSize store size of pod table field in WatchClient
func RecordPodTableSize(podTableSize int64) {
	stats.Record(context.Background(), mPodTableSize.M(podTableSize))
}

// RecordNamespaceUpdated increments the metric that records namespace update events received.
func RecordNamespaceUpdated() {
	stats.Record(context.Background(), mNamespacesUpdated.M(int64(1)))
}

// RecordNamespaceAdded increments the metric that records namespace add events receiver.
func RecordNamespaceAdded() {
	stats.Record(context.Background(), mNamespacesAdded.M(int64(1)))
}

// RecordNamespaceDeleted increments the metric that records namespace events deleted.
func RecordNamespaceDeleted() {
	stats.Record(context.Background(), mNamespacesDeleted.M(int64(1)))
}
