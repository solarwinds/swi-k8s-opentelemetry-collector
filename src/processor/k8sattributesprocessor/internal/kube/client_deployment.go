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
	"strings"
	"time"

	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/observability"
)

// WatchDeploymentClient is child interface of WatchClient, responsible for watching Deployments.
type WatchDeploymentClient struct {
	client *WatchClient

	Deployments  map[DeploymentIdentifier]*Deployment
	Rules        ExtractionRulesDeployment
	Filters      Filters
	Associations []Association
	Exclude      ExcludesDeployments

	deleteQueue []deploymentDeleteRequest
	informer    cache.SharedInformer
}

// New initializes a new k8s Client.
func NewWatchDeploymentClient(
	client *WatchClient,
	clientDeployment *ClientDeployment) (*WatchDeploymentClient, error) {
	c := &WatchDeploymentClient{
		client: client,

		Rules:        clientDeployment.ExtractionRules,
		Filters:      clientDeployment.Filters,
		Associations: clientDeployment.Associations,
		Exclude:      clientDeployment.Excludes,
	}
	go c.deleteLoop(time.Second*30, defaultPodDeleteGracePeriod)

	c.Deployments = map[DeploymentIdentifier]*Deployment{}

	if clientDeployment.Informer == nil {
		clientDeployment.Informer = newDeploymentSharedInformer
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
	if c.extractDeploymentLabelsAnnotations() {
		c.informer = clientDeployment.Informer(c.client.kc, c.Filters.Namespace, labelSelector, fieldSelector)
	} else {
		c.informer = NewNoOpInformer(c.client.kc)
	}

	return c, err
}

// Start registers pod event handlers and starts watching the kubernetes cluster for pod changes.
func (c *WatchDeploymentClient) Start() {
	_, err := c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleDeploymentAdd,
		UpdateFunc: c.handleDeploymentUpdate,
		DeleteFunc: c.handleDeploymentDelete,
	})
	if err != nil {
		c.client.logger.Error("error adding event handler to deployment informer", zap.Error(err))
	}
	go c.informer.Run(c.client.stopCh)
}

func (c *WatchDeploymentClient) handleDeploymentAdd(obj interface{}) {
	observability.RecordDeploymentAdded()
	if deployment, ok := obj.(*appsv1.Deployment); ok {
		c.addOrUpdateDeployment(deployment)
	} else {
		c.client.logger.Error("object received was not of type appsv1.Deployment", zap.Any("received", obj))
	}
	deploymentTableSize := len(c.Deployments)
	observability.RecordDeploymentTableSize(int64(deploymentTableSize))
}

func (c *WatchDeploymentClient) handleDeploymentUpdate(old, new interface{}) {
	observability.RecordDeploymentUpdated()
	if deployment, ok := new.(*appsv1.Deployment); ok {
		c.addOrUpdateDeployment(deployment)
	} else {
		c.client.logger.Error("object received was not of type appsv1.Deployment", zap.Any("received", new))
	}
	deploymentTableSize := len(c.Deployments)
	observability.RecordDeploymentTableSize(int64(deploymentTableSize))
}

func (c *WatchDeploymentClient) handleDeploymentDelete(obj interface{}) {
	observability.RecordDeploymentDeleted()
	if deployment, ok := obj.(*appsv1.Deployment); ok {
		c.forgetDeployment(deployment)
	} else {
		c.client.logger.Error("object received was not of type appsv1.Deployment", zap.Any("received", obj))
	}
	deploymentTableSize := len(c.Deployments)
	observability.RecordDeploymentTableSize(int64(deploymentTableSize))
}

func (c *WatchDeploymentClient) deleteLoop(interval time.Duration, gracePeriod time.Duration) {
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
				if p, ok := c.Deployments[d.id]; ok {
					// Sanity check: make sure we are deleting the same deployment
					// and the underlying state (ip<>pod mapping) has not changed.
					if p.Name == d.deploymentName {
						delete(c.Deployments, d.id)
					}
				}
			}
			deploymentTableSize := len(c.Deployments)
			observability.RecordDeploymentTableSize(int64(deploymentTableSize))
			c.client.m.Unlock()

		case <-c.client.stopCh:
			return
		}
	}
}

// GetDeployment returns the deployment identifier.
func (c *WatchDeploymentClient) GetDeployment(identifier DeploymentIdentifier) (*Deployment, bool) {
	c.client.m.RLock()
	deployment, ok := c.Deployments[identifier]
	c.client.m.RUnlock()
	if ok {
		return deployment, ok
	}
	return nil, false
}

func (c *WatchDeploymentClient) extractDeploymentAttributes(deployment *appsv1.Deployment) map[string]string {
	tags := map[string]string{}

	if c.Rules.DeploymentUID {
		uid := deployment.GetUID()
		tags[conventions.AttributeK8SDeploymentUID] = string(uid)
	}

	for _, r := range c.Rules.Labels {
		r.extractFromDeploymentMetadata(deployment.Labels, tags, "k8s.deployment.labels.%s")
	}

	for _, r := range c.Rules.Annotations {
		r.extractFromDeploymentMetadata(deployment.Annotations, tags, "k8s.deployment.annotations.%s")
	}

	return tags
}

func (c *WatchDeploymentClient) deploymentFromAPI(deployment *appsv1.Deployment) *Deployment {
	newDeployment := &Deployment{
		Name:          deployment.Name,
		Namespace:     deployment.GetNamespace(),
		StartTime:     &deployment.CreationTimestamp,
		DeploymentUID: string(deployment.UID),
	}

	if c.shouldIgnoreDeployment(deployment) {
		newDeployment.Ignore = true
	} else {
		newDeployment.Attributes = c.extractDeploymentAttributes(deployment)
	}

	return newDeployment
}

// getIdentifiersFromAssocDeployment returns a list of DeploymentIdentifiers for the given deployment
func (c *WatchDeploymentClient) getIdentifiersFromAssocDeployment(deployment *Deployment) []DeploymentIdentifier {
	var ids []DeploymentIdentifier
	for _, assoc := range c.Associations {
		ret := DeploymentIdentifier{}
		skip := false
		for i, source := range assoc.Sources {
			// If association configured to take specific attribute from the deployment
			if source.From == ResourceSource {
				attr := ""
				switch source.Name {
				case conventions.AttributeK8SNamespaceName:
					attr = deployment.Namespace
				case conventions.AttributeK8SDeploymentName:
					attr = deployment.Name
				case conventions.AttributeK8SDeploymentUID:
					attr = deployment.DeploymentUID
				default:
					if v, ok := deployment.Attributes[source.Name]; ok {
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

func (c *WatchDeploymentClient) addOrUpdateDeployment(deployment *appsv1.Deployment) {
	newDeployment := c.deploymentFromAPI(deployment)

	c.client.m.Lock()
	defer c.client.m.Unlock()

	for _, id := range c.getIdentifiersFromAssocDeployment(newDeployment) {
		// Compare the initial creation timestamp for existing deployment and new deployment with the same identifier
		// and only replace the old deployment if the creation time of the new deployment is newer or equal.
		// This should handle the case where a new deployment with the same attributes is created,
		// but the update event for the old deployment comes in later.
		if d, ok := c.Deployments[id]; ok {
			if deployment.CreationTimestamp.Before(d.StartTime) {
				return
			}
		}
		c.Deployments[id] = newDeployment
	}
}

func (c *WatchDeploymentClient) forgetDeployment(deployment *appsv1.Deployment) {
	deploymentToRemove := c.deploymentFromAPI(deployment)

	for _, id := range c.getIdentifiersFromAssocDeployment(deploymentToRemove) {
		if d, ok := c.Deployments[id]; ok && d.Name == deployment.Name {
			p, ok := c.GetDeployment(id)

			if ok && p.Name == deployment.Name {
				c.appendDeploymentDeleteQueue(id, deployment.Name)
			}
		}
	}
}

func (c *WatchDeploymentClient) appendDeploymentDeleteQueue(deploymentID DeploymentIdentifier, deploymentName string) {
	c.client.deleteMut.Lock()
	c.deleteQueue = append(c.deleteQueue, deploymentDeleteRequest{
		id:             deploymentID,
		deploymentName: deploymentName,
		ts:             time.Now(),
	})
	c.client.deleteMut.Unlock()
}

func (c *WatchDeploymentClient) shouldIgnoreDeployment(deployment *appsv1.Deployment) bool {
	// Check if user requested the deployment to be ignored through annotations
	if v, ok := deployment.Annotations[ignoreAnnotation]; ok {
		if strings.ToLower(strings.TrimSpace(v)) == "true" {
			return true
		}
	}

	// Check if user requested the deployment to be ignored through configuration
	for _, excludedDeployment := range c.Exclude.Deployments {
		if excludedDeployment.Name.MatchString(deployment.Name) {
			return true
		}
	}

	return false
}

func (c *WatchDeploymentClient) extractDeploymentLabelsAnnotations() bool {
	for _, r := range c.Rules.Labels {
		if r.From == MetadataFromDeployment {
			return true
		}
	}

	for _, r := range c.Rules.Annotations {
		if r.From == MetadataFromDeployment {
			return true
		}
	}

	return false
}
