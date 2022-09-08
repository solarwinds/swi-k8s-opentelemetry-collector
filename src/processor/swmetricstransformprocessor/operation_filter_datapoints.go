// Copyright 2022 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor
// Changes customizing the original processor:
//	- removal of actions: toggle_scalar_data_type, experimental_scale_value, aggregate_labels, aggregate_label_values
//	- add custom action "filter_datapoints"
//	- rename types and functions to match the processor name

package swmetricstransformprocessor

import metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"

// filterDataPoints filters data points according to the provided data point value
func (mtp *swMetricsTransformProcessor) filterDataPoints(metric *metricspb.Metric, op internalOperation) {
	action := op.configOperation.DataPointValueAction
	for _, ts := range metric.Timeseries {
		n := 0
		for _, dp := range ts.Points {
			switch metric.MetricDescriptor.Type {
			case metricspb.MetricDescriptor_GAUGE_INT64, metricspb.MetricDescriptor_CUMULATIVE_INT64:
				if includeDataPoint(float64(dp.GetInt64Value()), op.configOperation.DataPointValue, action) {
					ts.Points[n] = dp
					n++
				}
			case metricspb.MetricDescriptor_GAUGE_DOUBLE, metricspb.MetricDescriptor_CUMULATIVE_DOUBLE:
				if includeDataPoint(dp.GetDoubleValue(), op.configOperation.DataPointValue, action) {
					ts.Points[n] = dp
					n++
				}
			}
		}
		ts.Points = ts.Points[:n]
	}
}

func includeDataPoint(dataPointValue float64, filterValue float64, action DataPointValueAction) bool {
	switch action {
	case Include:
		return dataPointValue == filterValue
	case Exclude:
		return dataPointValue != filterValue
	}

	return true
}
