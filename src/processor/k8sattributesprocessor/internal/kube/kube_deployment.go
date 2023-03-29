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

package kube // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"

import (
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MetadataFromDeployment is used to specify to extract metadata/labels/annotations from deployment
	MetadataFromDeployment = "deployment"
)

type ClientDeployment struct {
	ExtractionRules ExtractionRulesDeployment
	Excludes        ExcludesDeployments
	Filters         Filters
	Associations    []Association
	Informer        InformerProviderDeployment
}

// DeploymentIdentifier is a custom type to represent Pod identification
type DeploymentIdentifier [PodIdentifierMaxLength]PodIdentifierAttribute

// IsNotEmpty checks if PodIdentifier is empty or not
func (d *DeploymentIdentifier) IsNotEmpty() bool {
	return d[0].Source.From != ""
}

type Deployment struct {
	Name          string
	Namespace     string
	DeploymentUID string
	Attributes    map[string]string
	StartTime     *metav1.Time
	Ignore        bool
	DeletedAt     time.Time
}

type deploymentDeleteRequest struct {
	// id is identifier (Deployment UID) of deployment to remove from deployment map
	id DeploymentIdentifier
	// name contains name of deployment to remove from deployment map
	deploymentName string
	ts             time.Time
}

// ExtractionRulesDeployment is used to specify the information that needs to be extracted
// from deployments and added to the spans as tags.
type ExtractionRulesDeployment struct {
	DeploymentUID bool

	Annotations []FieldExtractionRule
	Labels      []FieldExtractionRule
}

func (r *FieldExtractionRule) extractFromDeploymentMetadata(metadata map[string]string, tags map[string]string, formatter string) {
	if r.From == MetadataFromDeployment {
		r.extractFromMetadata(metadata, tags, formatter)
	}
}

// ExcludeDeployments represent a Deployment name to ignore
type ExcludeDeployments struct {
	Name *regexp.Regexp
}

// ExcludesDeployments represent a list of Deployments to ignore
type ExcludesDeployments struct {
	Deployments []ExcludeDeployments
}
