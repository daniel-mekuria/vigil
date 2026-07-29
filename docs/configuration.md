# Configuration

StatLite supports Spring Boot Actuator and the canonical fixed StatLite Metrics
JSON profile for small applications that need basic health, traffic, latency,
CPU, and runtime memory monitoring without a full observability stack.
StatLite is not a Prometheus/Grafana replacement.

StatLite loads `statlite.yaml` by default. Override with `--config`:

```bash
./statlite --config examples/actuator.yaml
# or, with an installed binary:
statlite --config /etc/statlite/config.yaml
```

See `examples/` for starter templates (Actuator, StatLite Metrics, multi-target,
self-monitoring), `examples/python-fastapi-demo/` for a runnable FastAPI app,
and `examples/spring-actuator-demo/` for a standalone Spring Boot demo app.

## Server

```yaml
server:
  # Localhost by default; use 0.0.0.0 only behind firewall/VPN/proxy auth.
  listen: "127.0.0.1:9090"
```

StatLite has no built-in dashboard/API authentication. Keep `listen` on loopback unless access is protected externally.

## Storage

```yaml
storage:
  sqlite_path: "./statlite.sqlite"
  # Default is 90 days when omitted; set to 0 for unlimited retention.
  retention_days: 90
```

`sqlite_path` must be writable by the StatLite process. Runtime SQLite files (`*.sqlite`, `*.sqlite-shm`, `*.sqlite-wal`) should not be committed.

### Retention

StatLite keeps SQLite history for **90 days** by default. On startup, and then every 24 hours while running, it deletes poll snapshots older than the configured retention window; related metric samples and collector events are removed automatically.

Set `retention_days: 0` to disable cleanup and keep history indefinitely. Existing SQLite files are pruned on the first startup after retention is enabled unless retention is set to `0`.

## Polling

```yaml
polling:
  interval: "5m"
  timeout: "10s"
```

* `interval` — how often each target is polled (Go duration, required).
* `timeout` — per-poll HTTP timeout (Go duration; default `10s` if omitted).

## Targets

At least one target is required. Names must be unique.

### Spring Boot Actuator

```yaml
targets:
  - name: "my-app"
    actuator_base_url: "https://example.com/actuator"
    auth:
      type: "basic"
      username: "admin"
      password: "change-me"
```

Omit `type` (or treat as the default Actuator target). StatLite polls Actuator health and a fixed set of Micrometer metrics and normalizes them for the dashboard.

Missing optional metrics are handled gracefully: values may appear as `null` or charts may show gaps instead of failing the whole poll.

### Basic Auth

```yaml
auth:
  type: "basic"
  username: "${STATLITE_ACTUATOR_USERNAME}"
  password: "${STATLITE_ACTUATOR_PASSWORD}"
```

Only `basic` is supported in the MVP. Prefer environment variables for credentials, so they are not stored in plaintext YAML. Export them before starting StatLite (or set them with your service manager):

```bash
export STATLITE_ACTUATOR_USERNAME="admin"
export STATLITE_ACTUATOR_PASSWORD="replace-with-a-secret"
statlite --config /etc/statlite/config.yaml
```

StatLite expands environment variables across the entire YAML file once at startup, before it parses the config. Both `$VAR` and `${VAR}` work; unset variables expand to an empty string. Shell-style defaults such as `${VAR:-default}` are not supported. Use `$${` for a literal `${` in the config. Restart StatLite after changing an environment variable.

Plaintext credentials still work as a fallback. Restrict config file permissions when they are present:

```bash
chmod 600 /etc/statlite/config.yaml
chown statlite:statlite /etc/statlite/config.yaml
```

StatLite strips credentials from source endpoints before showing them in the dashboard or API responses.

### StatLite self-monitoring

```yaml
targets:
  - name: "statlite-self"
    type: "statlite-metrics"
    url: "http://127.0.0.1:9090/statlite/metrics"
```

`type: "statlite-metrics"` polls another StatLite (or this process) via
`/statlite/metrics`, the canonical fixed `statlite-metrics/v1` profile. The
same profile is also available to supported external application integrations;
it is not a general metrics protocol.

### StatLite Metrics v1

```yaml
targets:
  - name: "python-demo"
    type: "statlite-metrics"
    url: "http://127.0.0.1:8000/statlite/metrics"
```

Applications using `type: statlite-metrics` expose the fixed
`statlite-metrics/v1` JSON profile. See [StatLite Metrics v1](statlite-metrics-v1.md)
for the complete response format, field semantics, and implementation guidance.
StatLite performs one bounded JSON GET per poll; Basic Auth is not part of v1.

Root `statlite.yaml` uses this pattern so Quick Start works with no extra config. The first poll may fail until the HTTP server is listening; later polls should succeed.

### Host resources

The `statlite-self` target reports the local host CPU, memory, and filesystem
containing StatLite's SQLite database through `/statlite/metrics`. It is the
single target for StatLite's application, process, and host-resource charts.

For a remote application, a central StatLite instance cannot obtain that
machine's host resources unless the application emits the optional host fields
in `statlite-metrics/v1` or another StatLite instance runs on the remote host.

## Dashboard URL state

Selected target and time range are stored in the query string, so you can bookmark a view:

```text
/?target=catalog-api&range=1h
```

## API notes

* `/api/*` is early/internal and not yet a stable public API.
* `/healthz` exposes process version and readiness. Monitored-target poll failures do not mark the process unhealthy; SQLite failure does (`status: "error"`, HTTP 503). See the README section on version and health.

## Example files

| File | Purpose |
|------|---------|
| `statlite.yaml` (repo root) | Default Quick Start — monitors StatLite itself |
| `examples/actuator.yaml` | Single Spring Boot Actuator target with Basic Auth placeholders |
| `examples/statlite.yaml` | Monitor another StatLite instance with `statlite-metrics` |
| `examples/multi-target.yaml` | Illustrative multi-target mix (Actuator + StatLite Metrics + self) |
| `examples/python-fastapi-demo/` | Runnable FastAPI StatLite Metrics v1 demo |
| `examples/spring-actuator-demo/` | Standalone Spring Boot demo app that emits Actuator and Micrometer metrics |

## Systemd

A starter unit is in [statlite.service.example](statlite.service.example). Point `ExecStart` at your binary and config path. Installers do not install this unit automatically.
