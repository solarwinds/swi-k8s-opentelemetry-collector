package filterdatapointsprocessor

import (
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

// updateLabelOp updates labels and label values in metric based on given operation
func (mtp *filterDataPointsProcessor) updateLabelOp(metric *metricspb.Metric, mtpOp internalOperation) {
	op := mtpOp.configOperation
	for idx, label := range metric.MetricDescriptor.LabelKeys {
		if label.Key != op.Label {
			continue
		}

		if op.NewLabel != "" {
			label.Key = op.NewLabel
		}

		labelValuesMapping := mtpOp.valueActionsMapping
		for _, timeseries := range metric.Timeseries {
			newValue, ok := labelValuesMapping[timeseries.LabelValues[idx].Value]
			if ok {
				timeseries.LabelValues[idx].Value = newValue
			}
		}
	}
}
