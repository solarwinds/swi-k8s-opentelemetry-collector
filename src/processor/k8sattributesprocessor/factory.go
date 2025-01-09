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

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib
// Changes customizing the original source code: see CHANGELOG.md in deploy/helm directory

package swk8sattributesprocessor // import "github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/xconsumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper"
	"go.opentelemetry.io/collector/processor/xprocessor"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/internal/k8sconfig"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/metadata"
)

var (
	kubeClientProvider   = kube.ClientProvider(nil)
	consumerCapabilities = consumer.Capabilities{MutatesData: true}
	defaultExcludes      = ExcludeConfig{Pods: []ExcludePodConfig{{Name: "jaeger-agent"}, {Name: "jaeger-collector"}}}
)

// NewFactory returns a new factory for the k8s processor.
func NewFactory() processor.Factory {
	return xprocessor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		xprocessor.WithTraces(createTracesProcessor, metadata.TracesStability),
		xprocessor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
		xprocessor.WithLogs(createLogsProcessor, metadata.LogsStability),
		xprocessor.WithProfiles(createProfilesProcessor, metadata.ProfilesStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Exclude:   defaultExcludes,
		Extract: ExtractConfig{
			Metadata: enabledAttributes(),
		},
		WaitForMetadataTimeout: 10 * time.Second,
	}
}

func createTracesProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	next consumer.Traces,
) (processor.Traces, error) {
	return createTracesProcessorWithOptions(ctx, params, cfg, next)
}

func createLogsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
) (processor.Logs, error) {
	return createLogsProcessorWithOptions(ctx, params, cfg, nextLogsConsumer)
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextMetricsConsumer consumer.Metrics,
) (processor.Metrics, error) {
	return createMetricsProcessorWithOptions(ctx, params, cfg, nextMetricsConsumer)
}

func createProfilesProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextProfilesConsumer xconsumer.Profiles,
) (xprocessor.Profiles, error) {
	return createProfilesProcessorWithOptions(ctx, params, cfg, nextProfilesConsumer)
}

func createTracesProcessorWithOptions(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	next consumer.Traces,
	options ...option,
) (processor.Traces, error) {
	kp := createKubernetesProcessor(set, cfg, options...)

	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		next,
		kp.processTraces,
		processorhelper.WithCapabilities(consumerCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createMetricsProcessorWithOptions(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextMetricsConsumer consumer.Metrics,
	options ...option,
) (processor.Metrics, error) {
	kp := createKubernetesProcessor(set, cfg, options...)

	return processorhelper.NewMetrics(
		ctx,
		set,
		cfg,
		nextMetricsConsumer,
		kp.processMetrics,
		processorhelper.WithCapabilities(consumerCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createLogsProcessorWithOptions(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
	options ...option,
) (processor.Logs, error) {
	kp := createKubernetesProcessor(set, cfg, options...)

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextLogsConsumer,
		kp.processLogs,
		processorhelper.WithCapabilities(consumerCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createProfilesProcessorWithOptions(
	ctx context.Context,
	set processor.Settings,
	cfg component.Config,
	nextProfilesConsumer xconsumer.Profiles,
	options ...option,
) (xprocessor.Profiles, error) {
	kp := createKubernetesProcessor(set, cfg, options...)

	return xprocessorhelper.NewProfiles(
		ctx,
		set,
		cfg,
		nextProfilesConsumer,
		kp.processProfiles,
		xprocessorhelper.WithCapabilities(consumerCapabilities),
		xprocessorhelper.WithStart(kp.Start),
		xprocessorhelper.WithShutdown(kp.Shutdown),
	)
}

func createKubernetesProcessor(
	params processor.Settings,
	cfg component.Config,
	options ...option,
) *kubernetesprocessor {
	kp := &kubernetesprocessor{
		logger:            params.Logger,
		cfg:               cfg,
		options:           options,
		telemetrySettings: params.TelemetrySettings,
		resources:         make(map[string]*kubernetesProcessorResource),
	}

	return kp
}

func createProcessorOpts(cfg component.Config) []option {
	oCfg := cfg.(*Config)
	var opts []option
	if oCfg.Passthrough {
		opts = append(opts, withPassthrough())
	}

	if oCfg.SetObjectExistence {
		opts = append(opts, withSetObjectExistence())
	}

	// extraction rules
	opts = append(opts, withExtractMetadata(oCfg.Extract.Metadata...))
	opts = append(opts, withExtractLabels(oCfg.Extract.Labels...))
	opts = append(opts, withExtractAnnotations(oCfg.Extract.Annotations...))

	// filters
	opts = append(opts, withFilterNode(oCfg.Filter.Node, oCfg.Filter.NodeFromEnvVar))
	opts = append(opts, withFilterNamespace(oCfg.Filter.Namespace))
	opts = append(opts, withFilterLabels(oCfg.Filter.Labels...))
	opts = append(opts, withFilterFields(oCfg.Filter.Fields...))
	opts = append(opts, withAPIConfig(oCfg.APIConfig))

	opts = append(opts, withExtractPodAssociations(oCfg.Association...))

	opts = append(opts, withExcludes(oCfg.Exclude))

	opts = append(opts, withWaitForMetadataTimeout(oCfg.WaitForMetadataTimeout))
	if oCfg.WaitForMetadata {
		opts = append(opts, withWaitForMetadata(true))
	}

	opts = append(opts, createDeploymentProcessorOpts(oCfg.Deployment)...)
	opts = append(opts, createStatefulSetProcessorOpts(oCfg.StatefulSet)...)
	opts = append(opts, createReplicaSetProcessorOpts(oCfg.ReplicaSet)...)
	opts = append(opts, createDaemonSetProcessorOpts(oCfg.DaemonSet)...)
	opts = append(opts, createJobProcessorOpts(oCfg.Job)...)
	opts = append(opts, createCronJobProcessorOpts(oCfg.CronJob)...)
	opts = append(opts, createNodeProcessorOpts(oCfg.Node)...)
	opts = append(opts, createPersistentVolumeProcessorOpts(oCfg.PersistentVolume)...)
	opts = append(opts, createPersistentVolumeClaimProcessorOpts(oCfg.PersistentVolumeClaim)...)
	opts = append(opts, createServiceProcessorOpts(oCfg.Service)...)
	return opts
}
