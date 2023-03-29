// Copyright 2020 OpenTelemetry Authors
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

package k8sattributesprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"

func createDeploymentProcessorOpts(deploymentConfig DeploymentConfig) []option {
	var opts []option

	opts = append(opts, withDeployment(deploymentConfig))
	opts = append(opts, withExtractMetadataDeployment(deploymentConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsDeployment(deploymentConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsDeployment(deploymentConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceDeployment(deploymentConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsDeployment(deploymentConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsDeployment(deploymentConfig.Filter.Fields...))
	opts = append(opts, withExtractDeploymentAssociations(deploymentConfig.Association...))
	opts = append(opts, withExcludesDeployment(deploymentConfig.Exclude))

	return opts
}
