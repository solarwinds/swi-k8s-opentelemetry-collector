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

func (cp *containerprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()

	manifests, err := cp.extractPodManifests(resourceLogs)
	if err != nil {
		cp.logger.Warn("Failed to parse manifests")
		return plog.NewLogs(), err
	}

	if len(manifests) == 0 {
		return ld, nil
	}

	rl := NewContainerResourceLogs()
	rl.CopyTo(ld.ResourceLogs().AppendEmpty())
	lrs := rl.ScopeLogs().At(0).LogRecords()

	for _, m := range manifests {
		cp.logger.Debug("Processing manifest", zap.String("pod-name", m.Metadata.PodName))
		containers := transformManifestToContainerLogs(m)
		containers.MoveAndAppendTo(lrs)
	}

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

func isPodLog(attributes pcommon.Map) bool {
	kind, _ := attributes.Get("k8s.object.kind")
	return kind.Str() == "Pod"
}

func (cp *containerprocessor) Start(_ context.Context, _ component.Host) error {
	cp.logger.Info("Starting container processor")
	return nil
}

func (cp *containerprocessor) Shutdown(_ context.Context) error {
	cp.logger.Info("Shutting down container processor")
	return nil
}
