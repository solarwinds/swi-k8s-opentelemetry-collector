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

package k8sattributesprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

// ClientResource is a generic client for Kubernetes resources
// It is used to fetch information about resources from the API server
// and to watch for changes in the resources
// R is the type of the extraction rules
type kubernetesProcessorResource struct {
	kubernetesprocessor *kubernetesprocessor
	rules               kube.ExtractionRulesResource
	filters             kube.Filters
	associations        []kube.Association
	ignore              kube.ExcludesResources
}

func (kp *kubernetesProcessorResource) isEmpty() bool {
	return len(kp.rules.Annotations) == 0 &&
		len(kp.rules.Labels) == 0 &&
		!kp.rules.UID
}

func processGenericResource(
	kp *kubernetesprocessor,
	resourceType string,
	associations []kube.Association,
	ctx context.Context,
	resource pcommon.Resource) {
	identifierValue := extractIdentifier(ctx, resource.Attributes(), associations)
	kp.logger.Debug(fmt.Sprintf("evaluating %s identifier", resourceType), zap.Any("value", identifierValue))
	if identifierValue.IsNotEmpty() {
		if k8sResource, ok := kp.kc.GetResource(resourceType, identifierValue); ok {
			kp.logger.Debug(fmt.Sprintf("getting the %s", resourceType), zap.Any(resourceType, k8sResource))

			for key, val := range k8sResource.GetAttributes() {
				if _, found := resource.Attributes().Get(key); !found {
					resource.Attributes().PutStr(key, val)
				}
			}

			if kp.setObjectExistence {
				// add attribute indicating that the resource was found
				resource.Attributes().PutStr(fmt.Sprintf("sw.k8s.%s.found", resourceType), "true")
			}
		} else {
			if kp.setObjectExistence {
				// add attribute indicating that the resource was not found
				resource.Attributes().PutStr(fmt.Sprintf("sw.k8s.%s.found", resourceType), "false")
			}
		}
	}
}

func (kp *kubernetesprocessor) getClientResource(resource *kubernetesProcessorResource) *kube.ClientResource {
	if resource != nil && !resource.isEmpty() {
		return &kube.ClientResource{
			ExtractionRules: resource.rules,
			Excludes:        resource.ignore,
			Filters:         resource.filters,
			Associations:    resource.associations,
			Informer:        nil,
		}
	} else {
		return nil
	}
}

func extractIdentifier(ctx context.Context, attrs pcommon.Map, associations []kube.Association) kube.ResourceIdentifier {
	for _, asso := range associations {
		skip := false

		ret := kube.ResourceIdentifier{}
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

		// If all association sources have been resolved, return result
		if !skip {
			return ret
		}
	}
	return kube.ResourceIdentifier{}
}
