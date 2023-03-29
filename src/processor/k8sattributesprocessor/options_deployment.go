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
	"fmt"
	"regexp"

	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
)

// withExtractMetadataDeployment allows specifying options to control extraction of deployment metadata.
// If no fields explicitly provided, all metadata extracted by default.
func withExtractMetadataDeployment(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		if len(fields) == 0 {
			fields = []string{}
		}
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SDeploymentUID:
				p.deployment.rules.DeploymentUID = true
			default:
				return fmt.Errorf("\"%s\" is not a supported metadata field", field)
			}
		}
		return nil
	}
}

// withExtractLabelsDeployment allows specifying options to control extraction of deployment labels.
func withExtractLabelsDeployment(labels ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		labels, err := extractFieldRules("labels", labels...)
		if err != nil {
			return err
		}
		p.deployment.rules.Labels = labels
		return nil
	}
}

// withExtractAnnotationsDeployment allows specifying options to control extraction of deployment annotations tags.
func withExtractAnnotationsDeployment(annotations ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		annotations, err := extractFieldRules("annotations", annotations...)
		if err != nil {
			return err
		}
		p.deployment.rules.Annotations = annotations
		return nil
	}
}

// withFilterNamespaceDeployment allows specifying options to control filtering deployment by a namespace.
func withFilterNamespaceDeployment(ns string) option {
	return func(p *kubernetesprocessor) error {
		p.deployment.filters.Namespace = ns
		return nil
	}
}

// withFilterLabelsDeployment allows specifying options to control filtering deployment by labels.
func withFilterLabelsDeployment(filters ...FieldFilterConfig) option {
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
		p.deployment.filters.Labels = labels
		return nil
	}
}

// withFilterFieldsDeployment allows specifying options to control filtering deployment by fields.
func withFilterFieldsDeployment(filters ...FieldFilterConfig) option {
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
			default:
				return fmt.Errorf("'%s' is not a valid field filter operation for key=%s, value=%s", f.Op, f.Key, f.Value)
			}
			fields = append(fields, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.filters.Fields = fields
		return nil
	}
}

// withExtractDeploymentAssociations allows specifying options to associate pod metadata with incoming resource
func withExtractDeploymentAssociations(deploymentAssociations ...AssociationConfig) option {
	return func(p *kubernetesprocessor) error {
		associations := make([]kube.Association, 0, len(deploymentAssociations))
		var assoc kube.Association
		for _, association := range deploymentAssociations {
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
		p.deployment.associations = associations
		return nil
	}
}

// withDeployment allows specifying options to associate deployment metadata with incoming resource
func withDeployment(deployment DeploymentConfig) option {
	return func(p *kubernetesprocessor) error {
		p.deployment = &kubernetesProcessorDeployment{
			kubernetesprocessor: p,
			rules:               kube.ExtractionRulesDeployment{},
			filters:             kube.Filters{},
			associations:        []kube.Association{},
		}

		return nil
	}
}

// withExcludesDeployment allows specifying deployment to exclude
func withExcludesDeployment(podExclude ExcludeDeploymentConfig) option {
	return func(p *kubernetesprocessor) error {
		ignoredNames := kube.ExcludesDeployments{}
		names := podExclude.Deployments

		for _, name := range names {
			ignoredNames.Deployments = append(ignoredNames.Deployments, kube.ExcludeDeployments{Name: regexp.MustCompile(name.Name)})
		}
		p.deployment.ignore = ignoredNames
		return nil
	}
}
