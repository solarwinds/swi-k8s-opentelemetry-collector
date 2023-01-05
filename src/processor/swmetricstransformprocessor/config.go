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

import "go.opentelemetry.io/collector/component"

const (
	// IncludeFieldName is the mapstructure field name for Include field
	IncludeFieldName = "include"

	// MatchTypeFieldName is the mapstructure field name for MatchType field
	MatchTypeFieldName = "match_type"

	// MetricNameFieldName is the mapstructure field name for MetricName field
	MetricNameFieldName = "metric_name"

	// ActionFieldName is the mapstructure field name for Action field
	ActionFieldName = "action"

	// NewNameFieldName is the mapstructure field name for NewName field
	NewNameFieldName = "new_name"

	// LabelFieldName is the mapstructure field name for Label field
	LabelFieldName = "label"

	// NewLabelFieldName is the mapstructure field name for NewLabel field
	NewLabelFieldName = "new_label"

	// NewValueFieldName is the mapstructure field name for NewValue field
	NewValueFieldName = "new_value"

	// ScaleFieldName is the mapstructure field name for Scale field
	ScaleFieldName = "experimental_scale"

	// DataPointsFieldName is the mapstructure field name for Datapoint field
	DataPointsFieldName = "datapoint_value"

	// DataValueActionFieldName is the mapstructure field name for Datapoint field
	DataValueActionFieldName = "datapoint_value_action"
)

// Config defines configuration for Resource processor.
type Config struct {
	// Transform specifies a list of transforms on metrics with each transform focusing on one metric.
	Transforms []Transform `mapstructure:"transforms"`
}

// Transform defines the transformation applied to the specific metric
type Transform struct {

	// --- SPECIFY WHICH METRIC(S) TO MATCH ---

	// MetricIncludeFilter is used to select the metric(s) to operate on.
	// REQUIRED
	MetricIncludeFilter FilterConfig `mapstructure:",squash"`

	// MetricName is used to select the metric to operate on.
	// DEPRECATED. Use MetricIncludeFilter instead.
	MetricName string `mapstructure:"metric_name"`

	// --- SPECIFY THE ACTION TO TAKE ON THE MATCHED METRIC(S) ---

	// Action specifies the action performed on the matched metric. Action specifies
	// if the operations (specified below) are performed on metrics in place (update),
	// on an inserted clone (insert), or on a new combined metric that includes all
	// data points from the set of matching metrics (combine).
	// REQUIRED
	Action ConfigAction `mapstructure:"action"`

	// --- SPECIFY HOW TO TRANSFORM THE METRIC GENERATED AS A RESULT OF APPLYING THE ABOVE ACTION ---

	// NewName specifies the name of the new metric when inserting or updating.
	// REQUIRED only if Action is INSERT.
	NewName string `mapstructure:"new_name"`

	// Operations contains a list of operations that will be performed on the resulting metric(s).
	Operations []Operation `mapstructure:"operations"`
}

type FilterConfig struct {
	// Include specifies the metric(s) to operate on.
	Include string `mapstructure:"include"`

	// MatchType determines how the Include string is matched: <strict|regexp>.
	MatchType MatchType `mapstructure:"match_type"`

	// MatchLabels specifies the label set against which the metric filter will work.
	// This field is optional.
	MatchLabels map[string]string `mapstructure:"experimental_match_labels"`
}

// Operation defines the specific operation performed on the selected metrics.
type Operation struct {
	// Action specifies the action performed for this operation.
	// REQUIRED
	Action OperationAction `mapstructure:"action"`

	// Label identifies the exact label to operate on.
	Label string `mapstructure:"label"`

	// NewLabel determines the name to rename the identified label to.
	NewLabel string `mapstructure:"new_label"`

	// NewValue is used to set a new label value either when the operation is `AggregatedValues` or `AddLabel`.
	NewValue string `mapstructure:"new_value"`

	// ValueActions is a list of renaming actions for label values.
	ValueActions []ValueAction `mapstructure:"value_actions"`

	// LabelValue identifies the exact label value to operate on
	LabelValue string `mapstructure:"label_value"`

	// DataPointValue identifies data point values ​​that should be included/excluded in the output
	DataPointValue float64 `mapstructure:"datapoint_value"`

	// DataPointValueAction is a list of actions for data point values.
	DataPointValueAction DataPointValueAction `mapstructure:"datapoint_value_action"`
}

// ValueAction renames label values.
type ValueAction struct {
	// Value specifies the current label value.
	Value string `mapstructure:"value"`

	// NewValue specifies the label value to rename to.
	NewValue string `mapstructure:"new_value"`
}

// DataPointValueAction is the enum to capture the type of action to perform on a data point.
type DataPointValueAction string

const (
	// Include datapoints that match the provided value
	Include DataPointValueAction = "include"

	// Exclude datapoints that match the provided value
	Exclude DataPointValueAction = "exclude"
)

var dataPointActions = []DataPointValueAction{Include, Exclude}

func (da DataPointValueAction) isValid() bool {
	for _, datapointAction := range dataPointActions {
		if da == datapointAction {
			return true
		}
	}

	return false
}

// ConfigAction is the enum to capture the type of action to perform on a metric.
type ConfigAction string

const (
	// Insert adds a new metric to the batch with a new name.
	Insert ConfigAction = "insert"

	// Update updates an existing metric.
	Update ConfigAction = "update"
)

var actions = []ConfigAction{Insert, Update}

func (ca ConfigAction) isValid() bool {
	for _, configAction := range actions {
		if ca == configAction {
			return true
		}
	}

	return false
}

// OperationAction is the enum to capture the thress types of actions to perform for an operation.
type OperationAction string

const (
	// AddLabel adds a new label to an existing metric.
	AddLabel OperationAction = "add_label"

	// UpdateLabel applies name changes to label and/or label values.
	UpdateLabel OperationAction = "update_label"

	// DeleteLabelValue deletes a label value by also removing all the points associated with this label value
	DeleteLabelValue OperationAction = "delete_label_value"

	// FilterDataPoints deletes all data points that do not contain the specified value
	FilterDataPoints OperationAction = "filter_datapoints"
)

var operationActions = []OperationAction{AddLabel, UpdateLabel, DeleteLabelValue, FilterDataPoints}

func (oa OperationAction) isValid() bool {
	for _, operationAction := range operationActions {
		if oa == operationAction {
			return true
		}
	}

	return false
}

// MatchType is the enum to capture the two types of matching metric(s) that should have operations applied to them.
type MatchType string

const (
	// StrictMatchType is the FilterType for filtering by exact string matches.
	StrictMatchType MatchType = "strict"

	// RegexpMatchType is the FilterType for filtering by regexp string matches.
	RegexpMatchType MatchType = "regexp"
)

var matchTypes = []MatchType{StrictMatchType, RegexpMatchType}

var _ component.Config = (*Config)(nil)

func (mt MatchType) isValid() bool {
	for _, matchType := range matchTypes {
		if mt == matchType {
			return true
		}
	}

	return false
}
