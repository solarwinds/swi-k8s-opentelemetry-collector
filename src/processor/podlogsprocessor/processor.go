package podlogsprocessor

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type kubernetesprocessor struct {
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

func (kp *kubernetesprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	containers := make([]Container, 0)

	manifests, err := kp.extractManifests(resourceLogs)
	if err != nil {
		kp.logger.Info("ERROR while extracting manifests")
		return plog.NewLogs(), err
	}

	if len(manifests) == 0 {
		kp.logger.Info("No manifests found")
		return ld, nil
	}

	for _, m := range manifests {
		containers = append(containers, m.extractContainers(kp.logger)...)
	}

	newResourceLogs := ld.ResourceLogs().AppendEmpty()
	newResource := newResourceLogs.Resource()
	newResource.Attributes().PutStr("sw.k8s.log.type", "manifest")
	addContainerResourceLog(newResourceLogs, containers)

	return ld, nil
}

func (kp *kubernetesprocessor) extractManifests(resourceLogs plog.ResourceLogsSlice) ([]Manifest, error) {
	manifests := make([]Manifest, 0)
	resourceLogsLength := resourceLogs.Len()

	for i := range resourceLogsLength {
		resourceLog := resourceLogs.At(i)
		scopeLogs := resourceLog.ScopeLogs()
		scopeLogsLength := scopeLogs.Len()

		for j := range scopeLogsLength {
			scopeLog := scopeLogs.At(j)
			logRecords := scopeLog.LogRecords()
			logRecordsLength := logRecords.Len()

			for k := range logRecordsLength {
				logRecord := logRecords.At(k)
				attrs := logRecord.Attributes()

				if !isPodLog(attrs) {
					continue
				}

				body := logRecord.Body().AsString()
				var logResult Manifest

				err := json.Unmarshal([]byte(body), &logResult)
				if err != nil {
					// TODO: Handle error, return empty or original logs?
					kp.logger.Info("ERROR while unmarshaling")
					return nil, err
				}

				manifests = append(manifests, logResult)

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

func (p *kubernetesprocessor) logResourceAttributes(rl plog.ResourceLogs) {
	j, _ := json.Marshal(rl)
	p.logger.Info("Resource logs", zap.String("resource-logs", string(j)))
}

func (kp *kubernetesprocessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (kp *kubernetesprocessor) Shutdown(_ context.Context) error {
	return nil
}
