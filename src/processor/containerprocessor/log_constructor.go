// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
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

package containerprocessor

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"os"
)

const (
	k8sLogType             = "sw.k8s.log.type"
	clusterUidEnv          = "CLUSTER_UID"
	k8sContainerEntityType = "KubernetesContainer"
	entityState            = "entity_state"

	// Attributes for OTel entity events identification
	otelEntityEventAsLog = "otel.entity.event_as_log"
	otelEntityEventType  = "otel.entity.event.type"
	swEntityType         = "otel.entity.type"

	// Attributes for telemetry mapping
	otelEntityId    = "otel.entity.id"
	swK8sClusterUid = "sw.k8s.cluster.uid"

	// Attributes containing additional information about container
	otelEntityAttributes = "otel.entity.attributes"
	k8sContainerStatus   = "sw.k8s.container.status"
	k8sContainerInit     = "sw.k8s.container.init"
	k8sContainerSidecar  = "sw.k8s.container.sidecar"
)

// addContainersResourceLog adds a new ResourceLogs to the provided Logs structure
// and sets required attributes on "resource" and "scopeLogs"
func addContainersResourceLog(ld plog.Logs) plog.ResourceLogs {
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr(k8sLogType, "entitystateevent")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().Attributes().PutBool(otelEntityEventAsLog, true)
	return rl
}

// transformManifestToContainerLogs returns a new plog.LogRecordSlice and appends
// all LogRecords containing container information from the provided Manifest.
func transformManifestToContainerLogs(m Manifest, t pcommon.Timestamp) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()

	containers := m.getContainers()
	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetObservedTimestamp(t)
		addContainerAttributes(lr.Attributes(), m.Metadata, c)
	}

	return lrs
}

// addContainerAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerAttributes(attrs pcommon.Map, md Metadata, c Container) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, k8sContainerEntityType)

	// Telemetry mappings
	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(conventions.AttributeK8SPodName, md.PodName)
	tm.PutStr(conventions.AttributeK8SNamespaceName, md.Namespace)
	tm.PutStr(conventions.AttributeK8SContainerName, c.Name)
	tm.PutStr(swK8sClusterUid, os.Getenv(clusterUidEnv))

	// Entity attributes
	ea := attrs.PutEmptyMap(otelEntityAttributes)
	ea.PutStr(conventions.AttributeContainerID, c.ContainerId)
	ea.PutStr(k8sContainerStatus, c.State)
	ea.PutBool(k8sContainerInit, c.IsInitContainer)
	ea.PutBool(k8sContainerSidecar, c.IsSidecarContainer)
}
