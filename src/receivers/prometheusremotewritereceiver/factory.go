// Copyright The OpenTelemetry Authors
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

package prometheusremotewritereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"

	"go.opentelemetry.io/collector/config/confighttp"
)

const (
	// The value of "type" key in configuration.
	typeStr = "prometheusremotewrite"

	defaultHTTPEndpoint = "0.0.0.0:4318"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Metrics Generation processor.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithMetricsReceiver(createMetricsReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		Protocols: Protocols{
			HTTP: &confighttp.HTTPServerSettings{
				Endpoint: defaultHTTPEndpoint,
			},
		},
	}
}

func createMetricsReceiver(
	ctx context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	nextConsumer consumer.Metrics) (component.MetricsReceiver, error) {
	c := cfg.(*Config)
	r, err := newMetricsReceiver(ctx, c, params, nextConsumer)
	if err != nil {
		return nil, err
	}

	if err := r.(*prometheusRemoteWriteReceiver).registerMetricsConsumer(nextConsumer); err != nil {
		return nil, err
	}
	return r, nil
}

/*
// buildInternalConfig constructs the internal metric generation rules
func buildInternalConfig(config *Config) []internalRule {
	internalRules := make([]internalRule, len(config.Rules))

	for i, rule := range config.Rules {
		customRule := internalRule{
			name:      rule.Name,
			unit:      rule.Unit,
			ruleType:  string(rule.Type),
			metric1:   rule.Metric1,
			metric2:   rule.Metric2,
			operation: string(rule.Operation),
			scaleBy:   rule.ScaleBy,
		}
		internalRules[i] = customRule
	}
	return internalRules
}*/
