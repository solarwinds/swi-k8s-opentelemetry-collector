// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
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

package containerprocessor

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"
)

var (
	Type                 = component.MustNewType("containerprocessor")
	consumerCapabilities = consumer.Capabilities{MutatesData: false}
)

func NewFactory() processor.Factory {
	return xprocessor.NewFactory(
		Type,
		createDefaultConfig,
		xprocessor.WithLogs(createLogsProcessor, component.StabilityLevelBeta),
	)
}

type Config struct {
}

func createDefaultConfig() component.Config {
	return Config{}
}

func createLogsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
) (processor.Logs, error) {
	cp := &containerprocessor{
		logger:            params.Logger,
		cfg:               cfg,
		telemetrySettings: params.TelemetrySettings,
	}

	return processorhelper.NewLogs(
		ctx,
		params,
		createDefaultConfig(),
		nextLogsConsumer,
		cp.processLogs,
		processorhelper.WithCapabilities(consumerCapabilities),
		processorhelper.WithStart(cp.Start),
		processorhelper.WithShutdown(cp.Shutdown))
}
