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
   - `auto-instrumentation`: enable OBI auto-instrumentation (requires `operator`).
   - `no-logs`, `no-metrics`, `no-events`, `no-prometheus`: exclude individual subsystems.
   - `swo`: forward data to SolarWinds Observability. Requires
     `SOLARWINDS_OTEL_ENDPOINT` and `SOLARWINDS_API_TOKEN` in environment or `skaffold.env`.

```bash
skaffold dev -p operator,obi
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

## Tail Sampling Local Development

### Deploying with local image build

Use the `build-collector` profile together with `tail-sampling` to build a custom
collector image from the sibling `solarwinds-otel-collector-releases` submodule and deploy
it with tail sampling enabled:

```bash
skaffold run -p build-collector,tail-sampling,no-tests --set-values otel.gateway.autoscaler.minReplicas=2
```

Apply the bundled trace generator to exercise the sampling policies:

```bash
kubectl apply -f tests/deploy/test-trace-generator.yaml
```

### Chart assert: minReplicas must be ≥ 2

When `tailSampling.enabled=true` the Helm chart asserts that
`autoscaler.minReplicas >= 2`. This is required for correct trace-ID-based load balancing
across gateway replicas. Any Skaffold profile or values override that sets `minReplicas=1`
will cause `helm template` to abort with an assertion error. On Docker Desktop 2×400 Mi
replicas fit comfortably within the default memory limit.

### Buffer sizing

The `tailSamplingProcessor.num_traces` value must satisfy:

```
num_traces >= expected_new_traces_per_sec × decision_wait_seconds
```

If the buffer is too small, traces are evicted before sampling decisions are made and will
be dropped rather than sampled. Increase `num_traces` in `values.yaml` when running high
load tests.

### OTel attribute keys in policies and test traces

Use the modern OpenTelemetry semantic convention attribute keys. The legacy keys still
appear in old examples but will not match spans produced by current SDKs:

| Legacy (do NOT use) | Modern (use this) |
|---------------------|-------------------|
| `http.target`       | `url.path`        |
| `http.method`       | `http.request.method` |

Ensure both `test-trace-generator.yaml` span attributes and Helm values `policies[].http_status_code_filter` / `string_attribute_filter` use the same convention.

### build-collector profile

The `build-collector` Skaffold profile builds the collector image from
`../solarwinds-otel-collector-releases` (the sibling submodule). This is the mechanism
for testing unreleased collector changes end-to-end before an upstream image is published.
Combine it with any feature profile (e.g. `tail-sampling`) to get a fully integrated
local environment.

## OTel Feature Gates

- Before adding a `featureGates` entry to values.yaml or collector config, verify the gate
  still exists in the target collector version. Gates that are **graduated** (permanently
  enabled) are removed from the binary; referencing them causes pod crash-loops at startup.
  Check the gate's status in the release notes or source of `solarwinds-otel-collector-releases`.
- Example: `processor.tailsamplingprocessor.metadataasattr` was graduated in v0.145.x and
  must not be referenced in any config targeting that version or later.

## Cross-Repo Image Dependencies

- Helm chart changes that rely on new collector components (receivers, processors, etc.)
  require the corresponding `solarwinds-otel-collector-releases` PR to be merged and a new
  image published before live cluster testing is possible.
- During development, test config changes independently with `helm unittest deploy/helm`;
  defer live end-to-end validation until the upstream image PR is merged.

## E2E Cluster (AWS)

- Always use `skaffold run` (not `skaffold dev`) for automated/agent workflows — `skaffold dev`
  blocks the terminal indefinitely.
- Port-forwarding is NOT automatic with `skaffold run`. Set up manually:
  `kubectl --context e2e-cluster port-forward -n test-namespace svc/clickhouse 8123:8123`
- The `no-ebpf` Skaffold profile does NOT exist. Use `no-topology` to skip eBPF/Beyla.
- Docker Hub images (e.g., `clickhouse/clickhouse-server:latest`) may fail to pull due to
  rate limiting on AWS. Mirror to ECR if needed.
- The `timeseries-mock-service` may OOM with default memory limits when processing high
  entity event volumes. If pods restart with OOMKilled, increase memory in
  `tests/deploy/timeseries-mock-service/templates/deployment.yaml`.

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
