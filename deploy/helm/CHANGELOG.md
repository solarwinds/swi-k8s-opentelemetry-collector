# Changelog

All notable changes to Helm charts will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Changed
* Remove generated attribute k8s.deployment.name

## [2.1.0-beta.1] - 2023-02-09

### Added
* if telemetry is enabled (true by default) OTEL collectors will contain prometheus annotations so that telemetry is discovered by Prometheus [#152](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/152).
* configuration of `file_storage` extension is now available in `values.yaml`. [#157](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/157)
    * default `timeout` is now set to `5s`
* new events transform pipeline which sets **sw.namespace** attribute to **sw.events.inframon.k8s** [#155](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/155).

### Fixed
* Properly annotate configmap checksums [#151](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)
* Template now use `https_proxy_url` from the right place [#151(https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/151)

## [2.1.0-alpha.3] - 2023-01-31
### Added

* Add attribute k8s.deployment.name to exported logs
[#145](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/145).


## [2.1.0-alpha.2] - 2023-01-30
### Added

* Exposed some configuration of [filelogreciever](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) in `values.yaml` [#146](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/146).

### Changed

* Changed default value `start_at` property of [filelogreciever](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) to `end` [#146](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/146).

## [2.1.0-alpha.1] - 2023-01-25

### Added

* Added telemetry port to kubernetes "ports" [#129](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/129)
* Added `tolerations` config of Logs collection DaemonSet (with default to run on tainted nodes) [#141](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/141)

### Changed
* Upgraded OTEL collector image to [0.3.0](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/releases/tag/0.3.0) which brings following changes
    * OTEL upgraded to 0.69.0
    * Added `filestorage` so it can be in processors

### Fixed

* Added optimizations to Log collector preventing Out of Memory situations [#137](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/137)
    * Added `filelog.storage` to persist checkpoints in persistent storage, not in memory.
    * Setting `max_concurrent_files` to 10 (from default 1024), this is main memory optimization, reducing amount of concurrent log scans
    * Increased default memory limit to `700Mi` (which should be enough for large logs)
    * Having by default `150Mi` difference between OTEL memory limit and Kubernetes memory limit, so that OTEL has enough buffer (this prevents OOMing)

## [2.0.2] - 2023-01-18

### Added

- Initial Helm release.
- Create Error reason as mapping from combination of Event Reason and Type fields [#115](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/115)
- Add support for HTTPS proxies [#117](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/117)
