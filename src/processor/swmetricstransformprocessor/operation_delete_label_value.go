// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
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

import (
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

// deleteLabelValueOp deletes a label value and all data associated with it
func (mtp *swMetricsTransformProcessor) deleteLabelValueOp(metric *metricspb.Metric, mtpOp internalOperation) {
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
