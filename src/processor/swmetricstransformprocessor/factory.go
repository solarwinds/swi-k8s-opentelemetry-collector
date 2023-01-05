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
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "swmetricstransform"
	// The stability level of the processor.
	stability = component.StabilityLevelBeta
)

var consumerCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Metrics Transform processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability))
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	oCfg := cfg.(*Config)
	if err := validateConfiguration(oCfg); err != nil {
		return nil, err
	}

	hCfg, err := buildHelperConfig(oCfg, params.BuildInfo.Version)
	if err != nil {
		return nil, err
	}
	metricsProcessor := newSwMetricsTransformProcessor(params.Logger, hCfg)

	return processorhelper.NewMetricsProcessor(
		ctx,
		params,
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities))
}

// validateConfiguration validates the input configuration has all of the required fields for the processor
// An error is returned if there are any invalid inputs.
func validateConfiguration(config *Config) error {
	for _, transform := range config.Transforms {
		if transform.MetricIncludeFilter.Include == "" && transform.MetricName == "" {
			return fmt.Errorf("missing required field %q", IncludeFieldName)
		}

		if transform.MetricIncludeFilter.Include != "" && transform.MetricName != "" {
			return fmt.Errorf("cannot supply both %q and %q, use %q with %q match type", IncludeFieldName, MetricNameFieldName, IncludeFieldName, StrictMatchType)
		}

		if transform.MetricIncludeFilter.MatchType != "" && !transform.MetricIncludeFilter.MatchType.isValid() {
			return fmt.Errorf("%q must be in %q", MatchTypeFieldName, matchTypes)
		}

		if transform.MetricIncludeFilter.MatchType == RegexpMatchType {
			_, err := regexp.Compile(transform.MetricIncludeFilter.Include)
			if err != nil {
				return fmt.Errorf("%q, %w", IncludeFieldName, err)
			}
		}

		if !transform.Action.isValid() {
			return fmt.Errorf("%q must be in %q", ActionFieldName, actions)
		}

		if transform.Action == Insert && transform.NewName == "" {
			return fmt.Errorf("missing required field %q while %q is %v", NewNameFieldName, ActionFieldName, Insert)
		}

		for i, op := range transform.Operations {
			if !op.Action.isValid() {
				return fmt.Errorf("operation %v: %q must be in %q", i+1, ActionFieldName, operationActions)
			}

			if op.Action == UpdateLabel && op.Label == "" {
				return fmt.Errorf("operation %v: missing required field %q while %q is %v", i+1, LabelFieldName, ActionFieldName, UpdateLabel)
			}
			if op.Action == AddLabel && op.NewLabel == "" {
				return fmt.Errorf("operation %v: missing required field %q while %q is %v", i+1, NewLabelFieldName, ActionFieldName, AddLabel)
			}
			if op.Action == AddLabel && op.NewValue == "" {
				return fmt.Errorf("operation %v: missing required field %q while %q is %v", i+1, NewValueFieldName, ActionFieldName, AddLabel)
			}
			if op.Action == FilterDataPoints && !op.DataPointValueAction.isValid() {
				return fmt.Errorf("operation %v: %q must be in %q", i+1, DataValueActionFieldName, dataPointActions)
			}
		}
	}
	return nil
}

// buildHelperConfig constructs the maps that will be useful for the operations
func buildHelperConfig(config *Config, version string) ([]internalTransform, error) {
	helperDataTransforms := make([]internalTransform, len(config.Transforms))
	for i, t := range config.Transforms {

		// for backwards compatibility, convert metric name to an include filter
		if t.MetricName != "" {
			t.MetricIncludeFilter = FilterConfig{Include: t.MetricName}
			t.MetricName = ""
		}
		if t.MetricIncludeFilter.MatchType == "" {
			t.MetricIncludeFilter.MatchType = StrictMatchType
		}

		filter, err := createFilter(t.MetricIncludeFilter)
		if err != nil {
			return nil, err
		}

		helperT := internalTransform{
			MetricIncludeFilter: filter,
			Action:              t.Action,
			NewName:             t.NewName,
			Operations:          make([]internalOperation, len(t.Operations)),
		}

		for j, op := range t.Operations {
			op.NewValue = strings.ReplaceAll(op.NewValue, "{{version}}", version)

			mtpOp := internalOperation{
				configOperation: op,
			}
			if len(op.ValueActions) > 0 {
				mtpOp.valueActionsMapping = createLabelValueMapping(op.ValueActions, version)
			}

			helperT.Operations[j] = mtpOp
		}
		helperDataTransforms[i] = helperT
	}
	return helperDataTransforms, nil
}

func createFilter(filterConfig FilterConfig) (internalFilter, error) {
	switch filterConfig.MatchType {
	case StrictMatchType:
		matchers, err := getMatcherMap(filterConfig.MatchLabels, func(str string) (StringMatcher, error) { return strictMatcher(str), nil })
		if err != nil {
			return nil, err
		}
		return internalFilterStrict{include: filterConfig.Include, attrMatchers: matchers}, nil
	case RegexpMatchType:
		matchers, err := getMatcherMap(filterConfig.MatchLabels, func(str string) (StringMatcher, error) { return regexp.Compile(str) })
		if err != nil {
			return nil, err
		}
		return internalFilterRegexp{include: regexp.MustCompile(filterConfig.Include), attrMatchers: matchers}, nil
	}

	return nil, fmt.Errorf("invalid match type: %v", filterConfig.MatchType)
}

// createLabelValueMapping creates the labelValue rename mappings based on the valueActions
func createLabelValueMapping(valueActions []ValueAction, version string) map[string]string {
	mapping := make(map[string]string)
	for i := 0; i < len(valueActions); i++ {
		valueActions[i].NewValue = strings.ReplaceAll(valueActions[i].NewValue, "{{version}}", version)
		mapping[valueActions[i].Value] = valueActions[i].NewValue
	}
	return mapping
}

func getMatcherMap(strMap map[string]string, ctor func(string) (StringMatcher, error)) (map[string]StringMatcher, error) {
	out := make(map[string]StringMatcher)
	for k, v := range strMap {
		matcher, err := ctor(v)
		if err != nil {
			return nil, err
		}
		out[k] = matcher
	}
	return out, nil
}
