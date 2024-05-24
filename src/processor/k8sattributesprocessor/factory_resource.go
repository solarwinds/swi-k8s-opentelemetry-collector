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

package swk8sattributesprocessor // import "github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor"

import "github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"

func createDeploymentProcessorOpts(deploymentConfig DeploymentConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromDeployment))
	opts = append(opts, withExtractMetadataDeployment(deploymentConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromDeployment, deploymentConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromDeployment, deploymentConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromDeployment, deploymentConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromDeployment, deploymentConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromDeployment, deploymentConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromDeployment, deploymentConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromDeployment, deploymentConfig.Exclude.Deployments))

	return opts
}

func createStatefulSetProcessorOpts(statefulSetConfig StatefulSetConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromStatefulSet))
	opts = append(opts, withExtractMetadataStatefulSet(statefulSetConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromStatefulSet, statefulSetConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromStatefulSet, statefulSetConfig.Exclude.StatefulSet))

	return opts
}

func createReplicaSetProcessorOpts(replicaSetConfig ReplicaSetConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromReplicaSet))
	opts = append(opts, withExtractMetadataReplicaSet(replicaSetConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromReplicaSet, replicaSetConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromReplicaSet, replicaSetConfig.Exclude.ReplicaSets))

	return opts
}

func createDaemonSetProcessorOpts(daemonSetConfig DaemonSetConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromDaemonSet))
	opts = append(opts, withExtractMetadataDaemonSet(daemonSetConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromDaemonSet, daemonSetConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromDaemonSet, daemonSetConfig.Exclude.DaemonSets))

	return opts
}

func createJobProcessorOpts(jobConfig JobConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromJob))
	opts = append(opts, withExtractMetadataJob(jobConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromJob, jobConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromJob, jobConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromJob, jobConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromJob, jobConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromJob, jobConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromJob, jobConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromJob, jobConfig.Exclude.Jobs))

	return opts
}

func createCronJobProcessorOpts(cronJobConfig CronJobConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromCronJob))
	opts = append(opts, withExtractMetadataCronJob(cronJobConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromCronJob, cronJobConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromCronJob, cronJobConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromCronJob, cronJobConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromCronJob, cronJobConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromCronJob, cronJobConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromCronJob, cronJobConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromCronJob, cronJobConfig.Exclude.CronJobs))

	return opts
}

func createNodeProcessorOpts(nodeConfig NodeConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromNode))
	opts = append(opts, withExtractMetadataNode(nodeConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromNode, nodeConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromNode, nodeConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromNode, nodeConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromNode, nodeConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromNode, nodeConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromNode, nodeConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromNode, nodeConfig.Exclude.Nodes))

	return opts
}

func createPersistentVolumeProcessorOpts(persistentVolumeConfig PersistentVolumeConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromPersistentVolume))
	opts = append(opts, withExtractMetadataPersistentVolumes(persistentVolumeConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromPersistentVolume, persistentVolumeConfig.Exclude.PVs))

	return opts
}

func createPersistentVolumeClaimProcessorOpts(persistentVolumeClaimConfig PersistentVolumeClaimConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromPersistentVolumeClaim))
	opts = append(opts, withExtractMetadataPersistentVolumeClaims(persistentVolumeClaimConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromPersistentVolumeClaim, persistentVolumeClaimConfig.Exclude.PVCs))

	return opts
}

func createServiceProcessorOpts(serviceConfig ServiceConfig) []option {
	var opts []option

	opts = append(opts, withResource(kube.MetadataFromService))
	opts = append(opts, withExtractMetadataService(serviceConfig.Extract.Metadata...))
	opts = append(opts, withExtractLabelsGeneric(kube.MetadataFromService, serviceConfig.Extract.Labels...))
	opts = append(opts, withExtractAnnotationsGeneric(kube.MetadataFromService, serviceConfig.Extract.Annotations...))
	opts = append(opts, withFilterNamespaceGeneric(kube.MetadataFromService, serviceConfig.Filter.Namespace))
	opts = append(opts, withFilterLabelsGeneric(kube.MetadataFromService, serviceConfig.Filter.Labels...))
	opts = append(opts, withFilterFieldsGeneric(kube.MetadataFromService, serviceConfig.Filter.Fields...))
	opts = append(opts, withExtractAssociationsGeneric(kube.MetadataFromService, serviceConfig.Association...))
	opts = append(opts, withExcludesResource(kube.MetadataFromService, serviceConfig.Exclude.Services))

	return opts
}
