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

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// filterDataPoints filters data points according to the provided data point value
func filterDataPoints(metric pmetric.Metric, mtpOp internalOperation) {
	action := mtpOp.configOperation.DataPointValueAction

	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		metric.Gauge().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
			switch dp.ValueType() {

			case pmetric.NumberDataPointValueTypeInt:
				return !includeDataPoint(float64(dp.IntValue()), mtpOp.configOperation.DataPointValue, action)
			case pmetric.NumberDataPointValueTypeDouble:
				return !includeDataPoint(dp.DoubleValue(), mtpOp.configOperation.DataPointValue, action)
			}

			// if double or int value is not found, consider value to be zero
			return !includeDataPoint(0, mtpOp.configOperation.DataPointValue, action)
		})
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
