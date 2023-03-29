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

import (
	"context"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

type kubernetesProcessorDeployment struct {
	kubernetesprocessor *kubernetesprocessor
	rules               kube.ExtractionRulesDeployment
	filters             kube.Filters
	associations        []kube.Association
	ignore              kube.ExcludesDeployments
}

func (kp *kubernetesProcessorDeployment) isEmpty() bool {
	return len(kp.rules.Annotations) == 0 &&
		len(kp.rules.Labels) == 0 &&
		!kp.rules.DeploymentUID
}

func (kp *kubernetesprocessor) processResourceDeployment(ctx context.Context, resource pcommon.Resource) {
	deploymentIdentifierValue := extractDeploymentID(ctx, resource.Attributes(), kp.deployment.associations)
	kp.logger.Debug("evaluating deployment identifier", zap.Any("value", deploymentIdentifierValue))
	if deploymentIdentifierValue.IsNotEmpty() {
		if deployment, ok := kp.kc.GetDeployment(deploymentIdentifierValue); ok {
			kp.logger.Debug("getting the deployment", zap.Any("deployment", deployment))

			for key, val := range deployment.Attributes {
				if _, found := resource.Attributes().Get(key); !found {
					resource.Attributes().PutStr(key, val)
				}
			}
		}
	}
}

// extractDeploymentID returns pod identifier for first association matching all sources
func extractDeploymentID(ctx context.Context, attrs pcommon.Map, associations []kube.Association) kube.DeploymentIdentifier {
	for _, asso := range associations {
		skip := false

		ret := kube.DeploymentIdentifier{}
		for i, source := range asso.Sources {
			switch {
			case source.From == kube.ResourceSource:
				// Extract values based on configured resource_attribute.
				attributeValue := stringAttributeFromMap(attrs, source.Name)
				if attributeValue == "" {
					skip = true
					break
				}

				ret[i] = kube.PodIdentifierAttributeFromSource(source, attributeValue)
			}
		}

		// If all association sources has been resolved, return result
		if !skip {
			return ret
		}
	}
	return kube.DeploymentIdentifier{}
}

func (kp *kubernetesprocessor) getClientDeployment(deployment *kubernetesProcessorDeployment) *kube.ClientDeployment {
	if deployment != nil && !deployment.isEmpty() {
		return &kube.ClientDeployment{
			ExtractionRules: deployment.rules,
			Excludes:        deployment.ignore,
			Filters:         deployment.filters,
			Associations:    deployment.associations,
			Informer:        nil,
		}
	} else {
		return nil
	}
}
