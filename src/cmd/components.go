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

package main

import (
	healthcheckextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	k8sobserver "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver"
	filestorage "github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"
	attributesprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	cumulativetodeltaprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	deltatorateprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	filterprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	groupbyattrsprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"
	k8sattributesprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	metricsgenerationprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor"
	metricstransformprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	resourcedetectionprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	resourceprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	transformprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	filelogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	journaldreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"
	k8seventsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	k8sobjectsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver"
	prometheusreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	receivercreator "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	simpleprometheusreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver"
	windowseventlogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	forwardconnector "go.opentelemetry.io/collector/connector/forwardconnector"
	"go.opentelemetry.io/collector/exporter"
	otlpexporter "go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	batchprocessor "go.opentelemetry.io/collector/processor/batchprocessor"
	memorylimiterprocessor "go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	otlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
		filestorage.NewFactory(),
		k8sobserver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.ExtensionModules = make(map[component.Type]string, len(factories.Extensions))
	factories.ExtensionModules[healthcheckextension.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.116.0"
	factories.ExtensionModules[filestorage.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.116.0"
	factories.ExtensionModules[k8sobserver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.116.0"

	factories.Receivers, err = receiver.MakeFactoryMap(
		prometheusreceiver.NewFactory(),
		k8seventsreceiver.NewFactory(),
		k8sobjectsreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		journaldreceiver.NewFactory(),
		windowseventlogreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		receivercreator.NewFactory(),
		simpleprometheusreceiver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.ReceiverModules = make(map[component.Type]string, len(factories.Receivers))
	factories.ReceiverModules[prometheusreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.116.0"
	factories.ReceiverModules[k8seventsreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.116.0"
	factories.ReceiverModules[k8sobjectsreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.116.0"
	factories.ReceiverModules[filelogreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.116.0"
	factories.ReceiverModules[journaldreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.116.0"
	factories.ReceiverModules[windowseventlogreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.116.0"
	factories.ReceiverModules[otlpreceiver.NewFactory().Type()] = "go.opentelemetry.io/collector/receiver/otlpreceiver v0.116.0"
	factories.ReceiverModules[receivercreator.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.116.0"
	factories.ReceiverModules[simpleprometheusreceiver.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.116.0"

	factories.Exporters, err = exporter.MakeFactoryMap(
		otlpexporter.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.ExporterModules = make(map[component.Type]string, len(factories.Exporters))
	factories.ExporterModules[otlpexporter.NewFactory().Type()] = "go.opentelemetry.io/collector/exporter/otlpexporter v0.116.0"

	factories.Processors, err = processor.MakeFactoryMap(
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		groupbyattrsprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		deltatorateprocessor.NewFactory(),
		cumulativetodeltaprocessor.NewFactory(),
		metricsgenerationprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		k8sattributesprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.ProcessorModules = make(map[component.Type]string, len(factories.Processors))
	factories.ProcessorModules[batchprocessor.NewFactory().Type()] = "go.opentelemetry.io/collector/processor/batchprocessor v0.116.0"
	factories.ProcessorModules[memorylimiterprocessor.NewFactory().Type()] = "go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.116.0"
	factories.ProcessorModules[metricstransformprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.116.0"
	factories.ProcessorModules[transformprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.116.0"
	factories.ProcessorModules[groupbyattrsprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.116.0"
	factories.ProcessorModules[resourceprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.116.0"
	factories.ProcessorModules[deltatorateprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor v0.116.0"
	factories.ProcessorModules[cumulativetodeltaprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.116.0"
	factories.ProcessorModules[metricsgenerationprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor v0.116.0"
	factories.ProcessorModules[filterprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.116.0"
	factories.ProcessorModules[attributesprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.116.0"
	factories.ProcessorModules[k8sattributesprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.116.0"
	factories.ProcessorModules[resourcedetectionprocessor.NewFactory().Type()] = "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.116.0"

	factories.Connectors, err = connector.MakeFactoryMap(
		forwardconnector.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.ConnectorModules = make(map[component.Type]string, len(factories.Connectors))
	factories.ConnectorModules[forwardconnector.NewFactory().Type()] = "go.opentelemetry.io/collector/connector/forwardconnector v0.116.0"

	return factories, nil
}
