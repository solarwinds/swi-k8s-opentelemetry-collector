package containerprocessor

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"os"
	"time"
)

const (
	k8sLogType = "sw.k8s.log.type"

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
	rl.Resource().Attributes().PutStr(k8sLogType, "manifest")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().Attributes().PutBool(otelEntityEventAsLog, true)
	return rl
}

// transformManifestToContainerLogs returns a new plog.LogRecordSlice and appends
// all LogRecords containing container information from the provided Manifest.
func transformManifestToContainerLogs(m Manifest) plog.LogRecordSlice {
	lrs := plog.NewLogRecordSlice()
	t := getTimestampOfLatestChange(m.Status.Conditions)

	containers := m.getContainers()
	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetTimestamp(pcommon.NewTimestampFromTime(t))
		addContainerAttributes(lr.Attributes(), m.Metadata, c)
	}

	return lrs
}

// getTimestampOfLatestChange determines the timestamp of incoming change from "conditions" part of the manifest,
// by using the latest item from the slice. If the slice is empty, the current time is considered as the
// timestamp of the change.
func getTimestampOfLatestChange(changes []Condition) time.Time {
	var t time.Time
	var lastChange string
	var err error
	if len(changes) > 0 {
		lastChange = changes[len(changes)-1].Timestamp
		t, err = time.Parse(time.RFC3339, lastChange)
		if err != nil {
			t = time.Now()
		}
	} else {
		t = time.Now()
	}
	return t
}

// addContainerAttributes sets attributes on the provided map for the given Metadata and Container.
func addContainerAttributes(attrs pcommon.Map, md Metadata, c Container) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, "entity_state")
	attrs.PutStr(swEntityType, "KubernetesContainer")

	// Telemetry mappings
	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(conventions.AttributeK8SPodName, md.PodName)
	tm.PutStr(conventions.AttributeK8SNamespaceName, md.Namespace)
	tm.PutStr(conventions.AttributeK8SContainerName, c.Name)
	tm.PutStr(swK8sClusterUid, os.Getenv("CLUSTER_UID"))

	// Entity attributes
	ea := attrs.PutEmptyMap(otelEntityAttributes)
	ea.PutStr(conventions.AttributeContainerID, c.ContainerId)
	ea.PutStr(k8sContainerStatus, c.State)
	ea.PutBool(k8sContainerInit, false)
	ea.PutBool(k8sContainerSidecar, c.IsSidecarContainer)
}
