# Expand StatLite Metrics v1 Plan

Issue: #2
Status: Active
Created: 2026-07-29
Archived:

## Summary

Extend the fixed `statlite-metrics/v1` profile from application/process-only
metrics to application, process, and five optional host metrics. Make StatLite emit
and collect that same profile for its own dashboard, retire the separate
`statlite-health` dashboard protocol, and organize charts by the capabilities
actually present in a target response.

The repository currently has no `type: host` collector despite the proposal
referring to it as experimental. This plan adds the narrowly scoped local host
collector needed for the stated acceptance criteria; it will use the same
normalized host metric keys as embedded host fields and will not create a
synthetic target for embedded fields.

## Source Constraints

The following issue requirements must remain true throughout implementation:

* Keep `statlite-metrics/v1` fixed, small, language-neutral, label-free, and
  non-Prometheus; do not add arbitrary definitions, custom metrics, a Go-only
  target type, or a second StatLite-specific metrics protocol.
* Host metrics are lightweight operational estimates, not accounting-grade
  measurements. Prefer simple, dependency-light collection and graceful
  omission over complex platform-specific accuracy. More precise cgroup, CPU
  interval, filesystem, and aggregation semantics may be refined later. The
  purpose is to spot high CPU, filling RAM, or nearly full disk—not resource
  billing or forensic analysis.
* Rename producer field `cpu_usage` to `process_cpu_usage` before v1
  stabilizes. It is process CPU consumption in CPU cores, so it may exceed
  `1.0`; `host_cpu_usage` is a fraction of visible host capacity in `[0.0,
  1.0]`. Normalize their units as `cores` and `ratio`, respectively.
* Add optional `host_cpu_usage`, `host_memory_used_bytes`, and
  `host_memory_total_bytes`, plus `host_disk_used_bytes` and
  `host_disk_total_bytes`. Define host values as the execution environment
  visible to the producer, including container-visible limits/accounting.
  Derive memory and disk usage percentages in StatLite from their respective
  used/total byte pairs.
* Disk fields describe one producer-selected relevant filesystem, never every
  disk, device, volume, or mount. The byte values must be finite,
  non-negative, from the same selected filesystem, and used must not exceed
  total. Memory used must likewise not exceed memory total. A memory or disk
  percentage exists only when both fields are valid and total is positive.
  Missing or invalid disk fields do not invalidate other samples.
  Do not add filesystem paths, mount names, device labels, arrays, per-volume
  metrics, or a producer-supplied percentage to the profile.
* StatLite self-monitoring selects the filesystem containing its SQLite
  database. The FastAPI example defaults to the filesystem containing
  `Path.cwd()`; applications with a dedicated data directory should select it
  instead. In containers, disk values describe the filesystem visible inside
  the container on a best-effort basis; do not promise perfect interpretation
  of every cgroup configuration. Keep disk I/O counters and their
  delta/reset/rate semantics out of this issue.
* All profile metric fields remain optional. Keep valid samples if optional
  fields are absent or invalid, record descriptive collector warnings for
  invalid values, and store normalized samples through the existing common
  storage path.
* StatLite must expose `GET /statlite/metrics` as a valid profile response,
  including available HTTP metrics, process CPU, Go runtime memory, start time,
  uptime, and host CPU/memory/disk capacity. Its own metrics request must not
  inflate its application HTTP counters.
* The default target is exactly:

  ```yaml
  targets:
    - name: "statlite-self"
      type: "statlite-metrics"
      url: "http://127.0.0.1:9090/statlite/metrics"
  ```

* Self-monitoring must use the normal `statlite-metrics` collector. Remove
  `type: statlite-health`, its `type: statlite` alias, and the specialized
  health collector unless a documented compatibility audit finds concrete
  external adoption that requires a time-bounded alternative.
* `/healthz` may remain only as lightweight readiness: process availability,
  SQLite usability, version, and HTTP `200`/`503`. It is not a dashboard
  protocol and must not duplicate application, process, or host chart data.
* A local `type: host` collector is allowed for zero-integration local
  monitoring. It emits only the shared normalized host keys and displays only
  Host resources. Embedded host fields remain attached to their application
  target; never synthesize another host target.
* Dashboard sections are capability based: **Application**, **Process**, and
  **Host resources**. A StatLite Metrics target may display all three.
* Update `docs/statlite-metrics-v1.md`, `docs/product.md`,
  `docs/configuration.md`, affected configurations, the Python FastAPI
  example, and only necessary README summaries/links. Describe
  `statlite-metrics/v1` as StatLite's own canonical profile and `/healthz` as
  readiness only.

## Progress

Current Phase: Phase 2 — StatLite endpoint and collector migration
Current Chunk: Chunk 2.2 — Retire the legacy dashboard collector contract
Status: [x] Complete

### Phase Checklist

- [x] Phase 1 — Contract and normalized data
- [x] Phase 2 — StatLite endpoint and collector migration
- [ ] Phase 3 — Capability-based dashboard
- [ ] Phase 4 — Config, examples, documentation, and regression verification

### Blockers

- None. The repository-only usage audit currently finds `statlite-health` and
  `statlite` only in StatLite source, tests, and bundled examples; no known
  external compatibility dependency is recorded. Reconfirm this against the
  issue/release context before deleting the paths.

## Phases

### Phase 1 — Contract and normalized data

Goal: Specify and normalize the renamed process metric and five optional host
metrics, including safe validation, query shaping, derived memory/disk
percentages, and the local-host source.

Boundaries: Do not change the self endpoint or dashboard layout in this phase.
Do not introduce duplicate source-specific host keys, rollups, labels, or a
general system-metrics framework.

Exit Criteria: The profile parser and local host collector produce the shared
normalized host keys; series/downsampling can return host values and derived
memory/disk percentages with correct missing/invalid behavior.

### Phase 2 — StatLite endpoint and collector migration

Goal: Make StatLite dogfood `/statlite/metrics`, make `/healthz` readiness
only, and remove the separate health target contract after the compatibility
decision.

Boundaries: Preserve the existing common monitor/storage flow and readiness
semantics. Do not make target poll failure a process-readiness failure.

Exit Criteria: Default self-monitoring uses `statlite-metrics`; no production
configuration or collector construction path selects `statlite-health` or its
alias; StatLite's profile response is independently valid and scrape-safe.

### Phase 3 — Capability-based dashboard

Goal: Present application, process, and host data only when each capability is
available, without separating an application's embedded host values into a
second target.

Boundaries: Keep the dashboard collector-neutral and small. Retain current
range/downsampling behavior and do not add custom chart configuration.

Exit Criteria: Application/process/host sections have correct conditional
visibility, use the shared series fields, and show no-data states without
turning missing optional metrics into zero.

### Phase 4 — Config, examples, documentation, and regression verification

Goal: Publish one canonical profile and leave runnable examples/configuration
consistent with the implementation.

Boundaries: Keep `/healthz` documentation limited to operational readiness;
do not broaden supported integrations.

Exit Criteria: Docs/examples contain no preferred healthz monitoring path,
FastAPI emits the expanded profile, all example configs validate, and the full
test suite passes.

## Execution

### Chunk 1.1 — Lock the profile and host collection boundary

Status: [x] Complete

Preconditions:

- Review issue #2/release context for any concrete users of `statlite-health`
  or `type: statlite`; record any evidence in this plan before retaining
  compatibility. With no evidence, remove both paths in Chunk 2.2.
- Choose and document a dependency-light host sampler that reports the Linux
  host/container view available to the process and degrades by omitting
  unavailable metrics with warnings. A simple stateful CPU sampler may omit
  or report zero for its first sample; slightly different readings from
  separate sampler instances are acceptable. Do not add a broad monitoring
  dependency merely for these five fixed values. The sampler must accept a
  filesystem path and obtain capacity for the filesystem containing that path
  without exposing that path in a producer response.

Checklist:

- [x] (design) Amend `docs/statlite-metrics-v1.md`'s response and field table:
  replace `cpu_usage`, add all five host fields including
  `host_disk_used_bytes` and `host_disk_total_bytes`, define units/ranges,
  single-filesystem selection defaults, and container-visible semantics, and
  list memory/disk percentages as StatLite consumer-derived values rather than
  producer fields. Explicitly reserve only conceptual room for future disk I/O
  fields; do not define or implement their semantics.
- [x] (impl) Update `internal/collector/statlite_metrics.go` and
  `internal/collector/statlite_metrics_collector.go` to decode
  `process_cpu_usage` and the host fields, map them to `process_cpu_usage`,
  `host_cpu_usage`, `host_memory_used_bytes`, `host_memory_total_bytes`,
  `host_disk_used_bytes`, and `host_disk_total_bytes`; reject
  non-finite/negative values, host CPU values above `1.0`, memory used values
  above memory total, and disk used values above disk total when each pair is
  valid. Use unit `cores` for `process_cpu_usage` and `ratio` for
  `host_cpu_usage`.
- [x] (impl) Add the fixed local `host` target type in `internal/config` and
  `internal/app`, with a collector under `internal/collector` that emits only
  the same five normalized host keys. Give it no URL/auth requirement and no
  application/process samples. It must use the same dependency-light sampler,
  selecting the filesystem containing StatLite's configured SQLite database
  when easily available and otherwise falling back to the current working
  directory. Display it as a local endpoint such as `local host`, never a
  filesystem path.
- [x] (impl) Extend `internal/storage/types.go`, `series.go`, and
  `downsample.go` so each raw host sample first derives a memory/disk percentage
  only when its own used/total pair is valid and total is positive. Return the
  raw bytes and derived percentage gauges; when downsampling, average those
  already-derived percentage gauges rather than combining used/total values
  from different observations. Preserve nil for incomplete or invalid pairs
  and keep raw samples as the source of truth without storing percentages.
- [x] (test) Cover complete/sparse/invalid profile host and disk fields,
  renamed process CPU above `1.0`, host CPU range rejection, negative and
  non-finite disk values, used above total for both memory and disk, zero
  total, a missing byte of a pair, shared keys from the local host collector,
  and host/disk values through aggregation/downsampling. Assert percentages
  are calculated per sample and downsampling averages those gauges rather than
  combining unrelated byte observations.
- [x] (verify) Run focused collector/storage/config tests, then `go test ./...`.

Done Criteria:

- The collector accepts a minimal profile and any valid subset of metrics.
- Local and remote host and disk values have identical normalized key, kind,
  and unit.
- No stored sample is a derived host-memory or disk percentage; percentages
  are calculated only for query/dashboard output, and downsampling averages
  only percentages derived from individual valid samples.

### Chunk 2.1 — Emit StatLite's canonical profile

Status: [x] Complete

Preconditions:

- Chunk 1.1 is complete and the host sampler can be reused by the server
  without leaking collector response shapes into server code.

Checklist:

- [x] (impl) Add `/statlite/metrics` route and a profile-shaped server
  response in `internal/server`. Populate `schema`, non-empty status,
  `started_at`, and available existing request counters/latency, Go
  `runtime.MemStats.Alloc`, process CPU in CPU cores (remove the current
  division by `runtime.NumCPU()`), host CPU/memory values, and disk used/total
  bytes from the shared sampler for the filesystem containing the configured
  SQLite database. Omit disk fields and record an internal warning/log when
  that filesystem cannot be determined or sampled; do not fail this endpoint
  or readiness.
- [x] (impl) Refactor request accounting so `/statlite/metrics` is excluded
  from all application request/error/latency counters while normal dashboard
  and readiness requests retain the defined application accounting behavior.
  Add bounded request-duration total/max instrumentation if it is not already
  present in the shared server middleware.
- [x] (impl) Reduce `internal/server/health.go` to version plus readiness and
  SQLite status only. Retain 200/503 semantics, but remove the nested
  `statlite` metrics/polling payload and its dashboard-specific data.
- [x] (test) Assert the endpoint's schema, optional fields, process CPU units,
  host values when available, disk fields when sampling succeeds, database
  location selection, disk omission when unavailable, and exclusion of its
  own scrape. Assert `/healthz` still exposes version/readiness and no metrics
  protocol or duplicate disk payload.
- [x] (verify) Run `go test ./internal/server ./internal/collector`.

Done Criteria:

- A direct `GET /statlite/metrics` is consumable by
  `StatliteMetricsCollector` without special cases.
- `/healthz` cannot be mistaken for a dashboard metrics endpoint.

### Chunk 2.2 — Retire the legacy dashboard collector contract

Status: [x] Complete

Preconditions:

- Chunk 2.1 is complete.
- Compatibility audit in Chunk 1.1 records no evidence requiring a temporary
  migration path; otherwise stop and obtain explicit scope before retaining
  aliases.

Checklist:

- [x] (impl) Remove `TargetTypeStatliteHealth` and `TargetTypeStatliteLegacy`
  validation/display paths from `internal/config`, and remove legacy collector
  construction from `internal/app/bootstrap.go`.
- [x] (impl) Delete `internal/collector/statlite.go` and its specialized tests;
  remove only tests/config fixtures that exercise the deleted protocol.
- [x] (impl) Change root `statlite.yaml` to the required
  `statlite-self`/`statlite-metrics` `/statlite/metrics` target and update
  `examples/statlite.yaml` and `examples/multi-target.yaml` to eliminate
  healthz targets.
- [x] (test) Update config/bootstrap tests to reject removed types, construct
  the normal StatLite Metrics collector for self-monitoring, and load every
  changed example configuration.
- [x] (verify) Search the code/config/docs tree for `statlite-health`, legacy
  `type: statlite`, and `/healthz` target URLs; remaining `/healthz` references
  must be operational-readiness documentation/tests only.

Done Criteria:

- StatLite has exactly one dashboard integration contract for itself and
  external profile producers: `statlite-metrics/v1`.
- No permanent compatibility alias remains without recorded adoption evidence.

### Chunk 3.1 — Render capability-based application, process, and host sections

Status: [ ] Not started

Preconditions:

- Chunks 1.1 and 2.2 are complete and `/api/series` contains the new host
  fields.

Checklist:

- [ ] (impl) Restructure `internal/dashboard/static/index.html` into
  Application (requests, HTTP errors, latency), Process (runtime memory and
  process CPU), and Host resources with three independent charts: Host CPU,
  RAM, and Disk.
- [ ] (impl) Render each section only when its corresponding series has data;
  preserve explicit no-data notes inside displayed charts and never coerce
  absent values to zero.
- [ ] (impl) Bind Host CPU and RAM charts to `host_cpu_usage` and derived host
  memory percentage/bytes. Bind the separate Disk chart to derived disk
  percentage and raw `host_disk_used_bytes`/`host_disk_total_bytes`, including
  a current-value display such as `Disk — 18.4 GB / 40 GB · 46%`. Show Disk
  only for a valid used/total pair; never treat missing values as zero. A
  `statlite-metrics` target may render all sections; `host` renders only Host
  resources.
- [ ] (test) Extend server/API tests for serialized host/disk series and add
  lightweight static/dashboard assertions for section identifiers, chart data
  bindings, raw-byte display, conditional Disk visibility, derived percentage,
  and removal of legacy type help.
- [ ] (verify) Manually exercise a profile target with all metrics, a sparse
  profile target, and a local `host` target at 1h and a downsampled long range.

Done Criteria:

- One target response controls all of its visible capabilities.
- Host data embedded in an app response appears under that app's selected
  target and no extra target is created.

### Chunk 4.1 — Update producers, guidance, and release-quality checks

Status: [ ] Not started

Preconditions:

- Chunks 1.1 through 3.1 are complete.

Checklist:

- [ ] (impl) Update the FastAPI helper and README to emit
  `process_cpu_usage`, runtime-managed memory, and best-effort host CPU/memory
  plus disk capacity from `shutil.disk_usage(Path.cwd())`. Document that the
  working-directory selection is an example default, not a protocol rule, and
  production applications should select their important writable/persistent
  data directory when appropriate. Retain scrape exclusion and omit values it
  cannot safely determine.
- [ ] (impl) Update `docs/statlite-metrics-v1.md`, `docs/product.md`,
  `docs/configuration.md`, and necessary README text to name the profile as
  canonical for StatLite and external integrations, describe `type: host` as
  local-only, and limit `/healthz` to readiness/version/SQLite semantics.
- [ ] (impl) Update all relevant example configs and target-type summaries;
  remove claims that `/healthz` is a self-monitoring integration.
- [ ] (test) Add/adjust FastAPI example checks if the repository's test setup
  can run them without new heavy tooling; otherwise perform the documented
  curl response validation. Verify disk used/total values are emitted when
  available, describe the same filesystem, and are omitted rather than making
  the response invalid on failure; record the result in the chunk completion
  note.
- [ ] (verify) Run `go test ./...`, validate every bundled YAML example with
  existing config tests, run the FastAPI example response through the normal
  StatLite Metrics collector test fixture, and inspect the dashboard at narrow
  and long ranges.
- [ ] (verify) Update this plan's status, checklist items, and Outcome with
  the compatibility decision and any intentional platform limitations before
  archiving.

Done Criteria:

- Documentation and runnable examples present one canonical profile and no
  dashboard `/healthz` integration.
- All acceptance criteria below are demonstrated by tests or recorded manual
  checks.

## Acceptance Checklist

- [ ] `statlite-metrics/v1` accepts and stores all five optional host fields.
- [ ] `process_cpu_usage` replaces `cpu_usage` according to the recorded
  compatibility decision.
- [ ] Host memory percentage is derived correctly from used and total bytes.
- [ ] Disk usage percentage is derived correctly from used and total bytes.
- [ ] StatLite emits a valid `statlite-metrics/v1` response of its own.
- [ ] StatLite self-monitoring reports the filesystem containing its SQLite
  database when sampling succeeds.
- [ ] Default self-monitoring uses `type: statlite-metrics`.
- [ ] StatLite application, process, and host charts use the normal metrics
  collector.
- [ ] The specialized `statlite-health` path is removed or has an explicit,
  evidence-based deprecation record.
- [ ] `/healthz` is readiness/version/SQLite only.
- [ ] Local `type: host` and embedded host metrics share normalized keys.
- [ ] Local `type: host` and embedded disk metrics use identical normalized
  keys.
- [ ] The FastAPI example defaults to the filesystem containing its working
  directory.
- [ ] Disk is shown as a separate Host resources chart.
- [ ] Missing or invalid disk metrics do not invalidate other profile metrics.
- [ ] The profile exposes no filesystem paths, mount labels, devices, or
  multiple-disk collections.
- [ ] Documentation/examples no longer recommend `/healthz` for dashboard
  self-monitoring.

## Outcome

Required before archive.

- What was delivered:
- What changed from the original plan:
- What was intentionally skipped:
