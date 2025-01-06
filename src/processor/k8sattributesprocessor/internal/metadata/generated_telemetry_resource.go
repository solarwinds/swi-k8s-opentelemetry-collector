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

package metadata

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/otel/metric"
)

// addAdditionalMeters adds meters for the custom metrics that are not upstream, like `OtelsvcK8sJobAdded`
func addAdditionalMeters(builder *TelemetryBuilder, settings component.TelemetrySettings) error {
	var err, errs error

	// Deployment metrics
	builder.OtelsvcK8sDeploymentAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_deployment_added",
		metric.WithDescription("Number of deployment add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDeploymentUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_deployment_updated",
		metric.WithDescription("Number of deployment update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDeploymentDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_deployment_deleted",
		metric.WithDescription("Number of deployment delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDeploymentTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_deployment_table_size",
		metric.WithDescription("Size of table containing deployment info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// StatefulSet metrics
	builder.OtelsvcK8sStatefulSetAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_statefulset_added",
		metric.WithDescription("Number of statefulset add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sStatefulSetUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_statefulset_updated",
		metric.WithDescription("Number of statefulset update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sStatefulSetDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_statefulset_deleted",
		metric.WithDescription("Number of statefulset delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sStatefulSetTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_statefulset_table_size",
		metric.WithDescription("Size of table containing statefulset info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// ReplicaSet metrics
	builder.OtelsvcK8sReplicasetTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_replicaset_table_size",
		metric.WithDescription("Size of table containing replicaset info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// DaemonSet metrics
	builder.OtelsvcK8sDaemonSetAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_daemonset_added",
		metric.WithDescription("Number of daemonset add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDaemonSetUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_daemonset_updated",
		metric.WithDescription("Number of daemonset update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDaemonSetDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_daemonset_deleted",
		metric.WithDescription("Number of daemonset delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sDaemonSetTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_daemonset_table_size",
		metric.WithDescription("Size of table containing daemonset info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// Job metrics
	builder.OtelsvcK8sJobAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_job_added",
		metric.WithDescription("Number of job add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sJobUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_job_updated",
		metric.WithDescription("Number of job update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sJobDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_job_deleted",
		metric.WithDescription("Number of job delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sJobTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_job_table_size",
		metric.WithDescription("Size of table containing job info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// CronJob metrics
	builder.OtelsvcK8sCronJobAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_cronjob_added",
		metric.WithDescription("Number of cronjob add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sCronJobUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_cronjob_updated",
		metric.WithDescription("Number of cronjob update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sCronJobDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_cronjob_deleted",
		metric.WithDescription("Number of cronjob delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sCronJobTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_cronjob_table_size",
		metric.WithDescription("Size of table containing cronjob info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// Node metrics
	builder.OtelsvcK8sNodeTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_node_table_size",
		metric.WithDescription("Size of table containing node info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// Persistent volume metrics
	builder.OtelsvcK8sPersistentVolumeAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolume_added",
		metric.WithDescription("Number of persistentvolume add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolume_updated",
		metric.WithDescription("Number of persistentvolume update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolume_deleted",
		metric.WithDescription("Number of persistentvolume delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_persistentvolume_table_size",
		metric.WithDescription("Size of table containing persistentvolume info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// Persistent volume claim metrics
	builder.OtelsvcK8sPersistentVolumeClaimAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolumeclaim_added",
		metric.WithDescription("Number of persistentvolumeclaim add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeClaimUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolumeclaim_updated",
		metric.WithDescription("Number of persistentvolumeclaim update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeClaimDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_persistentvolumeclaim_deleted",
		metric.WithDescription("Number of persistentvolumeclaim delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sPersistentVolumeClaimTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_persistentvolumeclaim_table_size",
		metric.WithDescription("Size of table containing persistentvolumeclaim info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	// Service metrics
	builder.OtelsvcK8sServiceAdded, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_service_added",
		metric.WithDescription("Number of service add events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sServiceUpdated, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_service_updated",
		metric.WithDescription("Number of service update events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sServiceDeleted, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Counter(
		"otelcol_otelsvc_k8s_service_deleted",
		metric.WithDescription("Number of service delete events received"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)
	builder.OtelsvcK8sServiceTableSize, err = getLeveledMeter(builder.meter, configtelemetry.LevelBasic, settings.MetricsLevel).Int64Gauge(
		"otelcol_otelsvc_k8s_service_table_size",
		metric.WithDescription("Size of table containing service info"),
		metric.WithUnit("1"),
	)
	errs = errors.Join(errs, err)

	return errs
}
