package filterdatapointsprocessor

import (
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

// deleteLabelValueOp deletes a label value and all data associated with it
func (mtp *filterDataPointsProcessor) deleteLabelValueOp(metric *metricspb.Metric, mtpOp internalOperation) {
	op := mtpOp.configOperation
	for idx, label := range metric.MetricDescriptor.LabelKeys {
		if label.Key != op.Label {
			continue
		}

		newTimeseries := make([]*metricspb.TimeSeries, 0)
		for _, timeseries := range metric.Timeseries {
			if timeseries.LabelValues[idx].Value == op.LabelValue {
				continue
			}
			newTimeseries = append(newTimeseries, timeseries)
		}
		metric.Timeseries = newTimeseries
	}
}
