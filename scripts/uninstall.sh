#!/usr/bin/env sh
set -eu

SERVICE_NAME="${SERVICE_NAME:-vps-inspector}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${CONFIG_DIR:-/etc/vps-inspector}"
SERVICE_FILE="${SERVICE_FILE:-/etc/systemd/system/${SERVICE_NAME}.service}"
KEEP_CONFIG="${KEEP_CONFIG:-0}"

require_root() {
  if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root: curl -fsSL ... | sudo sh"
    exit 1
  fi
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

stop_service() {
  if systemctl list-unit-files "${SERVICE_NAME}.service" >/dev/null 2>&1; then
    systemctl disable --now "$SERVICE_NAME" >/dev/null 2>&1 || true
  fi
}

remove_files() {
  rm -f "${INSTALL_DIR}/vps-inspector"
  rm -f "$SERVICE_FILE"

  if [ "$KEEP_CONFIG" = "1" ]; then
    echo "Keeping config directory: ${CONFIG_DIR}"
  else
    rm -rf "$CONFIG_DIR"
  fi
}

reload_systemd() {
  systemctl daemon-reload
  systemctl reset-failed "$SERVICE_NAME" >/dev/null 2>&1 || true
}

main() {
  require_root
  need_cmd systemctl

  stop_service
  remove_files
  reload_systemd

  echo
  echo "VPS Inspector uninstalled."
  if [ "$KEEP_CONFIG" = "1" ]; then
    echo "Config was preserved at ${CONFIG_DIR}."
  fi
}

main "$@"
