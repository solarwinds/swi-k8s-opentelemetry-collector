package prometheustypeconverterprocessor

import (
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

type prometheusTypeConverterTest struct {
	name       string // test name
	transforms []internalTransform
	in         []*metricspb.Metric
	out        []*metricspb.Metric
}

var (
	// test cases
	standardTests = []prometheusTypeConverterTest{
		// Convert type to sum
		{
			name: "metric_convert_type_to_sum",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).build(),
			},
		},
	}
)
