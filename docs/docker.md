# Docker

StatLite publishes one multi-platform image for `linux/amd64` and `linux/arm64`:

```text
ghcr.io/pvrlabs/statlite:latest
```

## Run the demo

```bash
docker run --rm \
  -p 127.0.0.1:9090:9090 \
  ghcr.io/pvrlabs/statlite:latest
```

Open <http://127.0.0.1:9090>. The image includes a default configuration that
monitors StatLite itself through `statlite-metrics`; no configuration file is
required.

> [!NOTE]
> The first metrics sample may take up to one polling interval to appear.

The container stores SQLite data at `/data/statlite.sqlite`. Data is ephemeral
unless `/data` is mounted to persistent storage.

## Access and metrics

StatLite has no built-in dashboard or API authentication. The example publishes
port 9090 only on host loopback so the dashboard is not exposed directly to the
network. Use appropriate external access controls before publishing it more
broadly.

Container resource metrics describe the environment visible inside the
container, not the physical macOS host.

The image runs as the non-root `statlite` user and includes CA certificates for
HTTPS targets. Dashboard assets and the SQLite schema are embedded in the
binary.

## Build locally

From the repository root:

```bash
docker build \
  --build-arg VERSION=dev \
  -t statlite:local \
  .
```

Check the embedded version:

```bash
docker run --rm statlite:local --version
```

For local development, the bundled image configuration listens on
`0.0.0.0:9090` inside the container and the example above restricts access at
the host port.

Application examples are available under `examples/`, including the Spring
Boot and Python FastAPI demos.
