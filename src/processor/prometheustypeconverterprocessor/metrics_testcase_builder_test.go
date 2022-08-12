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
