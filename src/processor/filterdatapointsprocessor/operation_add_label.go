package filterdatapointsprocessor

import metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

func (mtp *filterDataPointsProcessor) addLabelOp(metric *metricspb.Metric, op internalOperation) {
	var lb = metricspb.LabelKey{
		Key: op.configOperation.NewLabel,
	}
	metric.MetricDescriptor.LabelKeys = append(metric.MetricDescriptor.LabelKeys, &lb)
	for _, ts := range metric.Timeseries {
		lv := &metricspb.LabelValue{
			Value:    op.configOperation.NewValue,
			HasValue: true,
		}
		ts.LabelValues = append(ts.LabelValues, lv)
	}
}
