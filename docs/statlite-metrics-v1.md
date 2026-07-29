# StatLite Metrics v1

`statlite-metrics/v1` is a fixed, language-neutral application metrics profile
for small integrations. It is not a general-purpose metrics protocol: it has
no labels, arbitrary metric definitions, or Prometheus/OpenMetrics syntax.

## Endpoint contract

An application exposes a `GET` endpoint that returns JSON with a successful
2xx HTTP status. The endpoint should produce an inexpensive, process-local
snapshot and should not perform application work such as a database query just
to serve metrics.

A complete response looks like this:

```json
{
  "schema": "statlite-metrics/v1",
  "status": "UP",
  "started_at": "2026-07-27T19:00:00Z",
  "metrics": {
    "requests_total": 1420,
    "responses_404_total": 18,
    "responses_4xx_total": 31,
    "responses_5xx_total": 4,
    "request_duration_seconds_total": 84.31,
    "request_duration_seconds_max": 1.42,
    "cpu_usage": 0.031,
    "runtime_heap_used_bytes": 25165824,
    "uptime_seconds": 1820
  }
}
```

The minimal valid response is:

```json
{
  "schema": "statlite-metrics/v1",
  "status": "UP"
}
```

`schema` and a non-empty `status` are required. `started_at`, `metrics`, and
every individual metric are optional. `status` is application-defined; v1 does
not impose a status enum. `started_at` uses RFC 3339 and is recommended because
it improves restart detection.

## Fields

| Field | Type | Unit | Optional | Semantics |
|---|---|---|---|---|
| `schema` | string | — | No | Must be `statlite-metrics/v1`. |
| `status` | string | — | No | Non-empty application health/status text. |
| `started_at` | string | RFC 3339 timestamp | Yes | Process start time; recommended for restart detection. |
| `metrics` | object | — | Yes | Container for the fixed application metric fields below. |
| `metrics.requests_total` | number | requests | Yes | Monotonic request counter. |
| `metrics.responses_404_total` | number | requests | Yes | Monotonic counter of HTTP 404 responses. |
| `metrics.responses_4xx_total` | number | requests | Yes | Monotonic counter of HTTP 4xx responses, including 404 responses. |
| `metrics.responses_5xx_total` | number | requests | Yes | Monotonic counter of HTTP 5xx responses. |
| `metrics.request_duration_seconds_total` | number | seconds | Yes | Monotonic total request duration. |
| `metrics.request_duration_seconds_max` | number | seconds | Yes | Maximum observed request duration in the application’s current observation window. The observation window is integration-defined in v1 and should be documented by the helper or application. |
| `metrics.cpu_usage` | number | CPU cores | Yes | Process CPU consumption expressed as CPU cores; it may exceed `1.0`. |
| `metrics.runtime_heap_used_bytes` | number | bytes | Yes | Memory currently managed by the language runtime, not total process RSS and not a maximum-memory value. |
| `metrics.uptime_seconds` | number | seconds | Yes | Process uptime. |

### Counters

`requests_total`, `responses_404_total`, `responses_4xx_total`,
`responses_5xx_total`, and `request_duration_seconds_total` are monotonic
counters. They normally reset only when the process restarts. StatLite computes
counter deltas when querying history and does not display negative deltas.

### Gauges

`request_duration_seconds_max`, `cpu_usage`, `runtime_heap_used_bytes`, and
`uptime_seconds` are gauges. Their values describe the current observation of
the process or application rather than a cumulative total.

## Producer guidance

* Collect HTTP counters in middleware so all relevant requests and responses
  are counted consistently.
* Expose current runtime and process values.
* Do not query a database when serving the metrics endpoint.
* Take a consistent snapshot of related values when practical.

See the [FastAPI example](../examples/python-fastapi-demo/) for a complete
Python implementation. It is illustrative; this document is the canonical
contract.

The configured StatLite target name is authoritative. The application should
not provide `target_name`, polling timestamps, or other StatLite-owned metadata.
Unknown fields are ignored for forward compatibility. Invalid optional fields
are skipped and reported as warnings without discarding otherwise valid metrics.

## Multiple processes

In-memory helpers report metrics for one process or worker. Applications using
multiple workers must either aggregate metrics themselves or expose each
process separately. StatLite does not imply that it aggregates application
workers.
