# Architecture

VPS Inspector is split into a small set of explicit layers.

## Layers

- `cmd/vps-inspector`: process setup, configuration loading, signal handling.
- `internal/httpapi`: HTTP transport concerns only.
- `internal/agent`: check registry, run lifecycle, result aggregation.
- `internal/checks`: isolated check implementations.
- `internal/status`: Linux system status collection from `/proc` and safe system probes.
- `internal/firewall`: firewall backend detection and port rule operations.
- `internal/runner`: safe execution boundary for any external command usage.

The dependency direction is intentionally one-way:

```text
cmd -> httpapi -> agent -> checks
               -> status
               -> firewall
                         -> runner
```

Checks return structured data and never write directly to HTTP responses. This keeps the API, CLI, and future scheduled jobs able to reuse the same inspection core.

## Run Model

A run is a snapshot of selected checks executed under one request. The API currently keeps recent runs in memory. A storage interface can be added later for SQLite without changing check implementations.
