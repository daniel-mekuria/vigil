# Changelog

This file summarizes the main user-facing changes in each StatLite release.
Detailed release notes are generated on GitHub from the commit history.

## v0.2.3 — 2026-08-10

- Vendored Chart.js and Orbitron font assets into the dashboard so it renders
  without external CDN access, including in offline and air-gapped environments.
- Bounded historical series queries so long-range dashboard requests avoid
  scanning excessive retained history.
- Improved long-session dashboard behavior, including pausing refresh while the
  browser tab is hidden.
- Limited Spring Actuator responses to 1 MiB to keep collector memory use
  bounded.
- Updated polling defaults and examples to use 30-second intervals, with
  guidance for more resource-conscious production deployments.

## v0.2.2 — 2026-08-07

- Added the StatLite version to the dashboard header.
- Published multi-platform Docker images for `linux/amd64` and `linux/arm64`.
- Improved Python integration documentation and examples.

## v0.2.1

- Added the minimal multi-platform Docker image and Docker deployment documentation.
- Added optional Spring host metrics for CPU, memory, and disk usage.
- Improved the dashboard layout and host-resource visualization.
