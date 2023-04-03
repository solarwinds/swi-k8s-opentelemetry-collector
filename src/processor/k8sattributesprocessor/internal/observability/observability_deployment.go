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
	)
}

var (
	mDeploymentsUpdated  = stats.Int64("otelsvc/k8s/deployment_updated", "Number of deployment update events received", "1")
	mDeploymentsAdded    = stats.Int64("otelsvc/k8s/deployment_added", "Number of deployment add events received", "1")
	mDeploymentsDeleted  = stats.Int64("otelsvc/k8s/deployment_deleted", "Number of deployment delete events received", "1")
	mDeploymentTableSize = stats.Int64("otelsvc/k8s/deployment_table_size", "Size of table containing deployment info", "1")
)

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
