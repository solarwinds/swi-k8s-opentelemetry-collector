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
	"fmt"
	"os"
	"regexp"
	"time"

	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/internal/k8sconfig"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/metadata"
)

const (
	filterOPEquals       = "equals"
	filterOPNotEquals    = "not-equals"
	filterOPExists       = "exists"
	filterOPDoesNotExist = "does-not-exist"
	metadataPodIP        = "k8s.pod.ip"
	metadataPodStartTime = "k8s.pod.start_time"
	specPodHostName      = "k8s.pod.hostname"
	// TODO: use k8s.cluster.uid, container.image.repo_digests
	// from semconv when available,
	//   replace clusterUID with conventions.AttributeK8SClusterUid
	//   replace containerRepoDigests with conventions.AttributeContainerImageRepoDigests
	clusterUID                = "k8s.cluster.uid"
	containerImageRepoDigests = "container.image.repo_digests"
)

// option represents a configuration option that can be passes.
// to the k8s-tagger
type option func(*kubernetesprocessor) error

// withAPIConfig provides k8s API related configuration to the processor.
// It defaults the authentication method to in-cluster auth using service accounts.
func withAPIConfig(cfg k8sconfig.APIConfig) option {
	return func(p *kubernetesprocessor) error {
		p.apiConfig = cfg
		return p.apiConfig.Validate()
	}
}

// withPassthrough enables passthrough mode. In passthrough mode, the processor
// only detects and tags the pod IP and does not invoke any k8s APIs.
func withPassthrough() option {
	return func(p *kubernetesprocessor) error {
		p.passthroughMode = true
		return nil
	}
}

// withSetObjectExistence enables mode where `sw.k8s.<object type>.found` attributes will be instrumented
func withSetObjectExistence() option {
	return func(p *kubernetesprocessor) error {
		p.setObjectExistence = true
		return nil
	}
}

// enabledAttributes returns the list of resource attributes enabled by default.
func enabledAttributes() (attributes []string) {
	defaultConfig := metadata.DefaultResourceAttributesConfig()
	if defaultConfig.K8sClusterUID.Enabled {
		attributes = append(attributes, clusterUID)
	}
	if defaultConfig.ContainerID.Enabled {
		attributes = append(attributes, conventions.AttributeContainerID)
	}
	if defaultConfig.ContainerImageName.Enabled {
		attributes = append(attributes, conventions.AttributeContainerImageName)
	}
	if defaultConfig.ContainerImageRepoDigests.Enabled {
		attributes = append(attributes, containerImageRepoDigests)
	}
	if defaultConfig.ContainerImageTag.Enabled {
		attributes = append(attributes, conventions.AttributeContainerImageTag)
	}
	if defaultConfig.K8sContainerName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SContainerName)
	}
	if defaultConfig.K8sCronjobName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SCronJobName)
	}
	if defaultConfig.K8sDaemonsetName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SDaemonSetName)
	}
	if defaultConfig.K8sDaemonsetUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SDaemonSetUID)
	}
	if defaultConfig.K8sDeploymentName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SDeploymentName)
	}
	if defaultConfig.K8sDeploymentUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SDeploymentUID)
	}
	if defaultConfig.K8sJobName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SJobName)
	}
	if defaultConfig.K8sJobUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SJobUID)
	}
	if defaultConfig.K8sNamespaceName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SNamespaceName)
	}
	if defaultConfig.K8sNodeName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SNodeName)
	}
	if defaultConfig.K8sPodHostname.Enabled {
		attributes = append(attributes, specPodHostName)
	}
	if defaultConfig.K8sPodName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SPodName)
	}
	if defaultConfig.K8sPodStartTime.Enabled {
		attributes = append(attributes, metadataPodStartTime)
	}
	if defaultConfig.K8sPodUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SPodUID)
	}
	if defaultConfig.K8sPodIP.Enabled {
		attributes = append(attributes, metadataPodIP)
	}
	if defaultConfig.K8sReplicasetName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SReplicaSetName)
	}
	if defaultConfig.K8sReplicasetUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SReplicaSetUID)
	}
	if defaultConfig.K8sStatefulsetName.Enabled {
		attributes = append(attributes, conventions.AttributeK8SStatefulSetName)
	}
	if defaultConfig.K8sStatefulsetUID.Enabled {
		attributes = append(attributes, conventions.AttributeK8SStatefulSetUID)
	}
	return
}

// withExtractMetadata allows specifying options to control extraction of pod metadata.
// If no fields explicitly provided, the defaults are pulled from metadata.yaml.
func withExtractMetadata(fields ...string) option {
	return func(p *kubernetesprocessor) error {
		for _, field := range fields {
			switch field {
			case conventions.AttributeK8SNamespaceName:
				p.rules.Namespace = true
			case conventions.AttributeK8SPodName:
				p.rules.PodName = true
			case conventions.AttributeK8SPodUID:
				p.rules.PodUID = true
			case specPodHostName:
				p.rules.PodHostName = true
			case metadataPodStartTime:
				p.rules.StartTime = true
			case metadataPodIP:
				p.rules.PodIP = true
			case conventions.AttributeK8SDeploymentName:
				p.rules.Deployment = true
			case conventions.AttributeK8SReplicaSetName:
				p.rules.ReplicaSetName = true
			case conventions.AttributeK8SReplicaSetUID:
				p.rules.ReplicaSetID = true
			case conventions.AttributeK8SDaemonSetName:
				p.rules.DaemonSetName = true
			case conventions.AttributeK8SDaemonSetUID:
				p.rules.DaemonSetUID = true
			case conventions.AttributeK8SStatefulSetName:
				p.rules.StatefulSetName = true
			case conventions.AttributeK8SStatefulSetUID:
				p.rules.StatefulSetUID = true
			case conventions.AttributeK8SContainerName:
				p.rules.ContainerName = true
			case conventions.AttributeK8SJobName:
				p.rules.JobName = true
			case conventions.AttributeK8SJobUID:
				p.rules.JobUID = true
			case conventions.AttributeK8SCronJobName:
				p.rules.CronJobName = true
			case conventions.AttributeK8SNodeName:
				p.rules.Node = true
			case conventions.AttributeContainerID:
				p.rules.ContainerID = true
			case conventions.AttributeContainerImageName:
				p.rules.ContainerImageName = true
			case containerImageRepoDigests:
				p.rules.ContainerImageRepoDigests = true
			case conventions.AttributeContainerImageTag:
				p.rules.ContainerImageTag = true
			case clusterUID:
				p.rules.ClusterUID = true
			}
		}
		return nil
	}
}

// withExtractLabels allows specifying options to control extraction of pod labels.
func withExtractLabels(labels ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		labels, err := extractFieldRules("labels", labels...)
		if err != nil {
			return err
		}
		p.rules.Labels = labels
		return nil
	}
}

// withExtractAnnotations allows specifying options to control extraction of pod annotations tags.
func withExtractAnnotations(annotations ...FieldExtractConfig) option {
	return func(p *kubernetesprocessor) error {
		annotations, err := extractFieldRules("annotations", annotations...)
		if err != nil {
			return err
		}
		p.rules.Annotations = annotations
		return nil
	}
}

func extractFieldRules(fieldType string, fields ...FieldExtractConfig) ([]kube.FieldExtractionRule, error) {
	var rules []kube.FieldExtractionRule
	for _, a := range fields {
		name := a.TagName

		switch a.From {
		// By default if the From field is not set for labels and annotations we want to extract them from pod
		case "", kube.MetadataFromPod:
			a.From = kube.MetadataFromPod
		case kube.MetadataFromNamespace:
			a.From = kube.MetadataFromNamespace
		case kube.MetadataFromDeployment:
			a.From = kube.MetadataFromDeployment
		case kube.MetadataFromStatefulSet:
			a.From = kube.MetadataFromStatefulSet
		case kube.MetadataFromReplicaSet:
			a.From = kube.MetadataFromReplicaSet
		case kube.MetadataFromDaemonSet:
			a.From = kube.MetadataFromDaemonSet
		case kube.MetadataFromJob:
			a.From = kube.MetadataFromJob
		case kube.MetadataFromCronJob:
			a.From = kube.MetadataFromCronJob
		case kube.MetadataFromNode:
			a.From = kube.MetadataFromNode
		case kube.MetadataFromPersistentVolume:
			a.From = kube.MetadataFromPersistentVolume
		case kube.MetadataFromPersistentVolumeClaim:
			a.From = kube.MetadataFromPersistentVolumeClaim
		case kube.MetadataFromService:
			a.From = kube.MetadataFromService
		default:
			return rules, fmt.Errorf("%s is not a valid choice for From. Must be one of: pod, deployment, statefulset, replicaset, daemonset, job, cronjob, node, namespace, persistentvolume, persistentvolumeclaim, service", a.From)
		}

		if name == "" && a.Key != "" {
			// name for KeyRegex case is set at extraction time/runtime, skipped here
			if a.From == kube.MetadataFromPod {
				name = fmt.Sprintf("k8s.pod.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromNamespace {
				name = fmt.Sprintf("k8s.namespace.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromDeployment {
				name = fmt.Sprintf("k8s.deployment.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromStatefulSet {
				name = fmt.Sprintf("k8s.statefulset.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromReplicaSet {
				name = fmt.Sprintf("k8s.replicaset.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromDaemonSet {
				name = fmt.Sprintf("k8s.daemonset.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromJob {
				name = fmt.Sprintf("k8s.job.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromCronJob {
				name = fmt.Sprintf("k8s.cronjob.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromNode {
				name = fmt.Sprintf("k8s.node.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromPersistentVolume {
				name = fmt.Sprintf("k8s.persistentvolume.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromPersistentVolumeClaim {
				name = fmt.Sprintf("k8s.persistentvolumeclaim.%s.%s", fieldType, a.Key)
			} else if a.From == kube.MetadataFromService {
				name = fmt.Sprintf("k8s.service.%s.%s", fieldType, a.Key)
			}
		}

		var r *regexp.Regexp
		if a.Regex != "" {
			var err error
			r, err = regexp.Compile(a.Regex)
			if err != nil {
				return rules, err
			}
			names := r.SubexpNames()
			if len(names) != 2 || names[1] != "value" {
				return rules, fmt.Errorf("regex must contain exactly one named submatch (value)")
			}
		}

		var keyRegex *regexp.Regexp
		var hasKeyRegexReference bool
		if a.KeyRegex != "" {
			var err error
			keyRegex, err = regexp.Compile("^(?:" + a.KeyRegex + ")$")
			if err != nil {
				return rules, err
			}

			if keyRegex.NumSubexp() > 0 {
				hasKeyRegexReference = true
			}
		}

		rules = append(rules, kube.FieldExtractionRule{
			Name: name, Key: a.Key, KeyRegex: keyRegex, HasKeyRegexReference: hasKeyRegexReference, Regex: r, From: a.From,
		})
	}
	return rules, nil
}

// withFilterNode allows specifying options to control filtering pods by a node/host.
func withFilterNode(node, nodeFromEnvVar string) option {
	return func(p *kubernetesprocessor) error {
		if nodeFromEnvVar != "" {
			p.filters.Node = os.Getenv(nodeFromEnvVar)
			return nil
		}
		p.filters.Node = node
		return nil
	}
}

// withFilterNamespace allows specifying options to control filtering pods by a namespace.
func withFilterNamespace(ns string) option {
	return func(p *kubernetesprocessor) error {
		p.filters.Namespace = ns
		return nil
	}
}

// withFilterLabels allows specifying options to control filtering pods by pod labels.
func withFilterLabels(filters ...FieldFilterConfig) option {
	return func(p *kubernetesprocessor) error {
		var labels []kube.FieldFilter
		for _, f := range filters {
			var op selection.Operator
			switch f.Op {
			case filterOPNotEquals:
				op = selection.NotEquals
			case filterOPExists:
				op = selection.Exists
			case filterOPDoesNotExist:
				op = selection.DoesNotExist
			default:
				op = selection.Equals
			}
			labels = append(labels, kube.FieldFilter{
				Key:   f.Key,
				Value: f.Value,
				Op:    op,
			})
		}
		p.filters.Labels = labels
		return nil
	}
}

// withFilterFields allows specifying options to control filtering pods by pod fields.
func withFilterFields(filters ...FieldFilterConfig) option {
	return func(p *kubernetesprocessor) error {
		var fields []kube.FieldFilter
		for _, f := range filters {
			var op selection.Operator
			switch f.Op {
			case filterOPNotEquals:
				op = selection.NotEquals
			default:
				op = selection.Equals
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

// withExtractPodAssociations allows specifying options to associate pod metadata with incoming resource
func withExtractPodAssociations(podAssociations ...AssociationConfig) option {
	return func(p *kubernetesprocessor) error {
		associations := make([]kube.Association, 0, len(podAssociations))
		var assoc kube.Association
		for _, association := range podAssociations {
			assoc = kube.Association{
				Sources: []kube.AssociationSource{},
			}

			var name string

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
			associations = append(associations, assoc)
		}
		p.podAssociations = associations
		return nil
	}
}

// withExcludes allows specifying pods to exclude
func withExcludes(podExclude ExcludeConfig) option {
	return func(p *kubernetesprocessor) error {
		ignoredNames := kube.Excludes{}
		names := podExclude.Pods

		if len(names) == 0 {
			names = []ExcludePodConfig{{Name: "jaeger-agent"}, {Name: "jaeger-collector"}}
		}
		for _, name := range names {
			ignoredNames.Pods = append(ignoredNames.Pods, kube.ExcludePods{Name: regexp.MustCompile(name.Name)})
		}
		p.podIgnore = ignoredNames
		return nil
	}
}

// withWaitForMetadata allows specifying whether to wait for pod metadata to be synced.
func withWaitForMetadata(wait bool) option {
	return func(p *kubernetesprocessor) error {
		p.waitForMetadata = wait
		return nil
	}
}

// withWaitForMetadataTimeout allows specifying the timeout for waiting for pod metadata to be synced.
func withWaitForMetadataTimeout(timeout time.Duration) option {
	return func(p *kubernetesprocessor) error {
		p.waitForMetadataTimeout = timeout
		return nil
	}
}
