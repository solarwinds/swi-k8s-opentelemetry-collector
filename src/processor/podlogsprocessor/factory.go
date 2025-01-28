package podlogsprocessor

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"
)

var (
	Type                 = component.MustNewType("podlogsprocessor")
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
	return createLogsProcessorWithOptions(ctx, params, cfg, nextLogsConsumer)
}

func createLogsProcessorWithOptions(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
) (processor.Logs, error) {
	kp := createKubernetesProcessor(set, cfg)

	return processorhelper.NewLogs(
		ctx,
		set,
		createDefaultConfig(),
		nextLogsConsumer,
		kp.processLogs,
		processorhelper.WithCapabilities(consumerCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createKubernetesProcessor(
	params processor.Settings,
	cfg component.Config,
) *containerprocessor {
	kp := &containerprocessor{
		logger:            params.Logger,
		cfg:               cfg,
		telemetrySettings: params.TelemetrySettings,
	}

	return kp
}
