# Changelog

All notable changes to Helm charts will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed

* Updated labels so that resources can be identified more easily
  * `app.kubernetes.io/name` changed to container application name (e.q. `swo-k8s-collector` for SWO k8s collector, `swo-agent` for SWO agent)
  * `app.kubernetes.io/part-of` always set to `swo-k8s-collector`

## [2.8.0-alpha.4] - 2023-10-09

### Fixed

- Detection of Node name for Fargate Nodes's metrics

### Added

- Scrape kube_service_* and kube_endpoint_* metrics

### Removed

- Removed attributes `net.host.name`, `net.host.port`, `http.scheme`, `prometheus`, `prometheus_replica` and `endpoint` from exported metrics

### Changed

- Updated docker image to 0.8.5 (see [Release notes](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.8.5))

## [2.8.0-alpha.3] - 2023-10-06

### Changed

- Changing log message attributes to respect OTEL log format

### Fixed

- Fixing nodeselector for kube-state-metrics so that it is deployed on linux nodes only

## [2.8.0-alpha.2] - 2023-10-04

### Added

- Add monitoring windows node logs

## [2.8.0-alpha.1] - 2023-09-11

### Added

- Added windows container for logs monitoring

## [2.7.0] - 2023-09-04

### Added

- Added new Helm settings `aws_fargate.enabled` and `aws_fargate.logs.*` that allow the k8s collector Helm chart to setup AWS EKS Fargate logging ConfigMap
  - Setting `prometheus.forceNamespace` can be used to force deployment of the bundled Prometheus to a specific non-Fargate namespace

### Changed

- Upgraded OTEL collector image to [0.8.2](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.8.2) which brings following changes
  - Bump library/golang from 1.20.7-bullseye to 1.21.0-bullseye and update some 3rd party dependencies
  - OTEL upgraded to v0.81.0
  - Updating `k8sattributes` to instrument attribute indicating that object exists
- Metrics no longer send `k8s.node.name` resource attribute if node does not exists in Kubernetes (for example in case of Fargate nodes)
- Adjusted Events collection to not produce resource attributes for entities that do not exists in Kubernetes
- DaemonSet for Log collection now restricts where it runs:
  - Fargate nodes are excluded
  - Only linux nodes with amd64 architecture are included

### Fixed
- Fixed Journal log collection on EKS (and other environment where journal logs are stored in `/var/log/journal`)

## [2.7.0-alpha.8] - 2023-08-31

### Changed
- Upgraded OTEL collector image to [0.8.2](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.8.2) which brings following changes
  - Bump library/golang from 1.20.7-bullseye to 1.21.0-bullseye and update some 3rd party dependencies

## [2.7.0-alpha.7] - 2023-08-30

### Fixed
- Usage metrics for nodes

## [2.7.0-alpha.6] - 2023-08-28

### Changed
- Metrics will no longer send `k8s.node.name` resource attribute if node does not exists in Kubernetes (for example in case of Fargate nodes)

## [2.7.0-alpha.5] - 2023-08-22

### Changed
- Adjusted bundled prometheus to not run on Fargate nodes by default
- Allowed use of `prometheus.forceNamespace` option of bundled prometheus, to force namespace where prometheus is deployed

## [2.7.0-alpha.4] - 2023-08-17

### Changed
- Adjusted Log group name used for Fargate logs
- Adjusted Events collection to not produce resource attributes for entities that do not exists in Kubernetes
- Upgraded OTEL collector image to [0.8.1](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.8.1) which brings following changes
  - Bump library/golang from 1.20.6-bullseye to 1.20.7-bullseye and update some 3rd party dependencies
  - Updating `k8sattributes` to instrument attribute indicating that object exists

## [2.7.0-alpha.3] - 2023-08-17

### Added
- There are new Helm settings `aws_fargate.enabled` and `aws_fargate.logs.enabled` that allow the k8s collector Helm chart to setup AWS EKS Fargate logging ConfigMap

### Changed
- Log collection DaemonSet now restrict where it runs:
  - Fargate nodes are excluded
  - Only linux nodes with amd64 architecture are included

### Fixed
- Fixed Journal log collection on EKS (and other environment where journal logs are stored in `/var/log/journal`)

## [2.7.0-alpha.2] - 2023-07-18

### Changed
- Upgraded OTEL collector image to [0.8.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.8.0) which brings following changes
  - OTEL upgraded to v0.81.0

## [2.6.0] - 2023-07-17

### Changed
- `k8s.pod.spec.cpu.limit` is calculated from kube-state-metrics (Kubernetes describe) and not from container runtime metrics. This should make the metric more reliable.
- We don't modify original prometheus metrics anymore. 

### Fixed
- Fixed filter on kube_namespace_status_phase, only values with 1 are sent

## [2.6.0-alpha.2] - 2023-07-11

### Fixed
- Fixed filter on kube_namespace_status_phase, only values with 1 are sent

## [2.6.0-alpha.1] - 2023-06-08

### Changed
- We don't modify original prometheus metrics anymore. 

## [2.5.0] - 2023-06-12

### Added
- Added PV and PVC metrics
  - Updating docker image to `0.7.0` (which is capable of instrumenting PV and PVC with labels/annotations)

### Changed
- Updating docker images `solarwinds/swo-agent`, `busybox`, `fullstorydev/grpcurl` and `alpine/k8s` to latest available versions

### Fixed
- Fixed CPU/Memory usage Pod metrics on latest cAdvisor/containerd (not relying on Pod level datapoints, but doing SUM of container datapoints)
- Fixed node level network metrics for environments where pod level network metrics are not available (for example Docker runtime)

## [2.5.0-alpha.6] - 2023-06-08

### Changed
- Updating docker images `solarwinds/swo-agent`, `busybox`, `fullstorydev/grpcurl` and `alpine/k8s` to latest available versions

## [2.5.0-alpha.5] - 2023-06-07
### Fixed
- Fixed node level network metrics for environments where pod level network metrics are not available (for example Docker runtime)

## [2.5.0-alpha.4] - 2023-06-05
### Fixed
- Fixed CPU/Memory usage Pod metrics on latest cAdvisor/containerd (not relying on Pod level datapoints, but doing SUM of container datapoints)

## [2.5.0-alpha.3] - 2023-06-02
### Added
- Added `k8s.kube_pod_spec_volumes_persistentvolumeclaims_info` metric to connect Pod and PVC

## [2.5.0-alpha.2] - 2023-05-31

### Changed
- `access_mode` is now published as resource attribute
- `kubelet_*` metrics are published to SWO with prefix `k8s` (to be consistent with other kubernetes related metrics)

## [2.5.0-alpha.1] - 2023-05-25
### Added
- Added PV and PVC metrics

### Changed
- Updating docker image to `0.7.0` (which is capable of instrumenting PV and PVC with labels/annotations)

## [2.4.1] - 2023-05-25

### Changed
- Updating docker image to `0.6.0` (which includes some security fixes and add forwardconnector OTEL component)

### Fixed
- Fixed filter on kube-state-metrics so that only specific metrics are sent 

## [2.4.0] - 2023-05-16

### Added
- Added new container metrics `k8s.container.fs.iops`, `k8s.container.fs.throughput`, `k8s.container.network.bytes_received` and `k8s.container.network.bytes_transmitted`
- Added scraping of `kube_pod_init_container_*` metrics
- Merics `k8s.container.spec.cpu.limit`, `k8s.container.spec.cpu.requests`, `k8s.container.spec.memory.requests`, `k8s.container.spec.memory.limit` and `k8s.container.status` now include datapoints for both init and non-init containers
- `kube-state-metrics` is now bundled with the Helm chart so that its metrics are predictable
- FIPS compliance

### Changed
- Upgraded OTEL collector image to [0.5.2](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.5.2) which brings following changes
  - FIPS support
  - Updated build dependencies (security fixes)

### Removed
- Removed metrics `k8s.cluster.memory.utilization` and `k8s.cluster.cpu.utilization` - they are replaced by composite metrics calculated by the SWO platform

### Fixed
- Fixed Autoupdate
  - Adjusted permissions to be able to update ClusterRoles for future increments
  - The update is now atomic, so in case it fails, it will rollback (it will not leave Helm release in Failed state)
- Metric `k8s.kube_pod_status_phase` should not send values with 0 anymore

## [2.4.0-alpha.6] - 2023-05-04

### Fixed
- `k8s.kube_pod_status_phase` should not send values with 0 anymore

## [2.4.0-alpha.5] - 2023-05-02

### Added
- FIPS compliance

### Fixed
- Updated docker image to [0.5.1](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.5.1) which contains security fixes
- Fixed Autoupdate
  - Adjusted permissions to be able to update clusterroles for future increments
  - The update is now atomic, so in case it fails it will rollback (it will not leave Helm release in Failed state).

## [2.4.0-alpha.4] - 2023-04-25

### Changed
- Updated metrics `k8s.container.fs.iops`, `k8s.container.fs.throughput` to be correctly bind by swo processing pipeline


## [2.4.0-alpha.3] - 2023-04-25

### Changed
- Updated metrics `k8s.container.fs.iops`, `k8s.container.fs.throughput` to be correctly bind by swo processing pipeline


## [2.4.0-alpha.2] - 2023-04-24

### Added
- kube-state-metrics is now bundled with the Helm chart so that its metrics are predictable

### Changed
- `k8s.cluster.memory.utilization` and `k8s.cluster.cpu.utilization` are no longer calculated. They are replaced by composite metric calculated by the platform
- `k8s.container.spec.cpu.limit`, `k8s.container.spec.cpu.requests`, `k8s.container.spec.memory.requests`, `k8s.container.spec.memory.limit` and `k8s.container.status` now includes datapoints for both init and non-init containers

## [2.4.0-alpha.1] - 2023-04-19

### Added
- Added new container metrics `k8s.container.fs.iops`, `k8s.container.fs.throughput`, `k8s.container.network.bytes_received`, `k8s.container.network.bytes_transmitted`
- Added scraping of `kube_pod_init_container_*` metrics

## [2.3.0] - 2023-04-13

### Added
- Added automatic extraction of Kubernetes labels and annotations from resources (Pods, Namespaces, Deployment, StatefulSet, ReplicaSet, DaemonSet, Job, CronJob, Node) and sent using resource attributes with metrics end events.
- A new option to deploy `prometheus` as part of the k8s collector chart installation, controlled by setting `prometheus.enabled: true` in `values.yaml`.
- New StatefulSet with light weight SWO Agent optionally deployed by default
- Added syslog attributes for log entry so that logs are properly displayed by LogViewer (`syslog.facility`, `syslog.version`, `syslog.procid`, `syslog.msgid`)
  - Added resource level attributes: `host.hostname` contains name of the pod (represented as System in LogViewer), `service.name` contains name of the container (represented as Program in LogViewer).
- New metrics are scraped from Prometheus: `k8s.kube_replicaset_spec_replicas`, `k8s.kube_replicaset_status_ready_replicas`, `k8s.kube_replicaset_status_replicas`
- Added metrics `k8s.cluster.version` which extract version from `kubernetes_build_info`. Metric `kubernetes_build_info` is no longer published

### Fixed
- Enabled `honor_labels` option to keep scraped labels unchanged
- Fixed `k8s.job.condition` resource attribute to handle Failed state
- Fixed calculation of `k8s.pod.spec.memory.limit` on newer container runtime (no longer use `container_spec_memory_limit_bytes`, but `kube_pod_container_resource_limits`)
- Fix grouping conditions for `container_network_*` and `container_fs_*` metrics to not rely on container attribute

## [2.3.0-alpha.7] - 2023-04-12
- Added automatic extraction of Kubernetes labels and annotations from events.

## [2.3.0-alpha.6] - 2023-04-06

### Added
- Added automatic extraction of Kubernetes labels and annotations from additional resources (Deployment, StatefulSet, ReplicaSet, DaemonSet, Job, CronJob, Node) and sent using resource attributes with metric

## [2.3.0-alpha.5] - 2023-04-06

### Changed
- Enabled honor_labels option to keep scraped data over server-side labels

### Fixed
- Fixed calculation of `k8s.pod.spec.memory.limit` on newer container runtime (no longer use `container_spec_memory_limit_bytes`, but `kube_pod_container_resource_limits`)

## [2.3.0-alpha.4] - 2023-03-29

### Added
- New StatefulSet with light weight SWO Agent optionaly deployed by default
- Added syslog attributes for log entry: `syslog.facility`, `syslog.version`, `syslog.procid`, `syslog.msgid`.
- Added resource level attributes: `host.hostname` contains name of the pod, `service.name` contains name of the container.

## [2.3.0-alpha.3] - 2023-03-24

### Changed

- Fixed k8s.job.condition resource attribute to handle Failed state

### Added

- New replicaset metrics `k8s.kube_replicaset_spec_replicas`, `k8s.kube_replicaset_status_ready_replicas`, `k8s.kube_replicaset_status_replicas`

## [2.3.0-alpha.2] - 2023-03-22

### Added

- A new option to deploy `prometheus` as part of the k8s collector chart installation, controlled by setting `prometheus.enabled: true` in `values.yaml`.
- Added automatic extraction of Kubernetes labels and annotations from resources (Pods, Namespaces) and sent using resource attributes with metric

## [2.3.0-alpha.1] - 2023-03-21

### Changed

- Fix grouping conditions for `container_network_*` and `container_fs_*` metrics to not rely on container attribute
- Added metrics k8s.cluster.version which extract version from kubernetes_build_info. Metric kubernetes_build_info is not published

## [2.2.0] - 2023-03-23

### Added

- Collect metrics about kube jobs.
- Attribute k8s.pod.name to events
- Added internal ip attribute for node [#168](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/168).
- Added metric k8s.kubernetes_build_info for calculation of sw.k8s.cluster.version
- InitContainers where Prometheus and OTEL endpoints are checked
- k8s.namespace.name moved from attributes to resource.attributes
- Filtering out [internal metrics generated by Prometheus scraper](https://prometheus.io/docs/concepts/jobs_instances/#automatically-generated-labels-and-time-series) (`scrape_duration_seconds`, `scrape_samples_post_metric_relabeling`, `scrape_samples_scraped`, `scrape_series_added`, `up`).
- introduced k8s.job.condition resource attribute for job which can be Pending, Complete or Failed
- introduced k8s.container.spec.cpu.limit metric for CPU quota
- Added possibility to deploy `PodMonitor` resources so that OTEL collector telemetry is scraped by Prometheus Operator (see [Prometheus Operator design](https://prometheus-operator.dev/docs/operator/design/))
- Added k8s.container.cpu.usage.seconds.rate metric
- Adding `container.id` and `container.runtime` attributes to `k8s.kube_pod_container_info` metric for unique container identification [#182](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/182).
- Added optional autoupdate support (set by `autoupdate.enabled` in `values.yaml`) [#196](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/196).

### Changed

- Fix grouping conditions for container*network*_ and container*fs*_ metrics to not relly on container attribute
- Added metrics k8s.cluster.version which extract version from kubernetes_build_info. Metric kubernetes_build_info is not published
- Filtering out datapoints for internal k8s containers (with name "POD", usually using image "pause")
- Upgraded OTEL collector image to [0.4.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.4.0) which brings following changes
  - OTEL upgraded to 0.73.0
  - Updated build dependencies

## [2.2.0-beta.1] - 2023-03-16

### Added

- Adding `container.id` and `container.runtime` attributes to `k8s.kube_pod_container_info` metric for unique container identification [#182](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/182).
- Added optional autoupdate support (set by `autoupdate.enabled` in `values.yaml`) [#196](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/196).

### Changed

- Upgraded OTEL collector image to [0.4.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.4.0) which brings following changes
  - OTEL upgraded to 0.73.0
  - Updated build dependencies

## [2.2.0-alpha.4] - 2023-03-14

### Added

- Added possibility to deploy `PodMonitor` resources so that OTEL collector telemetry is scraped by Prometheus Operator (see [Prometheus Operator design](https://prometheus-operator.dev/docs/operator/design/))
- Added k8s.container.cpu.usage.seconds.rate metric

### Changed

- k8s.job.condition, state name Pending changed to Active
- Filtering out datapoints for internal k8s containers (with name "POD", usually using image "pause")

## [2.2.0-alpha.3] - 2023-03-09

### Added

- k8s.namespace.name moved from attributes to resource.attributes
- Filtering out [internal metrics generated by Prometheus scraper](https://prometheus.io/docs/concepts/jobs_instances/#automatically-generated-labels-and-time-series) (`scrape_duration_seconds`, `scrape_samples_post_metric_relabeling`, `scrape_samples_scraped`, `scrape_series_added`, `up`).
- introduced k8s.job.condition resource attribute for job which can be Pending, Complete or Failed
- introduced k8s.container.spec.cpu.limit metric for CPU quota

## [2.2.0-alpha.2] - 2023-03-02

### Added

- InitContainers where Prometheus and OTEL endpoints are checked
- Added metric k8s.kubernetes_build_info for calculation of sw.k8s.cluster.version

## [2.2.0-alpha.1] - 2023-02-21

### Added

- Collect metrics about kube jobs.
- Attribute k8s.pod.name to events
- Added internal ip attribute for node [#168](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/168).

## [2.1.0] - 2023-02-16

### Added

- Added telemetry port to kubernetes "ports" [#129](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/129)
- Added `tolerations` config of Logs collection DaemonSet (with default to run on tainted nodes) [#141](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/141)
- if telemetry is enabled (true by default) OTEL collectors will contain prometheus annotations so that telemetry is discovered by Prometheus [#152](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/152).
- configuration of `file_storage` extension is now available in `values.yaml`. [#157](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/157)
  - default `timeout` is now set to `5s`
  - Log collector suffix clusterId into folder which it mounts for checkpoints (e.q.`/var/lib/swo/checkpoints/<clusterId>`). This avoid unpredictable errors in scenario when previous monitoring was not deleted. [#161](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/161)
- new events transform pipeline which sets **sw.namespace** attribute to **sw.events.inframon.k8s** [#155](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/155).
  [#145](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/145).
- Exposed some configuration of [filelogreciever](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) in `values.yaml` [#146](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/146).

### Changed

- Upgraded OTEL collector image to [0.3.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.3.0) which brings following changes
  - OTEL upgraded to 0.69.0
  - Added `filestorage` so it can be in processors

### Fixed

- Properly annotate configmap checksums [#151](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)
- Template now use `https_proxy_url` from the right place [#151(https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)
- Added optimizations to Log collector preventing Out of Memory situations [#137](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/137)
  - Added `filelog.storage` to persist checkpoints in persistent storage, not in memory.
  - Setting `max_concurrent_files` to 10 (from default 1024), this is main memory optimization, reducing amount of concurrent log scans
  - Increased default memory limit to `700Mi` (which should be enough for large logs)
  - Having by default `150Mi` difference between OTEL memory limit and Kubernetes memory limit, so that OTEL has enough buffer (this prevents OOMing)

## [2.1.0-beta.2] - 2023-02-15

### Changed

- Remove generated attribute k8s.deployment.name
- Log collector suffix clusterId into folder which it mounts for checkpoints (e.q.`/var/lib/swo/checkpoints/<clusterId>`). This avoid unpredictable errors in scenario when previous monitoring was not deleted. [#161](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/161)

## [2.1.0-beta.1] - 2023-02-09

### Added

- if telemetry is enabled (true by default) OTEL collectors will contain prometheus annotations so that telemetry is discovered by Prometheus [#152](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/152).
- configuration of `file_storage` extension is now available in `values.yaml`. [#157](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/157)
  - default `timeout` is now set to `5s`
- new events transform pipeline which sets **sw.namespace** attribute to **sw.events.inframon.k8s** [#155](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/155).

### Fixed

- Properly annotate configmap checksums [#151](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)
- Template now use `https_proxy_url` from the right place [#151(https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)

## [2.1.0-alpha.3] - 2023-01-31

### Added

- Add attribute k8s.deployment.name to exported logs
  [#145](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/145).

## [2.1.0-alpha.2] - 2023-01-30

### Added

- Exposed some configuration of [filelogreciever](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) in `values.yaml` [#146](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/146).

### Changed

- Changed default value `start_at` property of [filelogreciever](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) to `end` [#146](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/146).

## [2.1.0-alpha.1] - 2023-01-25

### Added

- Added telemetry port to kubernetes "ports" [#129](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/129)
- Added `tolerations` config of Logs collection DaemonSet (with default to run on tainted nodes) [#141](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/141)

### Changed

- Upgraded OTEL collector image to [0.3.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.3.0) which brings following changes
  - OTEL upgraded to 0.69.0
  - Added `filestorage` so it can be in processors

### Fixed

- Added optimizations to Log collector preventing Out of Memory situations [#137](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/137)
  - Added `filelog.storage` to persist checkpoints in persistent storage, not in memory.
  - Setting `max_concurrent_files` to 10 (from default 1024), this is main memory optimization, reducing amount of concurrent log scans
  - Increased default memory limit to `700Mi` (which should be enough for large logs)
  - Having by default `150Mi` difference between OTEL memory limit and Kubernetes memory limit, so that OTEL has enough buffer (this prevents OOMing)

## [2.0.2] - 2023-01-18

### Added

- Initial Helm release.
- Create Error reason as mapping from combination of Event Reason and Type fields [#115](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/115)
- Add support for HTTPS proxies [#117](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/117)
