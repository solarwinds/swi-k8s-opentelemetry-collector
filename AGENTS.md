## Development Workflow

### Local iteration with Skaffold

1. Start the default stack:

```bash
skaffold dev
```

   This deploys the collector, Prometheus, and the OTLP mock endpoints into your active
   Kubernetes context.

2. Adjust behaviour with Skaffold profiles. Common combinations:
   - `operator`: install cert-manager, OTEL operator, and discovery collectors.
   - `auto-instrumentation`: enable Beyla auto-instrumentation (requires `operator`).
   - `no-logs`, `no-metrics`, `no-events`, `no-prometheus`: exclude individual subsystems.
   - `swo`: forward data to SolarWinds Observability. Requires
     `SOLARWINDS_OTEL_ENDPOINT` and `SOLARWINDS_API_TOKEN` in environment or `skaffold.env`.

```bash
skaffold dev -p operator,beyla
```

3. To redeploy quickly after manifest or values changes, interrupt Skaffold and rerun the
   command; it will rebuild only the modified artifacts.

### Helm chart maintenance

- Update chart dependencies after modifying `deploy/helm/Chart.yaml`:

```bash
helm dependency update deploy/helm
```

- Keep `deploy/helm/values.yaml` backwards compatible and validate against
  `values.schema.json` using your editor or `helm lint`.

## Testing Instructions

### Helm unit tests

- Run the assertions:

```bash
helm unittest deploy/helm
```

- Refresh golden snapshots when chart output intentionally changes:

```bash
helm unittest -u deploy/helm
```

### Integration tests

- The fastest, fully automated path is the Makefile wrapper:

```bash
make integration-test
```

  This target builds images with Skaffold, deploys the test stack, triggers the
  `integration-test` CronJob, tails pod logs, and tears everything down. Failures export
  pod logs and mock JSON payloads into the `pod-logs/` folder for analysis.

### Performance profiling

- Enable profiling by setting `diagnostics.profiling.enabled=true` in Helm values, port
  forward a collector pod (`kubectl port-forward ... 1777:pprof`), and run
  `go tool pprof http://localhost:1777/debug/pprof/<profile>`.

## Build and Deployment

- Use Skaffold for repeatable builds:

```bash
skaffold build --file-output=/tmp/tags.json
```

- Delete the stack and images when finished:

```bash
skaffold delete
```

- For chart releases, bump `version` (and `appVersion` if images change) in
  `deploy/helm/Chart.yaml`, run `helm dependency update deploy/helm`, and commit the updated
  `Chart.lock`.

## Pull Request Guidelines

- Branch from `master` and keep your feature branch up to date with upstream changes.
- Document chart-visible changes in the `Unreleased` section of `deploy/helm/CHANGELOG.md`.
- Run the relevant checks locally (at minimum `helm unittest deploy/helm` and
  `make integration-test`) before requesting review.
- Keep test framework updates (Python tooling, mock services) separate from functional
  chart or collector changes.

## Debugging and Troubleshooting

- Missing CRDs during Skaffold deploy often indicate a stale Helm repository cache. Repair
  with `helm repo update` and rerun the deploy.
- If chart downloads fail with `no cached repo found`, add the legacy stable repository:

```bash
helm repo add stable https://charts.helm.sh/stable --force-update
```

- Integration test failures:
  - Inspect streamed pod logs in the terminal; the Makefile background log tail stops when
    the job finishes.
  - Review artifacts in `pod-logs/` (collector pods, mock JSON exports) for regression
    analysis.
  - Reproduce individual CronJob executions with
    `kubectl create job --from=cronjob/integration-test debug-run -n test-namespace`.
- To inspect exported telemetry manually, curl the mock endpoints:

```bash
curl http://localhost:8088/logs.json
curl http://localhost:8088/metrics.json
```

- When working against remote clusters, ensure the registry login has not expired and the
  namespace referenced by `TEST_CLUSTER_NAMESPACE` exists before deploying.
