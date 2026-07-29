# FastAPI StatLite Metrics demo

This small FastAPI application exposes a Hello World route and the fixed
`statlite-metrics/v1` JSON profile. It uses the copyable
`statlite_metrics.py` helper and has no Prometheus dependency.

## Run the demo

From this directory, create an environment and install the two runtime
dependencies:

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.txt
```

Start one Uvicorn worker:

```bash
uvicorn app:app --host 127.0.0.1 --port 8000 --workers 1
```

The helper stores counters in memory. Use one application process; multiple
workers would each expose a separate partial view of the application.

In another terminal, generate traffic and inspect the profile:

```bash
curl -s http://127.0.0.1:8000/
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8000/missing
curl -s http://127.0.0.1:8000/statlite/metrics
```

Requests to `/statlite/metrics` are excluded from all application counters and
latency totals, so scraping cannot inflate the traffic being monitored.

## Run StatLite

From the repository's `statlite/` directory, with the FastAPI process still
running:

```bash
go run ./cmd/statlite --config examples/python-fastapi-demo/statlite.yaml
```

Open the dashboard at <http://127.0.0.1:9091>. StatLite polls the FastAPI
endpoint every 10 seconds. The SQLite file is created in the current working
directory.

## Metric semantics

The helper emits cumulative request, 404, 4xx, 5xx, and duration metrics. A
404 is included in both `responses_404_total` and `responses_4xx_total`, as the
v1 contract defines 404 as a 4xx subset. Unhandled application exceptions are
counted as 5xx when FastAPI lets the middleware observe them.

`process_cpu_usage` is process CPU time divided by wall-clock time since the
previous metrics snapshot. Its unit is CPU cores: `1.0` represents one fully
used logical CPU core and values can exceed `1.0`. `uptime_seconds` and
`started_at` describe this application process.

`runtime_heap_used_bytes`, when present, is the current Python-managed
allocation size reported by `tracemalloc`. It is runtime-specific traced
allocation data, not RSS or a container limit. The profile intentionally has
no runtime-neutral maximum-memory metric; it also does not label the peak
traced allocation as a maximum heap. If tracing is unavailable, the optional
runtime-memory field is omitted.

This is a compact single-process reference implementation, not a general
metrics framework or a replacement for a full observability stack.

The helper intentionally omits optional host CPU, memory, and disk fields. A
central StatLite instance cannot infer those resources for this remote
application; emit the optional fields deliberately or run StatLite on that
host when host visibility is needed.

## Check the helper

The helper checks use only Python's standard library and do not require the
FastAPI dependencies:

```bash
python3 -m unittest examples/python-fastapi-demo/test_statlite_metrics.py
```
