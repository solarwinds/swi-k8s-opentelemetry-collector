package containerprocessor

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type containerprocessor struct {
	cfg               component.Config
	telemetrySettings component.TelemetrySettings
	logger            *zap.Logger
}

type Container struct {
	Name               string
	ContainerId        string
	State              string
	IsInitContainer    bool
	IsSidecarContainer bool
	Timestamp          string
}

func (cp *containerprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	containers := make([]Container, 0)

	manifests, err := cp.extractPodManifests(resourceLogs)
	if err != nil {
		cp.logger.Warn("Failed to parse manifest for pod log record")
		return ld, err
	}

	if len(manifests) == 0 {
		cp.logger.Debug("No manifests found")
		return ld, nil
	}

	for _, m := range manifests {
		containers = append(containers, m.extractContainers(cp.logger)...)
	}

	newResourceLogs := ld.ResourceLogs().AppendEmpty()
	newResource := newResourceLogs.Resource()
	newResource.Attributes().PutStr("sw.k8s.log.type", "manifest")
	addContainerResourceLog(newResourceLogs, containers)

	return ld, nil
}

func (cp *containerprocessor) extractPodManifests(resourceLogs plog.ResourceLogsSlice) ([]Manifest, error) {
	manifests := make([]Manifest, 0)

	for i := range resourceLogs.Len() {
		rl := resourceLogs.At(i)
		scopeLogs := rl.ScopeLogs()

		for j := range scopeLogs.Len() {
			sl := scopeLogs.At(j)
			logRecords := sl.LogRecords()

			for k := range logRecords.Len() {
				lr := logRecords.At(k)
				attrs := lr.Attributes()

				if !isPodLog(attrs) {
					break
				}

				body := lr.Body().AsString()
				var m Manifest

				err := json.Unmarshal([]byte(body), &m)
				if err != nil {
					cp.logger.Error("Error while unmarshalling manifest", zap.Error(err))
					return nil, err
				}
				manifests = append(manifests, m)
			}
		}
	}
	return manifests, nil
}

func addContainerResourceLog(rl plog.ResourceLogs, containers []Container) {
	newScopeLogs := rl.ScopeLogs().AppendEmpty()
	for c := range containers {
		lr := newScopeLogs.LogRecords().AppendEmpty()
		lr.Attributes().PutStr("k8s.pod.container", containers[c].Name)
		lr.Attributes().PutStr("k8s.pod.container.id", containers[c].ContainerId)
		lr.Attributes().PutStr("k8s.pod.container.state", containers[c].State)
		lr.Attributes().PutStr("k8s.pod.container.timestamp", containers[c].Timestamp)
		lr.Attributes().PutStr("k8s.pod.container.isInitContainer", "false")
		lr.Attributes().PutStr("k8s.pod.container.isSidecarContainer", "false")
	}
}

func isPodLog(attributes pcommon.Map) bool {
	kind, _ := attributes.Get("k8s.object.kind")
	return kind.Str() == "Pod"
}

func (cp *containerprocessor) logResourceAttributes(rl plog.ResourceLogs) {
	j, _ := json.Marshal(rl)
	cp.logger.Info("Resource logs", zap.String("resource-logs", string(j)))
}

func (cp *containerprocessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (cp *containerprocessor) Shutdown(_ context.Context) error {
	return nil
}
