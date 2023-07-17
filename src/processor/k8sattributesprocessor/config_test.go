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

package k8sattributesprocessor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id: component.NewID(typeStr),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
				Exclude:   ExcludeConfig{Pods: []ExcludePodConfig{{Name: "jaeger-agent"}, {Name: "jaeger-collector"}}},
			},
		},
		{
			id: component.NewIDWithName(typeStr, "2"),
			expected: &Config{
				APIConfig:   k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeKubeConfig},
				Passthrough: false,
				Extract: ExtractConfig{
					Metadata: []string{"k8s.pod.name", "k8s.pod.uid", "k8s.deployment.name", "k8s.namespace.name", "k8s.node.name", "k8s.pod.start_time"},
					Annotations: []FieldExtractConfig{
						{TagName: "a1", Key: "annotation-one", From: "pod"},
						{TagName: "a2", Key: "annotation-two", Regex: "field=(?P<value>.+)", From: kube.MetadataFromPod},
					},
					Labels: []FieldExtractConfig{
						{TagName: "l1", Key: "label1", From: "pod"},
						{TagName: "l2", Key: "label2", Regex: "field=(?P<value>.+)", From: kube.MetadataFromPod},
					},
				},
				Filter: FilterConfig{
					Namespace:      "ns2",
					Node:           "ip-111.us-west-2.compute.internal",
					NodeFromEnvVar: "K8S_NODE",
					Labels: []FieldFilterConfig{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2", Op: "not-equals"},
					},
					Fields: []FieldFilterConfig{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2", Op: "not-equals"},
					},
				},
				Association: []AssociationConfig{
					{
						Sources: []AssociationSourceConfig{
							{
								From: "resource_attribute",
								Name: "ip",
							},
						},
					},
					{
						Sources: []AssociationSourceConfig{
							{
								From: "resource_attribute",
								Name: "k8s.pod.ip",
							},
						},
					},
					{
						Sources: []AssociationSourceConfig{
							{
								From: "resource_attribute",
								Name: "host.name",
							},
						},
					},
					{
						Sources: []AssociationSourceConfig{
							{
								From: "connection",
								Name: "ip",
							},
						},
					},
				},
				Deployment: DeploymentConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.deployment.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.deployment.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromDeployment},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.deployment.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromDeployment},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.deployment.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				StatefulSet: StatefulSetConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.statefulset.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.statefulset.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromStatefulSet},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.statefulset.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromStatefulSet},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.statefulset.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				ReplicaSet: ReplicaSetConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.replicaset.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.replicaset.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromReplicaSet},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.replicaset.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromReplicaSet},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.replicaset.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				DaemonSet: DaemonSetConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.daemonset.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.daemonset.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromDaemonSet},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.daemonset.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromDaemonSet},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.daemonset.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},

				Job: JobConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.job.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.job.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromJob},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.job.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromJob},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.job.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				CronJob: CronJobConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.cronjob.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.cronjob.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromCronJob},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.cronjob.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromCronJob},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.cronjob.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				Node: NodeConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.node.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.node.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromNode},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.node.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromNode},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.node.name",
								},
							},
						},
					},
				},
				PersistentVolume: PersistentVolumeConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.persistentvolume.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.persistentvolume.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromPersistentVolume},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.persistentvolume.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromPersistentVolume},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.persistentvolume.name",
								},
							},
						},
					},
				},
				PersistentVolumeClaim: PersistentVolumeClaimConfig{
					Extract: ExtractConfig{
						Metadata: []string{"k8s.persistentvolumeclaim.uid"},
						Annotations: []FieldExtractConfig{
							{TagName: "k8s.persistentvolumeclaim.annotations.$$1", KeyRegex: "(.*)", From: kube.MetadataFromPersistentVolumeClaim},
						},
						Labels: []FieldExtractConfig{
							{TagName: "k8s.persistentvolumeclaim.labels.$$1", KeyRegex: "(.*)", From: kube.MetadataFromPersistentVolumeClaim},
						},
					},
					Association: []AssociationConfig{
						{
							Sources: []AssociationSourceConfig{
								{
									From: "resource_attribute",
									Name: "k8s.persistentvolumeclaim.name",
								},
								{
									From: "resource_attribute",
									Name: "k8s.namespace.name",
								},
							},
						},
					},
				},
				Exclude: ExcludeConfig{
					Pods: []ExcludePodConfig{
						{Name: "jaeger-agent"},
						{Name: "jaeger-collector"},
					},
				},
			},
		},
		{
			id: component.NewIDWithName(typeStr, "3"),
			expected: &Config{
				APIConfig:   k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeKubeConfig},
				Passthrough: false,
				Extract: ExtractConfig{
					Annotations: []FieldExtractConfig{
						{KeyRegex: "opentel.*", From: kube.MetadataFromPod},
					},
					Labels: []FieldExtractConfig{
						{KeyRegex: "opentel.*", From: kube.MetadataFromPod},
					},
				},
				Exclude: ExcludeConfig{
					Pods: []ExcludePodConfig{
						{Name: "jaeger-agent"},
						{Name: "jaeger-collector"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
