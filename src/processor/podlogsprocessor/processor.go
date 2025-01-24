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

type podStatus struct {
	ContainerStatuses []struct {
		Name        string                 `json:"name"`
		ContainerId string                 `json:"containerID"`
		State       map[string]interface{} `json:"state"`
	}
}

type bodyLog struct {
	Kind      string    `json:"kind"`
	PodStatus podStatus `json:"status"`
}

func (kp *kubernetesprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
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
				attributes := logRecord.Attributes()

				if !isPodLog(attributes) {
					continue
				}

				body := logRecord.Body().AsString()
				var logResult bodyLog

				err := json.Unmarshal([]byte(body), &logResult)
				if err != nil {
					// TODO: Handle error, return empty or original logs?
					kp.logger.Info("ERROR while unmarshaling")
					return plog.Logs{}, err
				}

				attrs := logRecord.Attributes()
				statuses := attrs.PutEmptySlice("containers")
				for _, c := range logResult.PodStatus.ContainerStatuses {
					containerMap := statuses.AppendEmpty().SetEmptyMap()
					containerMap.PutStr("name", c.Name)
					containerMap.PutStr("id", c.ContainerId)
					for key, _ := range c.State {
						containerMap.PutStr("status", key)
						break
					}
				}
			}
		}
	}

	return ld, nil
}

func isPodLog(attributes pcommon.Map) bool {
	kind, _ := attributes.Get("k8s.object.kind")
	return kind.Str() == "Pod"
}

func (p *kubernetesprocessor) logResourceAttributes(rl plog.ResourceLogs) {
	j, _ := json.Marshal(rl)
	p.logger.Info("Resource logs", zap.String("resource-logs", string(j)))
}

func (p *kubernetesprocessor) addAttributes(_ context.Context, attributes pcommon.Map) {
	attributes.PutStr("janca.attribute", "ACHJO")
}

func (kp *kubernetesprocessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (kp *kubernetesprocessor) Shutdown(_ context.Context) error {
	return nil
}
