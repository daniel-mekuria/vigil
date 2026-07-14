# vigil

A self-hosted observability platform with metrics collection, dashboards, and alerting. Lightweight enough for small servers, powerful enough for production infrastructure.

## Overview

Vigil is a single-binary observability solution that collects metrics from your infrastructure, stores them efficiently, and provides a web UI for dashboards and alerting. It's designed to be simple to deploy — no 10-microservice stack, just one binary and a config file.

## Structure

```
cmd/              - CLI entrypoints
internal/         - Core logic: storage, query engine, alerting, web UI
docker/           - Docker build files
Dockerfile        - Container image definition
install.sh        - One-line install script
run.sh            - Development run script
statlite.yaml     - Default configuration
scripts/          - Operational scripts
docs/             - Documentation
examples/         - Configuration examples
```

## Features

- **Metrics collection** — pull-based scraping (Prometheus-compatible) and push-based ingestion
- **Time-series storage** — embedded storage engine, no external database required
- **Dashboards** — built-in web UI for visualization, no Grafana needed
- **Alerting** — threshold and anomaly-based alerts with multi-channel notifications
- **Service discovery** — automatic target discovery for Kubernetes, Docker, static configs
- **Single binary** — everything in one process, minimal resource usage

## Getting Started

### Quick install

```bash
curl -sSL https://github.com/daniel-mekuria/vigil/releases/latest/download/install.sh | bash
```

### Docker

```bash
docker run -d \
  -p 9090:9090 \
  -v ./statlite.yaml:/etc/vigil/config.yaml \
  -v vigil-data:/var/lib/vigil \
  --name vigil \
  daniel-mekuria/vigil:latest
```

### From source

```bash
go build -o vigil ./cmd
./vigil --config statlite.yaml
```

The web UI will be available at `http://localhost:9090`.

## Configuration

```yaml
# statlite.yaml
server:
  listen: ":9090"
  auth:
    enabled: true
    admin_password: "${ADMIN_PASSWORD}"

storage:
  path: "/var/lib/vigil"
  retention_days: 90

scrape:
  interval: 15s
  targets:
    - name: "node"
      type: node_exporter
      host: "localhost:9100"
    - name: "app"
      type: prometheus
      host: "localhost:8080/metrics"

alerts:
  rules:
    - name: high_cpu
      expr: "avg(rate(node_cpu_seconds{mode='idle'}[5m])) < 0.2"
      for: 10m
      severity: warning
      notify: [slack, email]

notify:
  slack:
    webhook: "${SLACK_WEBHOOK}"
  email:
    smtp_host: "smtp.gmail.com"
    recipients: ["oncall@example.com"]
```

## API

Vigil exposes a Prometheus-compatible API:

- `GET /api/v1/query` — instant query
- `GET /api/v1/query_range` — range query
- `GET /metrics` — Vigil's own metrics
- `GET /api/v1/alerts` — current alert state

## Development

```bash
./run.sh              # start in development mode
go test ./...         # run tests
go test -race ./...   # run with race detector
```

## License

MIT
