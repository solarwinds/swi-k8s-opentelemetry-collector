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

package kube // import "github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"

import (
	"regexp"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MetadataFromDeployment is used to specify to extract metadata/labels/annotations from deployment
	MetadataFromDeployment = "deployment"
	// MetadataFromStatefulSet is used to specify to extract metadata/labels/annotations from statefulset
	MetadataFromStatefulSet = "statefulset"
	// MetadataFromReplicaSet is used to specify to extract metadata/labels/annotations from replicaset
	MetadataFromReplicaSet = "replicaset"
	// MetadataFromDaemonSet is used to specify to extract metadata/labels/annotations from daemonset
	MetadataFromDaemonSet = "daemonset"
	// MetadataFromJob is used to specify to extract metadata/labels/annotations from job
	MetadataFromJob = "job"
	// MetadataFromCronJob is used to specify to extract metadata/labels/annotations from cronjob
	MetadataFromCronJob = "cronjob"
	// MetadataFromNode is used to specify to extract metadata/labels/annotations from node
	MetadataFromNode = "node"
	// MetadataFromPersistentVolume is used to specify to extract metadata/labels/annotations from persistent volume
	MetadataFromPersistentVolume = "persistentvolume"
	// MetadataFromPersistentVolumeClaim is used to specify to extract metadata/labels/annotations from persistent volume claim
	MetadataFromPersistentVolumeClaim = "persistentvolumeclaim"
	// MetadataFromService is used to specify to extract metadata/labels/annotations from service
	MetadataFromService = "service"
)

// ClientResource is a generic client for Kubernetes resources
// It is used to fetch information about resources from the API server
// and to watch for changes in the resources
// R is the type of the extraction rules
type ClientResource struct {
	ExtractionRules ExtractionRulesResource
	Excludes        ExcludesResources
	Filters         Filters
	Associations    []Association
	Informer        InformerProvider
}

// ResourceIdentifier is a custom type to represent resource identification identification
type ResourceIdentifier [PodIdentifierMaxLength]PodIdentifierAttribute

// IsNotEmpty checks if PodIdentifier is empty or not
func (d *ResourceIdentifier) IsNotEmpty() bool {
	return d[0].Source.From != ""
}

type ResourceAttributes interface {
	GetName() string
	GetNamespace() string
	GetUID() string
	GetAttributes() map[string]string
	GetIgnore() bool
}

type KubernetesResource interface {
	ResourceAttributes
	GetStartTime() *metav1.Time
	GetDeletedAt() time.Time
}

type Resource struct {
	Name       string
	Namespace  string
	UID        string
	Attributes map[string]string
	Ignore     bool
	StartTime  *metav1.Time
	DeletedAt  time.Time
}

func (r Resource) GetName() string {
	return r.Name
}

func (r Resource) GetNamespace() string {
	return r.Namespace
}

func (r Resource) GetUID() string {
	return r.UID
}

func (r Resource) GetAttributes() map[string]string {
	return r.Attributes
}

func (r Resource) GetIgnore() bool {
	return r.Ignore
}

func (d Resource) GetStartTime() *metav1.Time {
	return d.StartTime
}

func (d Resource) GetDeletedAt() time.Time {
	return d.DeletedAt
}

type resourceDeleteRequest struct {
	// id is identifier of the resource to remove from the map
	id ResourceIdentifier
	// name contains name of the resource to remove from the map
	resourceName string
	ts           time.Time
}

// ExtractionRulesResource contains common fields for extraction rules of various resources.
type ExtractionRulesResource struct {
	Annotations []FieldExtractionRule
	Labels      []FieldExtractionRule

	UID bool
}

func (r *FieldExtractionRule) extractFromResourceMetadata(fromType string, metadata map[string]string, tags map[string]string, formatter string) {
	if r.From == fromType {
		r.extractFromMetadata(metadata, tags, formatter)
	}
}

// ExcludeResources represent a Deployment name to ignore
type ExcludeResources struct {
	Name *regexp.Regexp
}

// ExcludesResources represent a list of Deployments to ignore
type ExcludesResources struct {
	Resources []ExcludeResources
}
