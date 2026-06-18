#!/usr/bin/env sh
set -eu

SERVICE_NAME="${SERVICE_NAME:-vps-inspector}"
PROJECT_ROOT="${VPS_CONTROL_PANEL_HOME:-/vps-control-panel}"
INSTALL_DIR="${INSTALL_DIR:-${PROJECT_ROOT}/bin}"
CONFIG_DIR="${CONFIG_DIR:-${PROJECT_ROOT}/config}"
SERVICE_DIR="${SERVICE_DIR:-${PROJECT_ROOT}/systemd}"
DATA_DIR="${DATA_DIR:-${PROJECT_ROOT}/data}"
LOG_DIR="${LOG_DIR:-${PROJECT_ROOT}/logs}"
TMP_DIR="${TMP_DIR:-${PROJECT_ROOT}/tmp}"
SERVICE_FILE="${SERVICE_FILE:-${SERVICE_DIR}/${SERVICE_NAME}.service}"
SERVICE_LINK="${SERVICE_LINK:-/etc/systemd/system/${SERVICE_NAME}.service}"
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

validate_project_root() {
  case "$PROJECT_ROOT" in
    ""|"/"|"/bin"|"/boot"|"/dev"|"/etc"|"/home"|"/lib"|"/lib64"|"/opt"|"/proc"|"/root"|"/run"|"/sbin"|"/sys"|"/tmp"|"/usr"|"/var")
      echo "Unsafe VPS_CONTROL_PANEL_HOME: ${PROJECT_ROOT}" >&2
      exit 1
      ;;
  esac
}

stop_service() {
  if systemctl list-unit-files "${SERVICE_NAME}.service" >/dev/null 2>&1; then
    systemctl disable --now "$SERVICE_NAME" >/dev/null 2>&1 || true
  fi
}

remove_files() {
  rm -f "${INSTALL_DIR}/vps-inspector"
  rm -f "$SERVICE_LINK"
  rm -f "$SERVICE_FILE"

  if [ "$KEEP_CONFIG" = "1" ]; then
    echo "Keeping config directory: ${CONFIG_DIR}"
    rm -rf "$INSTALL_DIR" "$SERVICE_DIR" "$DATA_DIR" "$LOG_DIR" "$TMP_DIR"
  else
    rm -rf "$PROJECT_ROOT"
  fi
}

reload_systemd() {
  systemctl daemon-reload
  systemctl reset-failed "$SERVICE_NAME" >/dev/null 2>&1 || true
}

main() {
  require_root
  need_cmd systemctl
  validate_project_root

  stop_service
  remove_files
  reload_systemd

  echo
  echo "VPS Inspector uninstalled."
  if [ "$KEEP_CONFIG" = "1" ]; then
    echo "Config was preserved at ${CONFIG_DIR}."
  else
    echo "Removed project root: ${PROJECT_ROOT}"
  fi
}

main "$@"
