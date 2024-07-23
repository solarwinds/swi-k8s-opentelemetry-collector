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
	"net"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/internal/coreinternal/clientutil"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"
)

// extractPodIds returns pod identifier for first association matching all sources
func extractPodID(ctx context.Context, attrs pcommon.Map, associations []kube.Association) kube.PodIdentifier {
	// If pod association is not set
	if len(associations) == 0 {
		return extractPodIDNoAssociations(ctx, attrs)
	}

	connectionIP := clientutil.Address(client.FromContext(ctx))
	for _, asso := range associations {
		skip := false

		ret := kube.PodIdentifier{}
		for i, source := range asso.Sources {
			// If association configured to take IP address from connection
			switch {
			case source.From == kube.ConnectionSource:
				if connectionIP == "" {
					skip = true
					break
				}
				ret[i] = kube.PodIdentifierAttributeFromConnection(connectionIP)
			case source.From == kube.ResourceSource:
				// Extract values based on configured resource_attribute.
				attributeValue := stringAttributeFromMap(attrs, source.Name)
				if attributeValue == "" {
					skip = true
					break
				}

				// If association configured by resource_attribute
				// In k8s environment, host.name label set to a pod IP address.
				// If the value doesn't represent an IP address, we skip it.
				if asso.Name == conventions.AttributeHostName && net.ParseIP(attributeValue) == nil {
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
	return kube.PodIdentifier{}
}

// extractPodIds returns pod identifier for first association matching all sources
func extractPodIDNoAssociations(ctx context.Context, attrs pcommon.Map) kube.PodIdentifier {
	var podIP, labelIP string
	podIP = stringAttributeFromMap(attrs, kube.K8sIPLabelName)
	if podIP != "" {
		return kube.PodIdentifier{
			kube.PodIdentifierAttributeFromConnection(podIP),
		}
	}

	labelIP = stringAttributeFromMap(attrs, clientIPLabelName)
	if labelIP != "" {
		return kube.PodIdentifier{
			kube.PodIdentifierAttributeFromConnection(labelIP),
		}
	}

	connectionIP := clientutil.Address(client.FromContext(ctx))
	if connectionIP != "" {
		return kube.PodIdentifier{
			kube.PodIdentifierAttributeFromConnection(connectionIP),
		}
	}

	hostname := stringAttributeFromMap(attrs, conventions.AttributeHostName)
	if net.ParseIP(hostname) != nil {
		return kube.PodIdentifier{
			kube.PodIdentifierAttributeFromConnection(hostname),
		}
	}

	return kube.PodIdentifier{}
}

func stringAttributeFromMap(attrs pcommon.Map, key string) string {
	if val, ok := attrs.Get(key); ok {
		if val.Type() == pcommon.ValueTypeStr {
			return val.Str()
		}
	}
	return ""
}
