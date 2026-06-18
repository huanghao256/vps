#!/usr/bin/env sh
set -eu

REPO="${REPO:-huanghao256/vps}"
SERVICE_NAME="${SERVICE_NAME:-vps-inspector}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${CONFIG_DIR:-/etc/vps-inspector}"
ADDR="${VPS_INSPECTOR_ADDR:-0.0.0.0:8719}"
TOKEN="${VPS_INSPECTOR_AUTH_TOKEN:-}"

require_root() {
  if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root: curl -fsSL ... | sudo sh"
    exit 1
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

random_token() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 24
    return
  fi
  date +%s | sha256sum | awk '{print $1}'
}

download_url() {
  arch="$(detect_arch)"
  if [ "${VERSION:-latest}" = "latest" ]; then
    echo "https://github.com/${REPO}/releases/latest/download/vps-inspector_linux_${arch}.tar.gz"
  else
    echo "https://github.com/${REPO}/releases/download/${VERSION}/vps-inspector_linux_${arch}.tar.gz"
  fi
}

install_binary() {
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT
  url="$(download_url)"

  echo "Downloading ${url}"
  curl -fL "$url" -o "$tmp_dir/vps-inspector.tar.gz"
  tar -xzf "$tmp_dir/vps-inspector.tar.gz" -C "$tmp_dir"
  install -m 0755 "$tmp_dir/vps-inspector" "${INSTALL_DIR}/vps-inspector"
}

write_config() {
  mkdir -p "$CONFIG_DIR"
  if [ -z "$TOKEN" ]; then
    TOKEN="$(random_token)"
  fi
  cat > "${CONFIG_DIR}/vps-inspector.env" <<EOF
VPS_INSPECTOR_ADDR=${ADDR}
VPS_INSPECTOR_AUTH_TOKEN=${TOKEN}
VPS_INSPECTOR_READ_TIMEOUT=10s
VPS_INSPECTOR_WRITE_TIMEOUT=60s
VPS_INSPECTOR_SHUTDOWN_TIMEOUT=10s
EOF
  chmod 0600 "${CONFIG_DIR}/vps-inspector.env"
}

write_service() {
  cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=VPS Inspector
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=${CONFIG_DIR}/vps-inspector.env
ExecStart=${INSTALL_DIR}/vps-inspector
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
}

start_service() {
  systemctl daemon-reload
  systemctl enable --now "$SERVICE_NAME"
}

main() {
  require_root
  need_cmd curl
  need_cmd tar
  need_cmd systemctl

  install_binary
  write_config
  write_service
  start_service

  echo
  echo "VPS Inspector installed."
  echo "URL: http://<server-ip>:${ADDR##*:}"
  echo "Token: ${TOKEN}"
  echo "Manage: systemctl status ${SERVICE_NAME}"
}

main "$@"
