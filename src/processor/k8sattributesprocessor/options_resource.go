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
	"fmt"
	"regexp"

	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
)

// withExtractMetadataDeployment allows specifying options to control extraction of resource metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataDeployment(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SDeploymentUID:
				p.resources[kube.MetadataFromDeployment].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataStatefulSet allows specifying options to control extraction of statefulset metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataStatefulSet(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SStatefulSetUID:
				p.resources[kube.MetadataFromStatefulSet].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataReplicaSet allows specifying options to control extraction of replicaset metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataReplicaSet(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SReplicaSetUID:
				p.resources[kube.MetadataFromReplicaSet].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataDaemonSet allows specifying options to control extraction of daemonset metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataDaemonSet(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SDaemonSetUID:
				p.resources[kube.MetadataFromDaemonSet].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataJob allows specifying options to control extraction of job metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataJob(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SJobUID:
				p.resources[kube.MetadataFromJob].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataCronJob allows specifying options to control extraction of cronjob metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataCronJob(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SCronJobUID:
				p.resources[kube.MetadataFromCronJob].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractMetadataNode allows specifying options to control extraction of node metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataNode(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SNodeUID:
				p.resources[kube.MetadataFromNode].rules.UID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractResourceAssociations allows specifying options to associate pod metadata with incoming resource
func withExtractResourceAssociations(resourceType string, resourceAssociations ...AssociationConfig) option {
	return func(p *kubernetesprocessor) error {
		associations := make([]kube.Association, 0, len(resourceAssociations))
		var assoc kube.Association
		for _, association := range resourceAssociations {
			assoc = kube.Association{
				Sources: []kube.AssociationSource{},
			}

			var name string

			if association.From != "" {
				if association.From == kube.ConnectionSource {
					name = ""
				} else {
					name = association.Name
				}
				assoc.Sources = append(assoc.Sources, kube.AssociationSource{
					From: association.From,
					Name: name,
				})
			} else {
				for _, associationSource := range association.Sources {
					if associationSource.From == kube.ConnectionSource {
						name = ""
					} else {
						name = associationSource.Name
					}
					assoc.Sources = append(assoc.Sources, kube.AssociationSource{
						From: associationSource.From,
						Name: name,
					})
				}
			}
			associations = append(associations, assoc)
		}
		p.resources[resourceType].associations = associations
		return nil
	}
}

// withResource allows specifying options to associate resource metadata with incoming resource
func withResource(resourceType string) option {
	return func(p *kubernetesprocessor) error {
		p.resources[resourceType] = &kubernetesProcessorResource{
			kubernetesprocessor: p,
			rules:               kube.ExtractionRulesResource{},
			filters:             kube.Filters{},
			associations:        []kube.Association{},
		}

		return nil
	}
}

// withExcludesDeployment allows specifying deployment to exclude
func withExcludesResource(resourceType string, exclude []ExcludePodConfig) option {
	return func(p *kubernetesprocessor) error {
		ignoredNames := kube.ExcludesResources{}
		names := exclude

		for _, name := range names {
			ignoredNames.Resources = append(ignoredNames.Resources, kube.ExcludeResources{Name: regexp.MustCompile(name.Name)})
		}
		p.resources[resourceType].ignore = ignoredNames
		return nil
	}
}

func withExtractLabelsGeneric(resourceType string, labels ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		labels, err := extractFieldRules("labels", labels...)
		if err != nil {
			return err
		}
		p.resources[resourceType].rules.Labels = labels
		return nil
	}
}

func withExtractAnnotationsGeneric(resourceType string, annotations ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		annotations, err := extractFieldRules("annotations", annotations...)
		if err != nil {
			return err
		}
		p.resources[resourceType].rules.Annotations = annotations
		return nil
	}
}

func withFilterNamespaceGeneric(resourceType string, ns string) option {
	return func(p *kubernetesprocessor) error {
		p.resources[resourceType].filters.Namespace = ns
		return nil
	}
}

func withFilterLabelsGeneric(resourceType string, filters ...FieldFilterConfig) option {
	return func(p *kubernetesprocessor) error {
		var labels []kube.FieldFilter
		for _, f := range filters {
			if f.Op == "" {
				f.Op = filterOPEquals
			}

			var op selection.Operator
			switch f.Op {
			case filterOPEquals:
				op = selection.Equals
			case filterOPNotEquals:
				op = selection.NotEquals
			case filterOPExists:
				op = selection.Exists
			case filterOPDoesNotExist:
				op = selection.DoesNotExist
			default:
				return fmt.Errorf("'%s' is not a valid label filter operation for key=%s, value=%s", f.Op, f.Key, f.Value)
			}
			labels = append(labels, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.resources[resourceType].filters.Labels = labels
		return nil
	}
}

func withFilterFieldsGeneric(resourceType string, filters ...FieldFilterConfig) option {
	return func(p *kubernetesprocessor) error {
		var fields []kube.FieldFilter
		for _, f := range filters {
			if f.Op == "" {
				f.Op = filterOPEquals
			}
			var op selection.Operator
			switch f.Op {
			case filterOPEquals:
				op = selection.Equals
			case filterOPNotEquals:
				op = selection.NotEquals
			case filterOPExists:
				op = selection.Exists
			case filterOPDoesNotExist:
				op = selection.DoesNotExist
			default:
				return fmt.Errorf("'%s' is not a valid field filter operation for key=%s, value=%s", f.Op, f.Key, f.Value)
			}
			fields = append(fields, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.resources[resourceType].filters.Fields = fields
		return nil
	}
}

func withExtractAssociationsGeneric(resourceType string, inputAssociations ...AssociationConfig) option {
	return func(p *kubernetesprocessor) error {
		associations := make([]kube.Association, 0, len(inputAssociations))
		var assoc kube.Association
		for _, association := range inputAssociations {
			assoc = kube.Association{
				Sources: []kube.AssociationSource{},
			}

			var name string

			if association.From != "" {
				if association.From == kube.ConnectionSource {
					name = ""
				} else {
					name = association.Name
				}
				assoc.Sources = append(assoc.Sources, kube.AssociationSource{
					From: association.From,
					Name: name,
				})
			} else {
				for _, associationSource := range association.Sources {
					if associationSource.From == kube.ConnectionSource {
						name = ""
					} else {
						name = associationSource.Name
					}
					assoc.Sources = append(assoc.Sources, kube.AssociationSource{
						From: associationSource.From,
						Name: name,
					})
				}
			}
			associations = append(associations, assoc)
		}

		p.resources[resourceType].associations = associations
		return nil
	}
}
