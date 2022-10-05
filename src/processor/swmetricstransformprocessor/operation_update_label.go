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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// updateLabelOp updates labels and label values in metric based on given operation
func updateLabelOp(metric pmetric.Metric, mtpOp internalOperation, f internalFilter) {
	op := mtpOp.configOperation
	rangeDataPointAttributes(metric, func(attrs pcommon.Map) bool {
		if !f.matchAttrs(attrs) {
			return true
		}

		attrKey := op.Label
		attrVal, ok := attrs.Get(attrKey)
		if !ok {
			return true
		}

		if op.NewLabel != "" {
			attrVal.CopyTo(attrs.PutEmpty(op.NewLabel))
			attrs.Remove(attrKey)
			attrKey = op.NewLabel
		}

		if newValue, ok := mtpOp.valueActionsMapping[attrVal.Str()]; ok {
			attrs.PutString(attrKey, newValue)
		}
		return true
	})
}
