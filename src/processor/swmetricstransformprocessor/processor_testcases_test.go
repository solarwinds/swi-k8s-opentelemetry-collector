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
	"regexp"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
)

type filterDataPointsTest struct {
	name       string // test name
	transforms []internalTransform
	in         []*metricspb.Metric
	out        []*metricspb.Metric
}

var (
	// test cases
	standardTests = []filterDataPointsTest{
		// UPDATE
		{
			name: "metric_name_update",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("new/metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
		},
		{
			name: "metric_name_update_chained",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					NewName:             "new/metric1",
				},
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric2"},
					Action:              Update,
					NewName:             "new/metric2",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_DOUBLE).build(),
				metricBuilder().setName("metric2").
					setDataType(metricspb.MetricDescriptor_GAUGE_DOUBLE).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("new/metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("new/metric2").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
		},
		{
			name: "metric_names_update_chained",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("^(metric)(?P<namedsubmatch>[12])$")},
					Action:              Update,
					NewName:             "new/$1/$namedsubmatch",
				},
				{
					MetricIncludeFilter: internalFilterStrict{include: "new/metric/1"},
					Action:              Update,
					NewName:             "new/new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("metric2").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("metric3").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("new/new/metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("new/metric/2").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("metric3").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
		},
		{
			name: "metric_name_update_nonexist",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "nonexist"},
					Action:              Update,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_DOUBLE).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).build(),
			},
		},
		{
			name: "metric_label_update",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   UpdateLabel,
								Label:    "label1",
								NewLabel: "new/label1",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).
					setLabels([]string{"label1"}).
					addTimeseries(1, []string{"value1"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).
					setLabels([]string{"new/label1"}).
					addTimeseries(1, []string{"value1"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_label_value_update",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action: UpdateLabel,
								Label:  "label1",
							},
							valueActionsMapping: map[string]string{
								"label1-value1": "new/label1-value1",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1"}).
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).
					addTimeseries(1, []string{"label1-value1"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1-value2"}).
					addInt64Point(1, 3, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1"}).
					setDataType(metricspb.MetricDescriptor_CUMULATIVE_INT64).
					addTimeseries(1, []string{"new/label1-value1"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1-value2"}).
					addInt64Point(1, 3, 2).
					build(),
			},
		},
		{
			name: "metric_name_insert_multiple",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Insert,
					NewName:             "new/metric1",
				},
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric2"},
					Action:              Insert,
					NewName:             "new/metric2",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("metric2").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("metric2").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("new/metric1").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
				metricBuilder().setName("new/metric2").setDataType(metricspb.MetricDescriptor_GAUGE_INT64).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1", matchLabels: map[string]StringMatcher{"label1": strictMatcher("value1")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile(`(.|\s)*\S(.|\s)*`)}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_two_datapoints_positive",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile("value3")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addTimeseries(2, []string{"value3", "value4"}).
					addInt64Point(0, 3, 2).
					addInt64Point(1, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addTimeseries(2, []string{"value3", "value4"}).
					addInt64Point(0, 3, 2).
					addInt64Point(1, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(2, []string{"value3", "value4"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_two_datapoints_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile("value3")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addTimeseries(2, []string{"value11", "value22"}).
					addInt64Point(0, 3, 2).
					addInt64Point(1, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addTimeseries(2, []string{"value11", "value22"}).
					addInt64Point(0, 3, 2).
					addInt64Point(1, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_with_full_value",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile("value1")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1", matchLabels: map[string]StringMatcher{"label1": strictMatcher("wrong_value")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile(".*wrong_ending")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict_missing_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1", matchLabels: map[string]StringMatcher{"missing_key": strictMatcher("value1")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"missing_key": regexp.MustCompile("value1")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_and_present_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile("value1"), "missing_key": regexp.MustCompile("value2")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_key_with_empty_expression",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"), matchLabels: map[string]StringMatcher{"label1": regexp.MustCompile("value1"), "missing_key": regexp.MustCompile("^$")}},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_label_update_with_metric_insert",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Insert,
					NewName:             "new/metric1",
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   UpdateLabel,
								Label:    "label1",
								NewLabel: "new/label1",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).build(),
				metricBuilder().setName("new/metric1").
					setLabels([]string{"label2", "new/label1"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value2", "value1"}).
					addInt64Point(0, 3, 2).build(),
			},
		},
		{
			name: "metric_label_value_update_with_metric_insert",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Insert,
					NewName:             "new/metric1",
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action: UpdateLabel,
								Label:  "label1",
							},
							valueActionsMapping: map[string]string{"label1-value1": "new/label1-value1"},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1-value1"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1-value2"}).
					addInt64Point(1, 4, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setLabels([]string{"label1"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1-value1"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1-value2"}).
					addInt64Point(1, 4, 2).
					build(),

				metricBuilder().setName("new/metric1").
					setLabels([]string{"label1"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"new/label1-value1"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1-value2"}).
					addInt64Point(1, 4, 2).
					build(),
			},
		},
		// Add Label to a metric
		{
			name: "update existing metric by adding a new label when there are no labels",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   AddLabel,
								NewLabel: "foo",
								NewValue: "bar",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, nil).
					addInt64Point(0, 3, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"foo"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"bar"}).
					addInt64Point(0, 3, 2).
					build(),
			},
		},
		{
			name: "update existing metric by adding a new label when there are labels",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   AddLabel,
								NewLabel: "foo",
								NewValue: "bar",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"foo", "label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"bar", "value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
		},
		{
			name: "update existing metric by adding a label that is duplicated in the list",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   AddLabel,
								NewLabel: "label1",
								NewValue: "value3",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
		},
		{
			name: "update_does_not_happen_because_target_metric_doesn't_exist",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "mymetric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:   AddLabel,
								NewLabel: "foo",
								NewValue: "bar",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric1").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"value1", "value2"}).
					addInt64Point(0, 3, 2).
					build(),
			},
		},
		// delete label value
		{
			name: "delete_a_label_value",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:     DeleteLabelValue,
								Label:      "label1",
								LabelValue: "label1value1",
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1value1", "label2value"}).
					addInt64Point(0, 3, 2).
					addTimeseries(1, []string{"label1value2", "label2value"}).
					addInt64Point(1, 4, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1value2", "label2value"}).
					addInt64Point(0, 4, 2).
					build(),
			},
		},
		// filter datapoints
		{
			name: "filter_datapoints",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:         FilterDataPoints,
								DataPointValue: 1,
							},
						},
					},
				},
			},
			in: []*metricspb.Metric{
				metricBuilder().setName("metric").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1value1", "label2value"}).
					addInt64Point(0, 1, 2).
					addTimeseries(1, []string{"label1value2", "label2value"}).
					addInt64Point(1, 0, 2).
					build(),
			},
			out: []*metricspb.Metric{
				metricBuilder().setName("metric").setLabels([]string{"label1", "label2"}).
					setDataType(metricspb.MetricDescriptor_GAUGE_INT64).
					addTimeseries(1, []string{"label1value1", "label2value"}).
					addInt64Point(0, 1, 2).
					build(),
			},
		},
	}
)
