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

package observability // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/observability"

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var (
	// Deployment metrics
	mDeploymentsUpdated  = stats.Int64("otelsvc/k8s/deployment_updated", "Number of deployment update events received", "1")
	mDeploymentsAdded    = stats.Int64("otelsvc/k8s/deployment_added", "Number of deployment add events received", "1")
	mDeploymentsDeleted  = stats.Int64("otelsvc/k8s/deployment_deleted", "Number of deployment delete events received", "1")
	mDeploymentTableSize = stats.Int64("otelsvc/k8s/deployment_table_size", "Size of table containing deployment info", "1")

	// StatefulSet metrics
	mStatefulSetsUpdated  = stats.Int64("otelsvc/k8s/statefulset_updated", "Number of statefulset update events received", "1")
	mStatefulSetsAdded    = stats.Int64("otelsvc/k8s/statefulset_added", "Number of statefulset add events received", "1")
	mStatefulSetsDeleted  = stats.Int64("otelsvc/k8s/statefulset_deleted", "Number of statefulset delete events received", "1")
	mStatefulSetTableSize = stats.Int64("otelsvc/k8s/statefulset_table_size", "Size of table containing statefulset info", "1")

	// ReplicaSet metrics
	mReplicaSetsUpdated  = stats.Int64("otelsvc/k8s/replicaset_updated", "Number of replicaset update events received", "1")
	mReplicaSetsAdded    = stats.Int64("otelsvc/k8s/replicaset_added", "Number of replicaset add events received", "1")
	mReplicaSetsDeleted  = stats.Int64("otelsvc/k8s/replicaset_deleted", "Number of replicaset delete events received", "1")
	mReplicaSetTableSize = stats.Int64("otelsvc/k8s/replicaset_table_size", "Size of table containing replicaset info", "1")

	// DaemonSet metrics
	mDaemonSetsUpdated  = stats.Int64("otelsvc/k8s/daemonset_updated", "Number of daemonset update events received", "1")
	mDaemonSetsAdded    = stats.Int64("otelsvc/k8s/daemonset_added", "Number of daemonset add events received", "1")
	mDaemonSetsDeleted  = stats.Int64("otelsvc/k8s/daemonset_deleted", "Number of daemonset delete events received", "1")
	mDaemonSetTableSize = stats.Int64("otelsvc/k8s/daemonset_table_size", "Size of table containing daemonset info", "1")

	// Job metrics
	mJobsUpdated  = stats.Int64("otelsvc/k8s/job_updated", "Number of job update events received", "1")
	mJobsAdded    = stats.Int64("otelsvc/k8s/job_added", "Number of job add events received", "1")
	mJobsDeleted  = stats.Int64("otelsvc/k8s/job_deleted", "Number of job delete events received", "1")
	mJobTableSize = stats.Int64("otelsvc/k8s/job_table_size", "Size of table containing job info", "1")

	// CronJob metrics
	mCronJobsUpdated  = stats.Int64("otelsvc/k8s/cronjob_updated", "Number of cronjob update events received", "1")
	mCronJobsAdded    = stats.Int64("otelsvc/k8s/cronjob_added", "Number of cronjob add events received", "1")
	mCronJobsDeleted  = stats.Int64("otelsvc/k8s/cronjob_deleted", "Number of cronjob delete events received", "1")
	mCronJobTableSize = stats.Int64("otelsvc/k8s/cronjob_table_size", "Size of table containing cronjob info", "1")

	// Node metrics
	mNodesUpdated  = stats.Int64("otelsvc/k8s/node_updated", "Number of node update events received", "1")
	mNodesAdded    = stats.Int64("otelsvc/k8s/node_added", "Number of node add events received", "1")
	mNodesDeleted  = stats.Int64("otelsvc/k8s/node_deleted", "Number of node delete events received", "1")
	mNodeTableSize = stats.Int64("otelsvc/k8s/node_table_size", "Size of table containing node info", "1")

	// Persistent volume metrics
	mPersistentVolumesUpdated  = stats.Int64("otelsvc/k8s/persistentvolume_updated", "Number of persistentvolume update events received", "1")
	mPersistentVolumesAdded    = stats.Int64("otelsvc/k8s/persistentvolume_added", "Number of persistentvolume add events received", "1")
	mPersistentVolumesDeleted  = stats.Int64("otelsvc/k8s/persistentvolume_deleted", "Number of persistentvolume delete events received", "1")
	mPersistentVolumeTableSize = stats.Int64("otelsvc/k8s/persistentvolume_table_size", "Size of table containing persistentvolume info", "1")

	// Persistent volume claim metrics
	mPersistentVolumeClaimsUpdated  = stats.Int64("otelsvc/k8s/persistentvolumeclaim_updated", "Number of persistentvolumeclaim update events received", "1")
	mPersistentVolumeClaimsAdded    = stats.Int64("otelsvc/k8s/persistentvolumeclaim_added", "Number of persistentvolumeclaim add events received", "1")
	mPersistentVolumeClaimsDeleted  = stats.Int64("otelsvc/k8s/persistentvolumeclaim_deleted", "Number of persistentvolumeclaim delete events received", "1")
	mPersistentVolumeClaimTableSize = stats.Int64("otelsvc/k8s/persistentvolumeclaim_table_size", "Size of table containing persistentvolumeclaim info", "1")

	// Service metrics
	mServicesUpdated  = stats.Int64("otelsvc/k8s/service_updated", "Number of service update events received", "1")
	mServicesAdded    = stats.Int64("otelsvc/k8s/service_added", "Number of service add events received", "1")
	mServicesDeleted  = stats.Int64("otelsvc/k8s/service_deleted", "Number of service delete events received", "1")
	mServiceTableSize = stats.Int64("otelsvc/k8s/service_table_size", "Size of table containing service info", "1")
)

// RecordDeploymentUpdated increments the metric that records deployment update events received.
func RecordDeploymentUpdated() {
	stats.Record(context.Background(), mDeploymentsUpdated.M(int64(1)))
}

// RecordDeploymentAdded increments the metric that records deployment add events receiver.
func RecordDeploymentAdded() {
	stats.Record(context.Background(), mDeploymentsAdded.M(int64(1)))
}

// RecordDeploymentDeleted increments the metric that records deployment events deleted.
func RecordDeploymentDeleted() {
	stats.Record(context.Background(), mDeploymentsDeleted.M(int64(1)))
}

// RecordDeploymentTableSize stores the size of the deployment table field in WatchClient
func RecordDeploymentTableSize(deploymentTableSize int64) {
	stats.Record(context.Background(), mDeploymentTableSize.M(deploymentTableSize))
}

// RecordStatefulSetUpdated increments the metric that records statefulset update events received.
func RecordStatefulSetUpdated() {
	stats.Record(context.Background(), mStatefulSetsUpdated.M(int64(1)))
}

// RecordStatefulSetAdded increments the metric that records statefulset add events receiver.
func RecordStatefulSetAdded() {
	stats.Record(context.Background(), mStatefulSetsAdded.M(int64(1)))
}

// RecordStatefulSetDeleted increments the metric that records statefulset events deleted.
func RecordStatefulSetDeleted() {
	stats.Record(context.Background(), mStatefulSetsDeleted.M(int64(1)))
}

// RecordStatefulSetTableSize stores the size of the statefulset table field in WatchClient
func RecordStatefulSetTableSize(statefulSetTableSize int64) {
	stats.Record(context.Background(), mStatefulSetTableSize.M(statefulSetTableSize))
}

// RecordReplicaSetUpdated increments the metric that records replicaset update events received.
func RecordReplicaSetUpdated() {
	stats.Record(context.Background(), mReplicaSetsUpdated.M(int64(1)))
}

// RecordReplicaSetAdded increments the metric that records replicaset add events received.
func RecordReplicaSetAdded() {
	stats.Record(context.Background(), mReplicaSetsAdded.M(int64(1)))
}

// RecordReplicaSetDeleted increments the metric that records replicaset delete events received.
func RecordReplicaSetDeleted() {
	stats.Record(context.Background(), mReplicaSetsDeleted.M(int64(1)))
}

// RecordReplicaSetTableSize stores the size of the replicaset table field in WatchClient
func RecordReplicaSetTableSize(replicaSetTableSize int64) {
	stats.Record(context.Background(), mReplicaSetTableSize.M(replicaSetTableSize))
}

// RecordDaemonSetUpdated increments the metric that records daemonset update events received.
func RecordDaemonSetUpdated() {
	stats.Record(context.Background(), mDaemonSetsUpdated.M(int64(1)))
}

// RecordDaemonSetAdded increments the metric that records daemonset add events received.
func RecordDaemonSetAdded() {
	stats.Record(context.Background(), mDaemonSetsAdded.M(int64(1)))
}

// RecordDaemonSetDeleted increments the metric that records daemonset delete events received.
func RecordDaemonSetDeleted() {
	stats.Record(context.Background(), mDaemonSetsDeleted.M(int64(1)))
}

// RecordDaemonSetTableSize stores the size of the daemonset table field in WatchClient
func RecordDaemonSetTableSize(daemonSetTableSize int64) {
	stats.Record(context.Background(), mDaemonSetTableSize.M(daemonSetTableSize))
}

// RecordJobUpdated increments the metric that records job update events received.
func RecordJobUpdated() {
	stats.Record(context.Background(), mJobsUpdated.M(int64(1)))
}

// RecordJobAdded increments the metric that records job add events received.
func RecordJobAdded() {
	stats.Record(context.Background(), mJobsAdded.M(int64(1)))
}

// RecordJobDeleted increments the metric that records job delete events received.
func RecordJobDeleted() {
	stats.Record(context.Background(), mJobsDeleted.M(int64(1)))
}

// RecordJobTableSize stores the size of the job table field in WatchClient
func RecordJobTableSize(jobTableSize int64) {
	stats.Record(context.Background(), mJobTableSize.M(jobTableSize))
}

// RecordCronJobUpdated increments the metric that records cronjob update events received.
func RecordCronJobUpdated() {
	stats.Record(context.Background(), mCronJobsUpdated.M(int64(1)))
}

// RecordCronJobAdded increments the metric that records cronjob add events received.
func RecordCronJobAdded() {
	stats.Record(context.Background(), mCronJobsAdded.M(int64(1)))
}

// RecordCronJobDeleted increments the metric that records cronjob delete events received.
func RecordCronJobDeleted() {
	stats.Record(context.Background(), mCronJobsDeleted.M(int64(1)))
}

// RecordCronJobTableSize stores the size of the cronjob table field in WatchClient
func RecordCronJobTableSize(cronJobTableSize int64) {
	stats.Record(context.Background(), mCronJobTableSize.M(cronJobTableSize))
}

// RecordNodeUpdated increments the metric that records node update events received.
func RecordNodeUpdated() {
	stats.Record(context.Background(), mNodesUpdated.M(int64(1)))
}

// RecordNodeAdded increments the metric that records node add events received.
func RecordNodeAdded() {
	stats.Record(context.Background(), mNodesAdded.M(int64(1)))
}

// RecordNodeDeleted increments the metric that records node delete events received.
func RecordNodeDeleted() {
	stats.Record(context.Background(), mNodesDeleted.M(int64(1)))
}

// RecordNodeTableSize stores the size of the node table field in WatchClient
func RecordNodeTableSize(nodeTableSize int64) {
	stats.Record(context.Background(), mNodeTableSize.M(nodeTableSize))
}

// RecordPersistentVolumeUpdated increments the metric that records persistent volume update events received.
func RecordPersistentVolumeUpdated() {
	stats.Record(context.Background(), mPersistentVolumesUpdated.M(int64(1)))
}

// RecordPersistentVolumeAdded increments the metric that records persistent volume add events received.
func RecordPersistentVolumeAdded() {
	stats.Record(context.Background(), mPersistentVolumesAdded.M(int64(1)))
}

// RecordPersistentVolumeDeleted increments the metric that records persistent volume delete events received.
func RecordPersistentVolumeDeleted() {
	stats.Record(context.Background(), mPersistentVolumesDeleted.M(int64(1)))
}

// RecordPersistentVolumeTableSize stores the size of the persistent volume table field in WatchClient
func RecordPersistentVolumeTableSize(persistentVolumeTableSize int64) {
	stats.Record(context.Background(), mPersistentVolumeTableSize.M(persistentVolumeTableSize))
}

// RecordPersistentVolumeClaimUpdated increments the metric that records persistent volume update events received.
func RecordPersistentVolumeClaimUpdated() {
	stats.Record(context.Background(), mPersistentVolumeClaimsUpdated.M(int64(1)))
}

// RecordPersistentVolumeClaimAdded increments the metric that records persistent volume add events received.
func RecordPersistentVolumeClaimAdded() {
	stats.Record(context.Background(), mPersistentVolumeClaimsAdded.M(int64(1)))
}

// RecordPersistentVolumeClaimDeleted increments the metric that records persistent volume delete events received.
func RecordPersistentVolumeClaimDeleted() {
	stats.Record(context.Background(), mPersistentVolumeClaimsDeleted.M(int64(1)))
}

// RecordPersistentVolumeClaimTableSize stores the size of the persistent volume table field in WatchClient
func RecordPersistentVolumeClaimTableSize(persistentVolumeClaimTableSize int64) {
	stats.Record(context.Background(), mPersistentVolumeTableSize.M(persistentVolumeClaimTableSize))
}

// RecordServiceUpdated increments the metric that records persistent volume update events received.
func RecordServiceUpdated() {
	stats.Record(context.Background(), mServicesUpdated.M(int64(1)))
}

// RecordServiceAdded increments the metric that records persistent volume add events received.
func RecordServiceAdded() {
	stats.Record(context.Background(), mServicesAdded.M(int64(1)))
}

// RecordServiceDeleted increments the metric that records persistent volume delete events received.
func RecordServiceDeleted() {
	stats.Record(context.Background(), mServicesDeleted.M(int64(1)))
}

// RecordServiceTableSize stores the size of the persistent volume table field in WatchClient
func RecordServiceTableSize(ServiceTableSize int64) {
	stats.Record(context.Background(), mPersistentVolumeTableSize.M(ServiceTableSize))
}

// Create views for each metric
var viewDeploymentsUpdated = &view.View{
	Name:        mDeploymentsUpdated.Name(),
	Description: mDeploymentsUpdated.Description(),
	Measure:     mDeploymentsUpdated,
	Aggregation: view.Sum(),
}

var viewDeploymentsAdded = &view.View{
	Name:        mDeploymentsAdded.Name(),
	Description: mDeploymentsAdded.Description(),
	Measure:     mDeploymentsAdded,
	Aggregation: view.Sum(),
}

var viewDeploymentsDeleted = &view.View{
	Name:        mDeploymentsDeleted.Name(),
	Description: mDeploymentsDeleted.Description(),
	Measure:     mDeploymentsDeleted,
	Aggregation: view.Sum(),
}

var viewDeploymentTableSize = &view.View{
	Name:        mDeploymentTableSize.Name(),
	Description: mDeploymentTableSize.Description(),
	Measure:     mDeploymentTableSize,
	Aggregation: view.LastValue(),
}

var viewStatefulSetsUpdated = &view.View{
	Name:        mStatefulSetsUpdated.Name(),
	Description: mStatefulSetsUpdated.Description(),
	Measure:     mStatefulSetsUpdated,
	Aggregation: view.Sum(),
}

var viewStatefulSetsAdded = &view.View{
	Name:        mStatefulSetsAdded.Name(),
	Description: mStatefulSetsAdded.Description(),
	Measure:     mStatefulSetsAdded,
	Aggregation: view.Sum(),
}

var viewStatefulSetsDeleted = &view.View{
	Name:        mStatefulSetsDeleted.Name(),
	Description: mStatefulSetsDeleted.Description(),
	Measure:     mStatefulSetsDeleted,
	Aggregation: view.Sum(),
}

var viewStatefulSetTableSize = &view.View{
	Name:        mStatefulSetTableSize.Name(),
	Description: mStatefulSetTableSize.Description(),
	Measure:     mStatefulSetTableSize,
	Aggregation: view.LastValue(),
}

var viewReplicaSetsUpdated = &view.View{
	Name:        mReplicaSetsUpdated.Name(),
	Description: mReplicaSetsUpdated.Description(),
	Measure:     mReplicaSetsUpdated,
	Aggregation: view.Sum(),
}

var viewReplicaSetsAdded = &view.View{
	Name:        mReplicaSetsAdded.Name(),
	Description: mReplicaSetsAdded.Description(),
	Measure:     mReplicaSetsAdded,
	Aggregation: view.Sum(),
}

var viewReplicaSetsDeleted = &view.View{
	Name:        mReplicaSetsDeleted.Name(),
	Description: mReplicaSetsDeleted.Description(),
	Measure:     mReplicaSetsDeleted,
	Aggregation: view.Sum(),
}

var viewReplicaSetTableSize = &view.View{
	Name:        mReplicaSetTableSize.Name(),
	Description: mReplicaSetTableSize.Description(),
	Measure:     mReplicaSetTableSize,
	Aggregation: view.LastValue(),
}

var viewDaemonSetsUpdated = &view.View{
	Name:        mDaemonSetsUpdated.Name(),
	Description: mDaemonSetsUpdated.Description(),
	Measure:     mDaemonSetsUpdated,
	Aggregation: view.Sum(),
}

var viewDaemonSetsAdded = &view.View{
	Name:        mDaemonSetsAdded.Name(),
	Description: mDaemonSetsAdded.Description(),
	Measure:     mDaemonSetsAdded,
	Aggregation: view.Sum(),
}

var viewDaemonSetsDeleted = &view.View{
	Name:        mDaemonSetsDeleted.Name(),
	Description: mDaemonSetsDeleted.Description(),
	Measure:     mDaemonSetsDeleted,
	Aggregation: view.Sum(),
}

var viewDaemonSetTableSize = &view.View{
	Name:        mDaemonSetTableSize.Name(),
	Description: mDaemonSetTableSize.Description(),
	Measure:     mDaemonSetTableSize,
	Aggregation: view.LastValue(),
}

var viewJobsUpdated = &view.View{
	Name:        mJobsUpdated.Name(),
	Description: mJobsUpdated.Description(),
	Measure:     mJobsUpdated,
	Aggregation: view.Sum(),
}

var viewJobsAdded = &view.View{
	Name:        mJobsAdded.Name(),
	Description: mJobsAdded.Description(),
	Measure:     mJobsAdded,
	Aggregation: view.Sum(),
}

var viewJobsDeleted = &view.View{
	Name:        mJobsDeleted.Name(),
	Description: mJobsDeleted.Description(),
	Measure:     mJobsDeleted,
	Aggregation: view.Sum(),
}

var viewJobTableSize = &view.View{
	Name:        mJobTableSize.Name(),
	Description: mJobTableSize.Description(),
	Measure:     mJobTableSize,
	Aggregation: view.LastValue(),
}

var viewCronJobsUpdated = &view.View{
	Name:        mCronJobsUpdated.Name(),
	Description: mCronJobsUpdated.Description(),
	Measure:     mCronJobsUpdated,
	Aggregation: view.Sum(),
}

var viewCronJobsAdded = &view.View{
	Name:        mCronJobsAdded.Name(),
	Description: mCronJobsAdded.Description(),
	Measure:     mCronJobsAdded,
	Aggregation: view.Sum(),
}

var viewCronJobsDeleted = &view.View{
	Name:        mCronJobsDeleted.Name(),
	Description: mCronJobsDeleted.Description(),
	Measure:     mCronJobsDeleted,
	Aggregation: view.Sum(),
}

var viewCronJobTableSize = &view.View{
	Name:        mCronJobTableSize.Name(),
	Description: mCronJobTableSize.Description(),
	Measure:     mCronJobTableSize,
	Aggregation: view.LastValue(),
}

var viewNodesUpdated = &view.View{
	Name:        mNodesUpdated.Name(),
	Description: mNodesUpdated.Description(),
	Measure:     mNodesUpdated,
	Aggregation: view.Sum(),
}

var viewNodesAdded = &view.View{
	Name:        mNodesAdded.Name(),
	Description: mNodesAdded.Description(),
	Measure:     mNodesAdded,
	Aggregation: view.Sum(),
}

var viewNodesDeleted = &view.View{
	Name:        mNodesDeleted.Name(),
	Description: mNodesDeleted.Description(),
	Measure:     mNodesDeleted,
	Aggregation: view.Sum(),
}

var viewNodeTableSize = &view.View{
	Name:        mNodeTableSize.Name(),
	Description: mNodeTableSize.Description(),
	Measure:     mNodeTableSize,
	Aggregation: view.LastValue(),
}

var viewPersistentVolumesUpdated = &view.View{
	Name:        mPersistentVolumesUpdated.Name(),
	Description: mPersistentVolumesUpdated.Description(),
	Measure:     mPersistentVolumesUpdated,
	Aggregation: view.Sum(),
}

var viewPersistentVolumesAdded = &view.View{
	Name:        mPersistentVolumesAdded.Name(),
	Description: mPersistentVolumesAdded.Description(),
	Measure:     mPersistentVolumesAdded,
	Aggregation: view.Sum(),
}

var viewPersistentVolumesDeleted = &view.View{
	Name:        mPersistentVolumesDeleted.Name(),
	Description: mPersistentVolumesDeleted.Description(),
	Measure:     mPersistentVolumesDeleted,
	Aggregation: view.Sum(),
}

var viewPersistentVolumeTableSize = &view.View{
	Name:        mPersistentVolumeTableSize.Name(),
	Description: mPersistentVolumeTableSize.Description(),
	Measure:     mPersistentVolumeTableSize,
	Aggregation: view.LastValue(),
}

var viewPersistentVolumeClaimsUpdated = &view.View{
	Name:        mPersistentVolumeClaimsUpdated.Name(),
	Description: mPersistentVolumeClaimsUpdated.Description(),
	Measure:     mPersistentVolumeClaimsUpdated,
	Aggregation: view.Sum(),
}

var viewPersistentVolumeClaimsAdded = &view.View{
	Name:        mPersistentVolumeClaimsAdded.Name(),
	Description: mPersistentVolumeClaimsAdded.Description(),
	Measure:     mPersistentVolumeClaimsAdded,
	Aggregation: view.Sum(),
}

var viewPersistentVolumeClaimsDeleted = &view.View{
	Name:        mPersistentVolumeClaimsDeleted.Name(),
	Description: mPersistentVolumeClaimsDeleted.Description(),
	Measure:     mPersistentVolumeClaimsDeleted,
	Aggregation: view.Sum(),
}

var viewPersistentVolumeClaimTableSize = &view.View{
	Name:        mPersistentVolumeClaimTableSize.Name(),
	Description: mPersistentVolumeClaimTableSize.Description(),
	Measure:     mPersistentVolumeClaimTableSize,
	Aggregation: view.LastValue(),
}

var viewServicesUpdated = &view.View{
	Name:        mServicesUpdated.Name(),
	Description: mServicesUpdated.Description(),
	Measure:     mServicesUpdated,
	Aggregation: view.Sum(),
}

var viewServicesAdded = &view.View{
	Name:        mServicesAdded.Name(),
	Description: mServicesAdded.Description(),
	Measure:     mServicesAdded,
	Aggregation: view.Sum(),
}

var viewServicesDeleted = &view.View{
	Name:        mServicesDeleted.Name(),
	Description: mServicesDeleted.Description(),
	Measure:     mServicesDeleted,
	Aggregation: view.Sum(),
}

var viewServiceTableSize = &view.View{
	Name:        mServiceTableSize.Name(),
	Description: mServiceTableSize.Description(),
	Measure:     mServiceTableSize,
	Aggregation: view.LastValue(),
}