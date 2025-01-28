package containerprocessor

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

const (
	k8sLogType = "sw.k8s.log.type"

	// Attributes for OTel entity events identification
	otelEntityEventAsLog = "otel.entity.event_as_log"
	otelEntityEventType  = "otel.entity.event.type"
	swEntityType         = "otel.entity.type"

	// Attributes for telemetry mapping
	otelEntityId     = "otel.entity.id"
	k8sContainerName = "k8s.container.name"
	k8sNamespaceName = "k8s.namespace.name"
	k8sPodName       = "k8s.pod.name"
	swK8sClusterUid  = "sw.k8s.cluster.uid"

	// Attributes containing additional information about container
	otelEntityAttributes = "otel.entity.attributes"
	k8sContainerStatus   = "sw.k8s.container.status"
	k8sContainerInit     = "sw.k8s.container.init"
	k8sContainerSidecar  = "sw.k8s.container.sidecar"
	k8sContainerId       = "container.id"
)

func NewContainerResourceLogs() plog.ResourceLogs {
	rl := plog.NewResourceLogs()
	rl.Resource().Attributes().PutStr(k8sLogType, "manifest")
	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().Attributes().PutBool(otelEntityEventAsLog, true)
	return rl
}

func transformManifestToContainerLogs(m Manifest) plog.LogRecordSlice {
	var lrs plog.LogRecordSlice
	conditions := m.Status.Conditions
	lastChange := conditions[len(conditions)-1].Timestamp
	t, err := time.Parse(time.RFC3339, lastChange)
	if err != nil {
		t = time.Now()
	}

	containers := m.getContainers()
	for _, c := range containers {
		lr := lrs.AppendEmpty()
		lr.SetTimestamp(pcommon.NewTimestampFromTime(t))
		addContainerAttributes(lr.Attributes(), m.Metadata, c)
	}

	return lrs
}

func addContainerAttributes(attrs pcommon.Map, md Metadata, c Container) {
	// Ingestion attributes
	attrs.PutStr(otelEntityEventType, "entity_state")
	attrs.PutStr(swEntityType, "KubernetesContainer")

	// Telemetry mappings
	tm := attrs.PutEmptyMap(otelEntityId)
	tm.PutStr(k8sPodName, md.PodName)
	tm.PutStr(k8sNamespaceName, md.Namespace)
	tm.PutStr(k8sContainerName, c.Name)
	tm.PutStr(swK8sClusterUid, md.Annotations.ClusterUid)

	// Entity attributes
	ea := attrs.PutEmptyMap(otelEntityAttributes)
	ea.PutStr(k8sContainerId, c.ContainerId)
	ea.PutStr(k8sContainerStatus, c.State)
	ea.PutBool(k8sContainerInit, false)
	ea.PutBool(k8sContainerSidecar, c.IsSidecarContainer)
}
