package prometheustypeconverterprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "prometheustypeconvert"
)

var consumerCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Metrics Transform processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		typeStr,
		createDefaultConfig,
		component.WithMetricsProcessor(createMetricsProcessor))
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
	}
}

func createMetricsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	oCfg := cfg.(*Config)
	if err := validateConfiguration(oCfg); err != nil {
		return nil, err
	}

	hCfg, err := buildHelperConfig(oCfg, params.BuildInfo.Version)
	if err != nil {
		return nil, err
	}
	metricsProcessor := newPrometheusTypeConverterProcessor(params.Logger, hCfg)

	return processorhelper.NewMetricsProcessor(
		cfg,
		nextConsumer,
		metricsProcessor.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities))
}

// validateConfiguration validates the input configuration has all of the required fields for the processor
// An error is returned if there are any invalid inputs.
func validateConfiguration(config *Config) error {
	for _, transform := range config.Transforms {
		if transform.MetricIncludeFilter.Include == "" {
			return fmt.Errorf("missing required field %q", IncludeFieldName)
		}

		if transform.ConvertType != "" && !transform.ConvertType.isValid() {
			return fmt.Errorf("%q must be in %q", ConvertTypeFieldName, convertTypes)
		}
	}
	return nil
}

// buildHelperConfig constructs the maps that will be useful for the operations
func buildHelperConfig(config *Config, version string) ([]internalTransform, error) {
	helperDataTransforms := make([]internalTransform, len(config.Transforms))
	for i, t := range config.Transforms {
		if t.ConvertType == "" {
			t.ConvertType = SumConvertType
		}

		filter, err := createFilter(t.MetricIncludeFilter)
		if err != nil {
			return nil, err
		}

		helperT := internalTransform{
			MetricIncludeFilter: filter,
			ConvertType:         t.ConvertType,
		}

		helperDataTransforms[i] = helperT
	}
	return helperDataTransforms, nil
}

func createFilter(filterConfig FilterConfig) (internalFilter, error) {
	return internalFilterStrict{include: filterConfig.Include}, nil
}
