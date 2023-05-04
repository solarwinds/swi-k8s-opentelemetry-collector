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

	"go.opentelemetry.io/collector/pdata/pmetric"
)

type metricsTransformTest struct {
	name       string // test name
	transforms []internalTransform
	in         []pmetric.Metric
	out        []pmetric.Metric
}

var (
	// test cases
	standardTests = []metricsTransformTest{
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric2").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric3").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "new/new/metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric/2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric3").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "value1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "new/label1").
					addIntDatapoint(1, 2, 3, "value1").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "label1-value1").
					addIntDatapoint(1, 2, 3, "label1-value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "new/label1-value1").
					addIntDatapoint(1, 2, 3, "label1-value2").build(),
			},
		},
		{
			name: "metric_label_update_label_and_label_value",
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
							valueActionsMapping: map[string]string{"label1-value1": "new/label1-value1"},
						},
					},
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "label1-value1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "new/label1").
					addIntDatapoint(1, 2, 3, "new/label1-value1").build(),
			},
		},
		{
			name: "metric_label_update_with_regexp_filter",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("^matched.*$")},
					Action:              Update,
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "matched-metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "label1-value1", "label2-value1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "unmatched-metric2", "label1", "label2").
					addIntDatapoint(1, 2, 3, "label1-value1", "label2-value1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "matched-metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "new/label1-value1", "label2-value1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "unmatched-metric2", "label1", "label2").
					addIntDatapoint(1, 2, 3, "label1-value1", "label2-value1").build(),
			},
		},
		// INSERT
		{
			name: "metric_name_insert",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1"},
					Action:              Insert,
					NewName:             "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "metric2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1",
						attrMatchers: map[string]StringMatcher{"label1": strictMatcher("value1")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 3, 2, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 3, 2, "value1", "value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1", "label2").
					addIntDatapoint(1, 3, 2, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile(`(.|\s)*\S(.|\s)*`)}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_two_datapoints_positive",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile("value3")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").
					addIntDatapoint(2, 2, 3, "value3", "value4").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").
					addIntDatapoint(2, 2, 3, "value3", "value4").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1", "label2").
					addIntDatapoint(2, 2, 3, "value3", "value4").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_two_datapoints_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile("value3")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").
					addIntDatapoint(2, 2, 3, "value11", "value22").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").
					addIntDatapoint(2, 2, 3, "value11", "value22").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_with_full_value",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile("value1")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1",
						attrMatchers: map[string]StringMatcher{"label1": strictMatcher("wrong_value")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_negative",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile(".*wrong_ending")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_strict_missing_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric1",
						attrMatchers: map[string]StringMatcher{"missing_key": strictMatcher("value1")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"missing_key": regexp.MustCompile("value1")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_and_present_key",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile("value1"),
							"missing_key": regexp.MustCompile("value2")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
		},
		{
			name: "metric_name_insert_with_match_label_regexp_missing_key_with_empty_expression",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterRegexp{include: regexp.MustCompile("metric1"),
						attrMatchers: map[string]StringMatcher{"label1": regexp.MustCompile("value1"),
							"missing_key": regexp.MustCompile("^$")}},
					Action:  Insert,
					NewName: "new/metric1",
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "new/label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "label1-value1").
					addIntDatapoint(1, 2, 4, "label1-value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1").
					addIntDatapoint(1, 2, 3, "label1-value1").
					addIntDatapoint(1, 2, 4, "label1-value2").build(),
				metricBuilder(pmetric.MetricTypeGauge, "new/metric1", "label1").
					addIntDatapoint(1, 2, 3, "new/label1-value1").
					addIntDatapoint(1, 2, 4, "label1-value2").build(),
			},
		},
		// Add Label to a metric
		{
			name: "update_existing_metric_by_adding_a_new_label_when_there_are_no_labels",
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1").addIntDatapoint(1, 2, 3).build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "foo").addIntDatapoint(1, 2, 3, "bar").build(),
			},
		},
		{
			name: "update_existing_metric_by_adding_a_new_label_when_there_are_labels",
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2", "foo").
					addIntDatapoint(1, 2, 3, "value1", "value2", "bar").build(),
			},
		},
		{
			name: "update_existing_metric_by_adding_a_label_that_is_duplicated_in_the_list",
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric1", "label1", "label2").
					addIntDatapoint(1, 2, 3, "value1", "value2").build(),
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 3, "label1value1", "label2value").
					addIntDatapoint(1, 2, 4, "label1value2", "label2value").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 4, "label1value2", "label2value").build(),
			},
		},
		{
			name: "delete_all_metric_datapoints",
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
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 3, "label1value1", "label2value").build(),
			},
			out: []pmetric.Metric{},
		},
		// filter datapoints
		{
			name: "filter_datapoints_include",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:               FilterDataPoints,
								DataPointValue:       1,
								DataPointValueAction: Include,
							},
						},
					},
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 1, "label1value1", "label2value").
					addIntDatapoint(1, 2, 0, "label1value2", "label2value").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 1, "label1value1", "label2value").build(),
			},
		},
		{
			name: "filter_datapoints_include_flag_datapoints",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:               FilterDataPoints,
								DataPointValue:       1,
								DataPointValueAction: Include,
							},
						},
					},
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addFlagDatapoint(1, 2, 1, "label1value2", "label2value").build(),
			},
			out: []pmetric.Metric{},
		},
		{
			name: "filter_datapoints_exclude",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:               FilterDataPoints,
								DataPointValue:       1,
								DataPointValueAction: Exclude,
							},
						},
					},
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 1, "label1value1", "label2value").
					addIntDatapoint(1, 2, 0, "label1value2", "label2value").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addIntDatapoint(1, 2, 0, "label1value2", "label2value").build(),
			},
		},
		{
			name: "filter_datapoints_exclude_flag_datapoints",
			transforms: []internalTransform{
				{
					MetricIncludeFilter: internalFilterStrict{include: "metric"},
					Action:              Update,
					Operations: []internalOperation{
						{
							configOperation: Operation{
								Action:               FilterDataPoints,
								DataPointValue:       1,
								DataPointValueAction: Exclude,
							},
						},
					},
				},
			},
			in: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addFlagDatapoint(1, 2, 1, "label1value2", "label2value").build(),
			},
			out: []pmetric.Metric{
				metricBuilder(pmetric.MetricTypeGauge, "metric", "label1", "label2").
					addFlagDatapoint(1, 2, 1, "label1value2", "label2value").build(),
			},
		},
	}
)
