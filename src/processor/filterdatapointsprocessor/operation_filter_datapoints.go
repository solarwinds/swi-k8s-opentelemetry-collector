package filterdatapointsprocessor

import metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

// filterDataPoints filters data points according to the provided data point value
func (mtp *filterDataPointsProcessor) filterDataPoints(metric *metricspb.Metric, op internalOperation) {
	for _, ts := range metric.Timeseries {
		n := 0
		for _, dp := range ts.Points {
			switch metric.MetricDescriptor.Type {
			case metricspb.MetricDescriptor_GAUGE_INT64, metricspb.MetricDescriptor_CUMULATIVE_INT64:
				if float64(dp.GetInt64Value()) == op.configOperation.DataPointValue {
					ts.Points[n] = dp
					n++
				}
			case metricspb.MetricDescriptor_GAUGE_DOUBLE, metricspb.MetricDescriptor_CUMULATIVE_DOUBLE:
				if dp.GetDoubleValue() == op.configOperation.DataPointValue {
					ts.Points[n] = dp
					n++
				}
			}
		}
		ts.Points = ts.Points[:n]
	}
}
