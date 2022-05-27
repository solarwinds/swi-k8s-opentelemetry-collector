package prometheustypeconverterprocessor

import (
	"context"

	agentmetricspb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	internaldata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus"
)

type prometheusTypeConverterProcessor struct {
	transforms []internalTransform
	logger     *zap.Logger
}

type internalTransform struct {
	MetricIncludeFilter internalFilter
	ConvertType         ConvertType
}

type internalFilter interface {
	getMatches(toMatch metricNameMapping) []*match
	getSubexpNames() []string
}

type match struct {
	metric *metricspb.Metric
}

type StringMatcher interface {
	MatchString(string) bool
}

type internalFilterStrict struct {
	include string
}

func (f internalFilterStrict) getMatches(toMatch metricNameMapping) []*match {
	if metrics, ok := toMatch[f.include]; ok {
		matches := make([]*match, 0)
		for _, metric := range metrics {
			if metric != nil {
				matches = append(matches, &match{metric: metric})
			}
		}
		return matches
	}

	return nil
}

func (f internalFilterStrict) getSubexpNames() []string {
	return nil
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

func newPrometheusTypeConverterProcessor(logger *zap.Logger, internalTransforms []internalTransform) *prometheusTypeConverterProcessor {
	return &prometheusTypeConverterProcessor{
		transforms: internalTransforms,
		logger:     logger,
	}
}

// processMetrics implements the ProcessMetricsFunc type.
func (mtp *prometheusTypeConverterProcessor) processMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	rms := md.ResourceMetrics()
	groupedMds := make([]*agentmetricspb.ExportMetricsServiceRequest, 0)

	out := pmetric.NewMetrics()

	for i := 0; i < rms.Len(); i++ {
		node, resource, metrics := internaldata.ResourceMetricsToOC(rms.At(i))

		nameToMetricMapping := newMetricNameMapping(metrics)
		for _, transform := range mtp.transforms {
			matchedMetrics := transform.MetricIncludeFilter.getMatches(nameToMetricMapping)

			for _, match := range matchedMetrics {
				mtp.update(match, transform)
			}
		}

		internaldata.OCToMetrics(node, resource, metrics).ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	for i := range groupedMds {
		internaldata.OCToMetrics(groupedMds[i].Node, groupedMds[i].Resource, groupedMds[i].Metrics).ResourceMetrics().MoveAndAppendTo(out.ResourceMetrics())
	}

	return out, nil
}

func (mtp *prometheusTypeConverterProcessor) update(match *match, transform internalTransform) {
	switch match.metric.MetricDescriptor.Type {
	case metricspb.MetricDescriptor_GAUGE_INT64:
		match.metric.MetricDescriptor.Type = metricspb.MetricDescriptor_CUMULATIVE_INT64
	case metricspb.MetricDescriptor_GAUGE_DOUBLE:
		match.metric.MetricDescriptor.Type = metricspb.MetricDescriptor_CUMULATIVE_DOUBLE
	}
}
