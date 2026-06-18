---
name: vps-inspector-dev
description: Use when developing, reviewing, refactoring, documenting, or extending the VPS Inspector repository, including Go backend APIs, Linux status collection, VPS quality checks, firewall port control, React/TypeScript UI, install/uninstall scripts, release workflow, or project coding standards.
---

# VPS Inspector Dev

Use this skill to keep VPS Inspector maintainable while adding features. Prefer small, scoped changes that preserve the existing package boundaries.

## Project Shape

- Runtime target: Linux VPS only. Do not add Windows runtime compatibility unless explicitly requested.
- Backend: Go, no unnecessary third-party dependencies.
- Frontend: React + TypeScript + Vite.
- Deployment: single Go binary can embed built frontend assets; Linux install is handled by POSIX shell scripts and systemd.
- Installed project-owned files must live under `/vps-control-panel` by default. Keep only the required systemd entry link under `/etc/systemd/system`.
- User-facing UI text should be Chinese unless the surrounding file is English documentation.

## Backend Boundaries

- `cmd/vps-inspector`: process setup only, config load, logger, signal handling.
- `internal/httpapi`: HTTP routing, auth, request/response handling. Do not put business logic here.
- `internal/agent`: check orchestration, run lifecycle, result aggregation.
- `internal/checks`: VPS quality checks implementing `agent.Check`.
- `internal/status`: Linux realtime status collection from `/proc` and minimal safe system probes.
- `internal/firewall`: firewall backend detection and port operations.
- `internal/config`: environment-driven config and safety validation.
- `internal/runner`: reusable safe command runner boundary.

When adding a backend feature, place domain logic in the domain package first, then expose it through `httpapi`.

## Go Code Style

- Add Go doc comments for exported packages, types, constructors, functions, methods, and constants.
- Comments should describe the contract, boundary, or non-obvious behavior; avoid restating the identifier name.
- Prefer unexported types and helpers unless another package genuinely needs them.
- Keep request validation close to the API boundary, and keep domain invariants inside the domain package.
- Use `context` for network, command, and long-running operations.
- Do not add broad dependencies for small standard-library problems.

## VPS Check Pattern

For a new VPS quality check:

1. Add a focused file under `internal/checks`.
2. Implement `agent.Check`.
3. Return structured `agent.Result` with stable `details` keys for the frontend.
4. Register it in `internal/checks/registry.go`.
5. Keep check timeouts bounded with `context`.
6. Prefer lightweight probes that do not waste user bandwidth.

Use Chinese summaries for user-facing check result text where practical.

## Firewall Safety

Firewall operations are high-risk. Keep these rules:

- Accept only structured inputs such as `{ port, protocol }`.
- Validate port range `1-65535`.
- Support only `tcp` and `udp` unless explicitly designed otherwise.
- Never execute user-supplied shell strings.
- Use `exec.CommandContext` with argument slices.
- Remember firewall enable/disable and rule changes generally require root.
- Keep install/uninstall scripts clear about systemd and root behavior.

## Linux Status Rules

System status is Linux-only:

- Use `/proc` and common Linux commands sparingly.
- Return clear errors when `/proc` is unavailable.
- Keep status snapshots fast enough for polling.
- Avoid adding expensive scans to the realtime status endpoint.

## Frontend Boundaries

- `web/src/App.tsx`: only top-level layout and page switching.
- `web/src/pages`: page-level modules such as system status, VPS detection, port control.
- `web/src/components`: reusable display components.
- `web/src/api.ts`: API client only.
- `web/src/types.ts`: shared API-facing types.
- `web/src/utils`: formatting and pure helpers.

Do not put page business logic, API calls, and large JSX all into `App.tsx`. If a page grows, split feature-specific components before it becomes hard to review.

## UI Expectations

- The three primary entries are `系统状态`, `VPS检测`, and `端口控制`.
- `系统状态` should show realtime Linux status: CPU, memory, disk, uptime, network traffic, IP, connections.
- `VPS检测` should emphasize `线路`, `延迟`, `带宽`, `稳定性`, `IP风控风险` with readable Chinese metrics and process visuals.
- `端口控制` should show firewall state, existing rules, and explicit add/delete controls.
- Do not expose a manual token editor in the UI. The installer prints `http://host:port/token`; the frontend reads the first URL segment and stores it locally. Keep the URL stable so refreshes continue to work.
- Avoid raw JSON in user-facing panels. Convert details into labels, scores, badges, cards, rings, tables, or process tracks.

## Install, Uninstall, Release

- `scripts/install.sh`: POSIX `sh`, root check, create `/vps-control-panel`, download release binary, write env file under `config/`, write service source under `systemd/`, symlink systemd entry, enable and restart service, print a tokenized access URL. Public IP detection must filter private ranges and support `VPS_INSPECTOR_PUBLIC_HOST` override.
- `scripts/uninstall.sh`: POSIX `sh`, stop/disable service, remove systemd link and `/vps-control-panel` by default, support `KEEP_CONFIG=1`.
- Release workflow builds frontend, copies `web/dist` into `internal/httpapi/webdist`, then builds Linux `amd64` and `arm64` binaries.
- Do not commit `web/dist`, `web/node_modules`, `bin`, `.gocache`, or `.npm-cache`.

## Validation

Before finishing meaningful code changes, run the relevant checks:

```powershell
$env:GOCACHE='d:\phpstudy_pro\vps\.gocache'; go test ./...
$env:GOCACHE='d:\phpstudy_pro\vps\.gocache'; go vet ./...
npm.cmd run build
```

On Linux/macOS, use the equivalent shell syntax:

```bash
go test ./...
go vet ./...
cd web && npm run build
```

If frontend dependencies are missing, run `npm install` or `npm ci` in `web/`. If sandboxed npm cache fails, use a project-local cache.

## Git Hygiene

- Do not push, tag, or publish releases unless the user explicitly asks.
- Keep commits focused.
- Do not revert user changes.
- Add docs when changing install, uninstall, release, or user-facing behavior.
