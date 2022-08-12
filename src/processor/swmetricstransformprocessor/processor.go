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
	"regexp"

	agentmetricspb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	internaldata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus"
)

type swMetricsTransformProcessor struct {
	transforms []internalTransform
	logger     *zap.Logger
}

type internalTransform struct {
	MetricIncludeFilter internalFilter
	Action              ConfigAction
	NewName             string
	Operations          []internalOperation
}

type internalOperation struct {
	configOperation     Operation
	valueActionsMapping map[string]string
	labelSetMap         map[string]bool
}

type internalFilter interface {
	getMatches(toMatch metricNameMapping) []*match
	getSubexpNames() []string
}

type match struct {
	metric     *metricspb.Metric
	pattern    *regexp.Regexp
	submatches []int
}

type StringMatcher interface {
	MatchString(string) bool
}

type strictMatcher string

func (s strictMatcher) MatchString(cmp string) bool {
	return string(s) == cmp
}

type internalFilterStrict struct {
	include     string
	matchLabels map[string]StringMatcher
}

func (f internalFilterStrict) getMatches(toMatch metricNameMapping) []*match {

	if metrics, ok := toMatch[f.include]; ok {
		matches := make([]*match, 0)
		for _, metric := range metrics {
			matchedMetric := labelMatched(f.matchLabels, metric)
			if matchedMetric != nil {
				matches = append(matches, &match{metric: matchedMetric})
			}
		}
		return matches
	}

	return nil
}

func (f internalFilterStrict) getSubexpNames() []string {
	return nil
}

type internalFilterRegexp struct {
	include     *regexp.Regexp
	matchLabels map[string]StringMatcher
}

func (f internalFilterRegexp) getMatches(toMatch metricNameMapping) []*match {
	matches := make([]*match, 0)
	for name, metrics := range toMatch {
		if submatches := f.include.FindStringSubmatchIndex(name); submatches != nil {
			for _, metric := range metrics {
				matchedMetric := labelMatched(f.matchLabels, metric)
				if matchedMetric != nil {
					matches = append(matches, &match{metric: matchedMetric, pattern: f.include, submatches: submatches})
				}
			}
		}
	}
	return matches
}

func (f internalFilterRegexp) getSubexpNames() []string {
	return f.include.SubexpNames()
}

func labelMatched(matchLabels map[string]StringMatcher, metric *metricspb.Metric) *metricspb.Metric {
	if len(matchLabels) == 0 {
		return metric
	}

	metricWithMatchedLabel := &metricspb.Metric{}
	metricWithMatchedLabel.MetricDescriptor = proto.Clone(metric.MetricDescriptor).(*metricspb.MetricDescriptor)
	metricWithMatchedLabel.Resource = proto.Clone(metric.Resource).(*resourcepb.Resource)

	var timeSeriesWithMatchedLabel []*metricspb.TimeSeries
	labelIndexValueMap := make(map[int]StringMatcher)

	for key, value := range matchLabels {
		keyFound := false

		for idx, label := range metric.MetricDescriptor.LabelKeys {
			if label.Key != key {
				continue
			}

			keyFound = true
			labelIndexValueMap[idx] = value
		}

		// if a label-key is not found then return nil only if the given label-value is non-empty. If a given label-value is empty
		// and the key is not found then move forward. In this approach we can make sure certain key is not present which is a valid use case.
		if !keyFound && !value.MatchString("") {
			return nil
		}
	}

	for _, timeseries := range metric.Timeseries {
		allValuesMatched := true
		for index, value := range labelIndexValueMap {
			if !value.MatchString(timeseries.LabelValues[index].Value) {
				allValuesMatched = false
				break
			}
		}
		if allValuesMatched {
			timeSeriesWithMatchedLabel = append(timeSeriesWithMatchedLabel, timeseries)
		}
	}

	if len(timeSeriesWithMatchedLabel) == 0 {
		return nil
	}

	metricWithMatchedLabel.Timeseries = timeSeriesWithMatchedLabel
	return metricWithMatchedLabel
}

type metricNameMapping map[string][]*metricspb.Metric

func newMetricNameMapping(metrics []*metricspb.Metric) metricNameMapping {
	mnm := metricNameMapping(make(map[string][]*metricspb.Metric, len(metrics)))
	for _, m := range metrics {
		mnm.add(m.MetricDescriptor.Name, m)
	}
	return mnm
}

func (mnm metricNameMapping) add(name string, metrics ...*metricspb.Metric) {
	mnm[name] = append(mnm[name], metrics...)
}

func (mnm metricNameMapping) remove(name string, metrics ...*metricspb.Metric) {
	for _, metric := range metrics {
		for j, m := range mnm[name] {
			if metric == m {
				mnm[name] = append(mnm[name][:j], mnm[name][j+1:]...)
				break
			}
		}
	}
}

func newSwMetricsTransformProcessor(logger *zap.Logger, internalTransforms []internalTransform) *swMetricsTransformProcessor {
	return &swMetricsTransformProcessor{
		transforms: internalTransforms,
		logger:     logger,
	}
}

// processMetrics implements the ProcessMetricsFunc type.
func (mtp *swMetricsTransformProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	rms := md.ResourceMetrics()
	groupedMds := make([]*agentmetricspb.ExportMetricsServiceRequest, 0)

	out := pmetric.NewMetrics()

	for i := 0; i < rms.Len(); i++ {
		node, resource, metrics := internaldata.ResourceMetricsToOC(rms.At(i))

		nameToMetricMapping := newMetricNameMapping(metrics)
		for _, transform := range mtp.transforms {
			matchedMetrics := transform.MetricIncludeFilter.getMatches(nameToMetricMapping)

			for _, match := range matchedMetrics {
				metricName := match.metric.MetricDescriptor.Name

				if transform.Action == Insert {
					match.metric = proto.Clone(match.metric).(*metricspb.Metric)
					metrics = append(metrics, match.metric)
				}

				mtp.update(match, transform)

				if transform.NewName != "" {
					if transform.Action == Update {
						nameToMetricMapping.remove(metricName, match.metric)
					}
					nameToMetricMapping.add(match.metric.MetricDescriptor.Name, match.metric)
				}
			}
		}

		internaldata.OCToMetrics(node, resource, metrics).ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	for i := range groupedMds {
		internaldata.OCToMetrics(groupedMds[i].Node, groupedMds[i].Resource, groupedMds[i].Metrics).ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	return out, nil
}

// update updates the metric content based on operations indicated in transform.
func (mtp *swMetricsTransformProcessor) update(match *match, transform internalTransform) {
	if transform.NewName != "" {
		if match.pattern == nil {
			match.metric.MetricDescriptor.Name = transform.NewName
		} else {
			match.metric.MetricDescriptor.Name = string(match.pattern.ExpandString([]byte{}, transform.NewName, match.metric.MetricDescriptor.Name, match.submatches))
		}
	}

	for _, op := range transform.Operations {
		switch op.configOperation.Action {
		case UpdateLabel:
			mtp.updateLabelOp(match.metric, op)
		case AddLabel:
			mtp.addLabelOp(match.metric, op)
		case DeleteLabelValue:
			mtp.deleteLabelValueOp(match.metric, op)
		case FilterDataPoints:
			mtp.filterDataPoints(match.metric, op)
		}
	}
}
