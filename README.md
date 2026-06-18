# VPS Inspector

English | [简体中文](README.zh-CN.md)

VPS Inspector is a lightweight self-hosted Linux VPS control panel and quality inspection tool. It runs on the VPS being tested and exposes a web UI for real-time system status, VPS line quality checks, and firewall port control.

The project is designed for open-source maintenance: small trusted core, explicit package boundaries, safe command execution, and a frontend that can be embedded into a single Go binary.

## Status

This repository currently contains the first production-shaped skeleton:

- Go HTTP API with graceful shutdown
- Token auth and safe default loopback binding
- Real-time Linux system status from `/proc`
- VPS quality checks for line profile, latency, bandwidth, stability, and IP risk
- Firewall port control through `ufw`, `firewalld`, `nftables`, or `iptables`
- Modular React + TypeScript frontend
- Dockerfile, CI workflow, and project docs

## Platform

VPS Inspector targets Linux VPS environments only. Windows is not a supported runtime.

## Quick Start

```bash
go run ./cmd/vps-inspector
```

Then open:

```text
http://127.0.0.1:8719
```

For remote access, set a token and bind explicitly:

```bash
VPS_INSPECTOR_ADDR=0.0.0.0:8719 VPS_INSPECTOR_AUTH_TOKEN=your-long-random-token go run ./cmd/vps-inspector
```

API requests should include:

```text
Authorization: Bearer your-long-random-token
```

## One-Line Install

After publishing a GitHub release, Linux users can install the panel without Go or Node:

```bash
curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo sh
```

Optional environment variables:

```bash
VPS_INSPECTOR_ADDR=0.0.0.0:8719 VPS_INSPECTOR_AUTH_TOKEN=your-token curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

To override the project root:

```bash
VPS_CONTROL_PANEL_HOME=/vps-control-panel curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

If automatic public IP detection is wrong, set it explicitly:

```bash
VPS_INSPECTOR_PUBLIC_HOST=8.163.47.95 curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/install.sh | sudo -E sh
```

The installer creates `/vps-control-panel`, downloads the latest release binary, writes the environment file under `/vps-control-panel/config`, installs a systemd service, and starts `vps-inspector`.
It prints a ready-to-open access URL with the generated token:

```text
Access URL: http://<server-ip>:8719/<generated-token>
```

Open that URL directly. The frontend stores the token automatically; there is no manual token input in the UI.
Project-owned files are kept under `/vps-control-panel`:

```text
/vps-control-panel/bin/       Binary
/vps-control-panel/config/    Environment config
/vps-control-panel/systemd/   Service file source
/vps-control-panel/data/      Reserved runtime data
/vps-control-panel/logs/      Reserved logs
/vps-control-panel/tmp/       Runtime temp files
```

The only file outside this directory is the systemd entry link at `/etc/systemd/system/vps-inspector.service`, which points back to `/vps-control-panel/systemd/vps-inspector.service`.

## One-Line Uninstall

Remove the service, binary, and configuration:

```bash
curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/uninstall.sh | sudo sh
```

Keep `/vps-control-panel/config` while removing the service, binary, and runtime directories:

```bash
KEEP_CONFIG=1 curl -fsSL https://raw.githubusercontent.com/huanghao256/vps/main/scripts/uninstall.sh | sudo -E sh
```

## Frontend

```bash
cd web
npm install
npm run build
```

The Go server embeds `web/dist` when it exists. During early backend-only development, the server returns a small built-in fallback page.

## Project Layout

```text
.codex/skills/          Codex skill for AI-assisted development
cmd/vps-inspector/       Application entrypoint
internal/agent/          Check orchestration and run lifecycle
internal/checks/         Individual VPS quality checks
internal/config/         Environment-driven configuration
internal/firewall/       Firewall backend detection and port rule operations
internal/httpapi/        HTTP routing, middleware, and handlers
internal/runner/         Safe command runner abstraction
internal/status/         Real-time Linux system status collection
web/                     React + TypeScript frontend
docs/                    Architecture and security notes
```

## AI Development Skill

This repository includes a Codex skill for future AI-assisted development:

```text
.codex/skills/vps-inspector-dev
```

Use it when asking an AI agent to modify this project:

```text
Use $vps-inspector-dev to implement this VPS Inspector change while preserving architecture, security, and validation standards.
```

The skill documents package boundaries, UI expectations, firewall safety rules, install/uninstall behavior, and validation commands.

## Security Defaults

- The server listens on `127.0.0.1:8719` by default.
- Public binding requires setting a strong auth token.
- Check implementations do not execute user-supplied shell strings.
- Firewall operations validate port and protocol before invoking system tools.
- Firewall enable/disable and rule changes usually require root privileges.
- Long-running checks use context timeouts.

See [docs/security.md](docs/security.md) for the security model.

## License

MIT
