# Changelog

All notable changes to Helm charts will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

* Added telemetry port to kubernetes "ports" [#129](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/129)
* Added `tolerations` config of Logs collection DaemonSet (with default to run on tainted nodes) [#141](https://github.com/solarwinds/swi-k8s-opentelemetry-collector/pull/141)

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
