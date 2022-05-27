package prometheustypeconverterprocessor

import (
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

type builder struct {
	metric *metricspb.Metric
}

// metricBuilder is used to build metrics for testing
func metricBuilder() builder {
	return builder{
		metric: &metricspb.Metric{
			MetricDescriptor: &metricspb.MetricDescriptor{},
			Timeseries:       make([]*metricspb.TimeSeries, 0),
		},
	}
}

// setName sets the name of the metric
func (b builder) setName(name string) builder {
	b.metric.MetricDescriptor.Name = name
	return b
}

// setDataType sets the data type of this metric
func (b builder) setDataType(dataType metricspb.MetricDescriptor_Type) builder {
	b.metric.MetricDescriptor.Type = dataType
	return b
}

// Build builds from the builder to the final metric
func (b builder) build() *metricspb.Metric {
	return b.metric
}
