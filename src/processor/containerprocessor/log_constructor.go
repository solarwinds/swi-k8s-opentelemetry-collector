package containerprocessor

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"time"
)

const (
	otelEntityEventAsLog = "otel.entity.event_as_log"
	k8sLogType           = "sw.k8s.log.type"
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
		decorateLogRecord(lr, m.Metadata, c)
	}

	return lrs
}

func decorateLogRecord(lr plog.LogRecord, md Metadata, c Container) {
	attrs := lr.Attributes()
	attrs.PutStr("sw.k8s.pod.name", md.PodName)
	attrs.PutStr("sw.k8s.pod.namespace", md.Namespace)
	attrs.PutStr("sw.k8s.container.name", c.Name)
	attrs.PutStr("sw.k8s.container.id", c.ContainerId)
	attrs.PutStr("sw.k8s.container.state", c.State)
	attrs.PutBool("sw.k8s.container.init", false)
	attrs.PutBool("sw.k8s.container.sidecar", c.IsSidecarContainer)
}
