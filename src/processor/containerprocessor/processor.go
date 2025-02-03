package containerprocessor

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"sync"
)

const (
	k8sObjectKind = "k8s.object.kind"
)

type containerprocessor struct {
	cfg               component.Config
	telemetrySettings component.TelemetrySettings
	logger            *zap.Logger
}

// processLogs go through all log records and parse information about containers from them.
// Containers are created based on all log records from all scope and resource logs.
// Containers related logs are appended as a new ResourceLogs to the plog.Logs structure that is processed at the time.
func (cp *containerprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	mCh := make(chan Manifest)
	errCh := make(chan error)

	logSlice := plog.NewLogRecordSlice()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go cp.generateLogRecords(mCh, wg, logSlice)
	go cp.generateManifests(mCh, errCh, wg, resourceLogs)

	select {
	case err := <-errCh:
		return ld, err
	default:
	}

	wg.Wait()

	if logSlice.Len() > 0 {
		rl := addContainersResourceLog(ld)
		lrs := rl.ScopeLogs().At(0).LogRecords()
		logSlice.CopyTo(lrs)
	}

	return ld, nil
}

// generateLogRecords appends all LogRecords containing container information to the provided LogRecordSlice.
func (cp *containerprocessor) generateLogRecords(mCh <-chan Manifest, wg *sync.WaitGroup, lrs plog.LogRecordSlice) {
	defer wg.Done()
	for m := range mCh {
		containers := transformManifestToContainerLogs(m)
		for i := range containers.Len() {
			lr := containers.At(i)
			lr.CopyTo(lrs.AppendEmpty())
		}
	}
}

// generateManifests extracts and parses manifests from log records that have k8s.object.kind set to "Pod".
func (cp *containerprocessor) generateManifests(mCh chan<- Manifest, errCh chan<- error, wg *sync.WaitGroup, resourceLogs plog.ResourceLogsSlice) {
	defer wg.Done()
	defer close(mCh)

	for i := range resourceLogs.Len() {
		rl := resourceLogs.At(i)
		scopeLogs := rl.ScopeLogs()

		for j := range scopeLogs.Len() {
			sl := scopeLogs.At(j)
			logRecords := sl.LogRecords()

			for k := range logRecords.Len() {
				lr := logRecords.At(k)
				attrs := lr.Attributes()

				// processor is interested in pods only, since containers are related to pods
				if !isPodLog(attrs) {
					continue
				}

				body := lr.Body().AsString()
				var m Manifest

				err := json.Unmarshal([]byte(body), &m)
				if err != nil {
					cp.logger.Error("Error while unmarshalling manifest", zap.Error(err))
					errCh <- err
					return
				} else {
					mCh <- m
				}
			}
		}
	}
}

func isPodLog(attributes pcommon.Map) bool {
	kind, _ := attributes.Get(k8sObjectKind)
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
