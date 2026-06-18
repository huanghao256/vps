#!/usr/bin/env sh
set -eu

REPO="${REPO:-huanghao256/vps}"
SERVICE_NAME="${SERVICE_NAME:-vps-inspector}"
PROJECT_ROOT="${VPS_CONTROL_PANEL_HOME:-/vps-control-panel}"
INSTALL_DIR="${INSTALL_DIR:-${PROJECT_ROOT}/bin}"
CONFIG_DIR="${CONFIG_DIR:-${PROJECT_ROOT}/config}"
SERVICE_DIR="${SERVICE_DIR:-${PROJECT_ROOT}/systemd}"
DATA_DIR="${DATA_DIR:-${PROJECT_ROOT}/data}"
LOG_DIR="${LOG_DIR:-${PROJECT_ROOT}/logs}"
TMP_DIR="${TMP_DIR:-${PROJECT_ROOT}/tmp}"
SERVICE_LINK="${SERVICE_LINK:-/etc/systemd/system/${SERVICE_NAME}.service}"
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

validate_project_root() {
  case "$PROJECT_ROOT" in
    ""|"/"|"/bin"|"/boot"|"/dev"|"/etc"|"/home"|"/lib"|"/lib64"|"/opt"|"/proc"|"/root"|"/run"|"/sbin"|"/sys"|"/tmp"|"/usr"|"/var")
      echo "Unsafe VPS_CONTROL_PANEL_HOME: ${PROJECT_ROOT}" >&2
      exit 1
      ;;
  esac
}

random_token() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 32 | tr -dc 'A-Za-z0-9' | cut -c 1-24
    return
  fi
  date +%s | sha256sum | awk '{print $1}'
}

is_public_ipv4() {
  ip="$1"
  case "$ip" in
    ""|127.*|10.*|192.168.*|169.254.*|0.*|255.*)
      return 1
      ;;
    172.*)
      second="$(printf '%s' "$ip" | awk -F. '{print $2}')"
      if [ "$second" -ge 16 ] 2>/dev/null && [ "$second" -le 31 ] 2>/dev/null; then
        return 1
      fi
      ;;
    100.*)
      second="$(printf '%s' "$ip" | awk -F. '{print $2}')"
      if [ "$second" -ge 64 ] 2>/dev/null && [ "$second" -le 127 ] 2>/dev/null; then
        return 1
      fi
      ;;
  esac
  printf '%s' "$ip" | grep -Eq '^[0-9]{1,3}(\.[0-9]{1,3}){3}$'
}

first_public_ip() {
  for ip in "$@"; do
    ip="$(printf '%s' "$ip" | tr -d '[:space:]')"
    if is_public_ipv4 "$ip"; then
      echo "$ip"
      return 0
    fi
  done
  return 1
}

curl_text() {
  curl -fsS --noproxy '*' --max-time 4 "$1" 2>/dev/null || true
}

public_host() {
  if [ -n "${VPS_INSPECTOR_PUBLIC_HOST:-}" ]; then
    echo "$VPS_INSPECTOR_PUBLIC_HOST"
    return
  fi

  host="$(first_public_ip \
    "$(curl_text https://api.ipify.org)" \
    "$(curl_text https://ifconfig.me/ip)" \
    "$(curl_text https://icanhazip.com)" \
    "$(curl_text https://checkip.amazonaws.com)" \
    "$(curl_text http://100.100.100.200/latest/meta-data/eipv4)" \
    "$(curl_text http://100.100.100.200/latest/meta-data/public-ipv4)" \
    "$(curl_text http://169.254.169.254/latest/meta-data/public-ipv4)" \
  )"
  if [ -n "$host" ]; then
    echo "$host"
    return
  fi

  first_public_ip $(hostname -I 2>/dev/null || true) || true
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
  mkdir -p "$INSTALL_DIR"
  curl -fL "$url" -o "$tmp_dir/vps-inspector.tar.gz"
  tar -xzf "$tmp_dir/vps-inspector.tar.gz" -C "$tmp_dir"
  install -m 0755 "$tmp_dir/vps-inspector" "${INSTALL_DIR}/vps-inspector"
}

prepare_project_root() {
  mkdir -p "$PROJECT_ROOT" "$INSTALL_DIR" "$CONFIG_DIR" "$SERVICE_DIR" "$DATA_DIR" "$LOG_DIR" "$TMP_DIR"
  chmod 0755 "$PROJECT_ROOT" "$INSTALL_DIR" "$SERVICE_DIR" "$DATA_DIR" "$LOG_DIR" "$TMP_DIR"
}

write_config() {
  if [ -z "$TOKEN" ]; then
    TOKEN="$(random_token)"
  fi
  cat > "${CONFIG_DIR}/vps-inspector.env" <<EOF
VPS_CONTROL_PANEL_HOME=${PROJECT_ROOT}
VPS_INSPECTOR_ADDR=${ADDR}
VPS_INSPECTOR_AUTH_TOKEN=${TOKEN}
VPS_INSPECTOR_READ_TIMEOUT=10s
VPS_INSPECTOR_WRITE_TIMEOUT=60s
VPS_INSPECTOR_SHUTDOWN_TIMEOUT=10s
TMPDIR=${TMP_DIR}
EOF
  chmod 0600 "${CONFIG_DIR}/vps-inspector.env"
}

write_service() {
  service_file="${SERVICE_DIR}/${SERVICE_NAME}.service"
  cat > "$service_file" <<EOF
[Unit]
Description=VPS Inspector
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=${CONFIG_DIR}/vps-inspector.env
WorkingDirectory=${PROJECT_ROOT}
ExecStart=${INSTALL_DIR}/vps-inspector
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
  chmod 0644 "$service_file"
  ln -sfn "$service_file" "$SERVICE_LINK"
}

start_service() {
  systemctl daemon-reload
  systemctl enable "$SERVICE_NAME" >/dev/null
  systemctl restart "$SERVICE_NAME"
}

main() {
  require_root
  need_cmd curl
  need_cmd tar
  need_cmd systemctl

  validate_project_root
  prepare_project_root
  install_binary
  write_config
  write_service
  start_service

  echo
  echo "VPS Inspector installed."
  echo "Root: ${PROJECT_ROOT}"
  host="$(public_host)"
  if [ -n "$host" ]; then
    echo "Access URL: http://${host}:${ADDR##*:}/${TOKEN}"
  else
    echo "Access URL: http://<server-ip>:${ADDR##*:}/${TOKEN}"
  fi
  echo "Manage: systemctl status ${SERVICE_NAME}"
}

main "$@"
