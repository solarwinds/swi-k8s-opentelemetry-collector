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
	"context"
	"strings"
	"time"

	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type WatchResourceClient[T KubernetesResource] struct {
	client *WatchClient

	Resources    map[ResourceIdentifier]T
	Rules        ExtractionRulesResource
	Filters      Filters
	Associations []Association
	Exclude      ExcludesResources

	deleteQueue []resourceDeleteRequest
	informer    cache.SharedInformer

	nameConvention               string
	uuidConvention               string
	resourceType                 string
	observabilityTableSizeFunc   func(tableSize int64)
	observabilityResourceAdded   func()
	observabilityResourceUpdated func()
	observabilityResourceDeleted func()
}

// New initializes a new k8s Client.
func NewWatchStatefulSetClient(
	client *WatchClient,
	clientStatefulSet *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientStatefulSet,
		MetadataFromStatefulSet,
		conventions.AttributeK8SStatefulSetName,
		conventions.AttributeK8SStatefulSetUID,
		client.telemetryBuilder.OtelsvcK8sStatefulSetTableSize,
		client.telemetryBuilder.OtelsvcK8sStatefulSetAdded,
		client.telemetryBuilder.OtelsvcK8sStatefulSetUpdated,
		client.telemetryBuilder.OtelsvcK8sStatefulSetDeleted,
		newStatefulSetSharedInformer,
	)
}

// New initializes a new k8s Client.
func NewWatchDeploymentClient(
	client *WatchClient,
	clientDeployment *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientDeployment,
		MetadataFromDeployment,
		conventions.AttributeK8SDeploymentName,
		conventions.AttributeK8SDeploymentUID,
		client.telemetryBuilder.OtelsvcK8sDeploymentTableSize,
		client.telemetryBuilder.OtelsvcK8sDeploymentAdded,
		client.telemetryBuilder.OtelsvcK8sDeploymentUpdated,
		client.telemetryBuilder.OtelsvcK8sDeploymentDeleted,
		newDeploymentSharedInformer,
	)
}

func NewWatchReplicaSetClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromReplicaSet,
		conventions.AttributeK8SReplicaSetName,
		conventions.AttributeK8SReplicaSetUID,
		client.telemetryBuilder.OtelsvcK8sReplicasetTableSize,
		client.telemetryBuilder.OtelsvcK8sReplicasetAdded,
		client.telemetryBuilder.OtelsvcK8sReplicasetUpdated,
		client.telemetryBuilder.OtelsvcK8sReplicasetDeleted,
		newReplicaSetSharedInformer,
	)
}

func NewWatchDaemonSetClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromDaemonSet,
		conventions.AttributeK8SDaemonSetName,
		conventions.AttributeK8SDaemonSetUID,
		client.telemetryBuilder.OtelsvcK8sDaemonSetTableSize,
		client.telemetryBuilder.OtelsvcK8sDaemonSetAdded,
		client.telemetryBuilder.OtelsvcK8sDaemonSetUpdated,
		client.telemetryBuilder.OtelsvcK8sDaemonSetDeleted,
		newDaemonSetSharedInformer,
	)
}

func NewWatchJobClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromJob,
		conventions.AttributeK8SJobName,
		conventions.AttributeK8SJobUID,
		client.telemetryBuilder.OtelsvcK8sJobTableSize,
		client.telemetryBuilder.OtelsvcK8sJobAdded,
		client.telemetryBuilder.OtelsvcK8sJobUpdated,
		client.telemetryBuilder.OtelsvcK8sJobDeleted,
		newJobSharedInformer,
	)
}

func NewWatchCronJobClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromCronJob,
		conventions.AttributeK8SCronJobName,
		conventions.AttributeK8SCronJobUID,
		client.telemetryBuilder.OtelsvcK8sCronJobTableSize,
		client.telemetryBuilder.OtelsvcK8sCronJobAdded,
		client.telemetryBuilder.OtelsvcK8sCronJobUpdated,
		client.telemetryBuilder.OtelsvcK8sCronJobDeleted,
		newCronJobSharedInformer,
	)
}

func NewWatchNodeClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromNode,
		conventions.AttributeK8SNodeName,
		conventions.AttributeK8SNodeUID,
		client.telemetryBuilder.OtelsvcK8sNodeTableSize,
		client.telemetryBuilder.OtelsvcK8sNodeAdded,
		client.telemetryBuilder.OtelsvcK8sNodeUpdated,
		client.telemetryBuilder.OtelsvcK8sNodeDeleted,
		newNodeSharedInformer,
	)
}

func NewWatchPersistentVolumeClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromPersistentVolume,
		"k8s.persistentvolume.name",
		"k8s.persistentvolume.uid",
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeTableSize,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeAdded,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeUpdated,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeDeleted,
		newPersistentVolumeSharedInformer,
	)
}

func NewWatchPersistentVolumeClaimClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromPersistentVolumeClaim,
		"k8s.persistentvolumeclaim.name",
		"k8s.persistentvolumeclaim.uid",
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeClaimTableSize,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeClaimAdded,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeClaimUpdated,
		client.telemetryBuilder.OtelsvcK8sPersistentVolumeClaimDeleted,
		newPersistentVolumeClaimSharedInformer,
	)
}

func NewWatchServiceClient(
	client *WatchClient,
	clientResource *ClientResource) (*WatchResourceClient[KubernetesResource], error) {
	return NewWatchResourceClient[KubernetesResource](
		client,
		clientResource,
		MetadataFromService,
		"k8s.service.name",
		"k8s.service.uid",
		client.telemetryBuilder.OtelsvcK8sServiceTableSize,
		client.telemetryBuilder.OtelsvcK8sServiceAdded,
		client.telemetryBuilder.OtelsvcK8sServiceUpdated,
		client.telemetryBuilder.OtelsvcK8sServiceDeleted,
		newServiceSharedInformer,
	)
}

// New initializes a new k8s Client.
func NewWatchResourceClient[T KubernetesResource](
	client *WatchClient,
	clientResource *ClientResource,
	resourceType string,
	nameConvention string,
	uuidConvention string,
	observabilityTableSize metric.Int64Gauge,
	observabilityResourceAdded metric.Int64Counter,
	observabilityResourceUpdated metric.Int64Counter,
	observabilityResourceDeleted metric.Int64Counter,
	informerProvider InformerProvider) (*WatchResourceClient[T], error) {
	c := &WatchResourceClient[T]{
		client: client,

		Rules:        clientResource.ExtractionRules,
		Filters:      clientResource.Filters,
		Associations: clientResource.Associations,
		Exclude:      clientResource.Excludes,

		nameConvention: nameConvention,
		uuidConvention: uuidConvention,
		resourceType:   resourceType,
		observabilityTableSizeFunc: func(tableSize int64) {
			observabilityTableSize.Record(context.Background(), tableSize)
		},
		observabilityResourceAdded:   func() { observabilityResourceAdded.Add(context.Background(), 1) },
		observabilityResourceUpdated: func() { observabilityResourceUpdated.Add(context.Background(), 1) },
		observabilityResourceDeleted: func() { observabilityResourceDeleted.Add(context.Background(), 1) },
	}
	go c.deleteLoop(time.Second*30, defaultPodDeleteGracePeriod)

	c.Resources = map[ResourceIdentifier]T{}

	if clientResource.Informer == nil {
		clientResource.Informer = informerProvider
	}

	labelSelector, fieldSelector, err := selectorsFromFilters(c.Filters)
	if err != nil {
		return nil, err
	}
	c.client.logger.Info(
		"k8s filtering",
		zap.String("labelSelector", labelSelector.String()),
		zap.String("fieldSelector", fieldSelector.String()),
	)
	if c.extractResourceLabelsAnnotations(resourceType) {
		c.informer = clientResource.Informer(c.client.kc, c.Filters.Namespace, labelSelector, fieldSelector)
		err = c.informer.SetTransform(
			func(object any) (any, error) {
				originalResource, success := object.(metav1.Object)
				if !success {
					return object, nil
				}

				return removeUnnecessaryResourceData(originalResource, c.Rules), nil
			},
		)
	} else {
		c.informer = NewNoOpInformer(c.client.kc)
	}

	return c, err
}

// Start registers pod event handlers and starts watching the kubernetes cluster for pod changes.
func (c *WatchResourceClient[T]) Start() (reg cache.ResourceEventHandlerRegistration, err error) {
	reg, err = c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleResourceAdd,
		UpdateFunc: c.handleResourceUpdate,
		DeleteFunc: c.handleResourceDelete,
	})
	if err == nil {
		go c.informer.Run(c.client.stopCh)
	}
	return
}

func (c *WatchResourceClient[T]) handleResourceAdd(obj any) {
	c.observabilityResourceAdded()
	if resource, ok := obj.(metav1.Object); ok {
		c.addOrUpdateResource(resource)
	} else {
		c.client.logger.Error("object received was not of type metav1.Object", zap.Any("received", obj))
	}
	resourceTableSize := len(c.Resources)
	c.observabilityTableSizeFunc(int64(resourceTableSize))
}

func (c *WatchResourceClient[T]) handleResourceUpdate(_, new any) {
	c.observabilityResourceUpdated()
	if resource, ok := new.(metav1.Object); ok {
		c.addOrUpdateResource(resource)
	} else {
		c.client.logger.Error("object received was not of type metav1.Object", zap.Any("received", new))
	}
	resourceTableSize := len(c.Resources)
	c.observabilityTableSizeFunc(int64(resourceTableSize))
}

func (c *WatchResourceClient[T]) handleResourceDelete(obj any) {
	c.observabilityResourceDeleted()
	if resource, ok := ignoreDeletedFinalStateUnknown(obj).(metav1.Object); ok {
		c.forgetResource(resource)
	} else {
		c.client.logger.Error("object received was not of type metav1.Object", zap.Any("received", obj))
	}
	resourceTableSize := len(c.Resources)
	c.observabilityTableSizeFunc(int64(resourceTableSize))
}

func (c *WatchResourceClient[T]) extractResourceAttributes(resource metav1.Object) map[string]string {
	tags := map[string]string{}

	if c.Rules.UID {
		uid := resource.GetUID()
		tags[c.uuidConvention] = string(uid)
	}

	for _, r := range c.Rules.Labels {
		r.extractFromResourceMetadata(c.resourceType, resource.GetLabels(), tags, "k8s."+c.resourceType+".labels.%s")
	}

	for _, r := range c.Rules.Annotations {
		r.extractFromResourceMetadata(c.resourceType, resource.GetAnnotations(), tags, "k8s."+c.resourceType+".annotations.%s")
	}

	return tags
}

func (c *WatchResourceClient[T]) resourceFromAPI(resource metav1.Object) KubernetesResource {
	timeStamp := resource.GetCreationTimestamp()
	newResource := &Resource{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
		UID:       string(resource.GetUID()),
		StartTime: &timeStamp,
	}

	if c.shouldIgnoreResource(resource) {
		newResource.Ignore = true
	} else {
		newResource.Attributes = c.extractResourceAttributes(resource)
	}

	return newResource
}

// getIdentifiersFromAssocDeployment returns a list of ResourceIdentifier for the given deployment
func (c *WatchResourceClient[T]) getIdentifiersFromAssocResource(resource KubernetesResource) []ResourceIdentifier {
	var ids []ResourceIdentifier
	for _, assoc := range c.Associations {
		ret := ResourceIdentifier{}
		skip := false
		for i, source := range assoc.Sources {
			// If association configured to take specific attribute from the deployment
			if source.From == ResourceSource {
				attr := ""
				switch source.Name {
				case conventions.AttributeK8SNamespaceName:
					attr = resource.GetNamespace()
				case c.nameConvention:
					attr = resource.GetName()
				case c.uuidConvention:
					attr = resource.GetUID()
				default:
					if v, ok := resource.GetAttributes()[source.Name]; ok {
						attr = v
					}
				}

				if attr == "" {
					skip = true
					break
				}
				ret[i] = PodIdentifierAttributeFromSource(source, attr)
			}
		}

		if !skip {
			ids = append(ids, ret)
		}
	}

	return ids
}

func (c *WatchResourceClient[T]) addOrUpdateResource(resource metav1.Object) {
	newResource := c.resourceFromAPI(resource)

	c.client.m.Lock()
	defer c.client.m.Unlock()

	for _, id := range c.getIdentifiersFromAssocResource(newResource) {
		// Compare the initial creation timestamp for existing deployment and new deployment with the same identifier
		// and only replace the old deployment if the creation time of the new deployment is newer or equal.
		// This should handle the case where a new deployment with the same attributes is created,
		// but the update event for the old deployment comes in later.
		if d, ok := c.Resources[id]; ok {
			timeStamp := resource.GetCreationTimestamp()
			if timeStamp.Before(d.GetStartTime()) {
				return
			}
		}
		c.Resources[id] = newResource.(T)
	}
}

func (c *WatchResourceClient[T]) forgetResource(resource metav1.Object) {
	deploymentToRemove := c.resourceFromAPI(resource)

	for _, id := range c.getIdentifiersFromAssocResource(deploymentToRemove) {
		if d, ok := c.Resources[id]; ok && d.GetName() == resource.GetName() {
			p, ok := c.GetResource(id)

			if ok && p.GetName() == resource.GetName() {
				c.appendResourceDeleteQueue(id, resource.GetName())
			}
		}
	}
}

func (c *WatchResourceClient[T]) shouldIgnoreResource(resource metav1.Object) bool {
	// Check if user requested the resource to be ignored through annotations
	if v, ok := resource.GetAnnotations()[ignoreAnnotation]; ok {
		if strings.ToLower(strings.TrimSpace(v)) == "true" {
			return true
		}
	}

	// Check if user requested the resource to be ignored through configuration
	for _, excludedResource := range c.Exclude.Resources {
		if excludedResource.Name.MatchString(resource.GetName()) {
			return true
		}
	}

	return false
}

func (c *WatchResourceClient[T]) deleteLoop(
	interval time.Duration,
	gracePeriod time.Duration) {
	// This loop runs after N seconds and deletes deployment from cache.
	// It iterates over the delete queue and deletes all that aren't
	// in the grace period anymore.
	for {
		select {
		case <-time.After(interval):
			var cutoff int
			now := time.Now()
			c.client.deleteMut.Lock()
			for i, d := range c.deleteQueue {
				if d.ts.Add(gracePeriod).After(now) {
					break
				}
				cutoff = i + 1
			}
			toDelete := c.deleteQueue[:cutoff]
			c.deleteQueue = c.deleteQueue[cutoff:]
			c.client.deleteMut.Unlock()

			c.client.m.Lock()
			for _, d := range toDelete {
				if p, ok := c.Resources[d.id]; ok {
					// Sanity check: make sure we are deleting the same deployment
					// and the underlying state (ip<>pod mapping) has not changed.
					if p.GetName() == d.resourceName {
						delete(c.Resources, d.id)
					}
				}
			}
			tableSize := len(c.Resources)
			c.observabilityTableSizeFunc(int64(tableSize))
			c.client.m.Unlock()

		case <-c.client.stopCh:
			return
		}
	}
}

// GetDeployment returns the deployment identifier.
func (c *WatchResourceClient[T]) GetResource(identifier ResourceIdentifier) (KubernetesResource, bool) {
	c.client.m.RLock()
	resource, ok := c.Resources[identifier]
	c.client.m.RUnlock()
	if ok {
		return resource, ok
	}
	return nil, false
}

func (c *WatchResourceClient[T]) appendResourceDeleteQueue(deploymentID ResourceIdentifier, resourceName string) {
	c.client.deleteMut.Lock()
	c.deleteQueue = append(c.deleteQueue, resourceDeleteRequest{
		id:           deploymentID,
		resourceName: resourceName,
		ts:           time.Now(),
	})
	c.client.deleteMut.Unlock()
}

func (c *WatchResourceClient[T]) extractResourceLabelsAnnotations(resourceType string) bool {
	for _, r := range c.Rules.Labels {
		if r.From == resourceType {
			return true
		}
	}

	for _, r := range c.Rules.Annotations {
		if r.From == resourceType {
			return true
		}
	}

	return false
}

// This function removes all data from resource except what is required by extraction rules
func removeUnnecessaryResourceData(resource metav1.Object, rules ExtractionRulesResource) metav1.Object {
	transformedResource := metav1.ObjectMeta{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}

	if rules.UID {
		transformedResource.SetUID(resource.GetUID())
	}

	if len(rules.Labels) > 0 {
		transformedResource.Labels = resource.GetLabels()
	}

	if len(rules.Annotations) > 0 {
		transformedResource.Annotations = resource.GetAnnotations()
	}

	transformedResource.SetOwnerReferences(resource.GetOwnerReferences())
	return &transformedResource
}
