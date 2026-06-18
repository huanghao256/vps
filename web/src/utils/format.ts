export function formatBytes(value: number) {
  if (!Number.isFinite(value) || value <= 0) return 'N/A';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size.toFixed(size >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export function formatDuration(seconds: number) {
  if (!Number.isFinite(seconds) || seconds <= 0) return 'N/A';
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return days > 0 ? `${days}d ${hours}h` : `${hours}h ${minutes}m`;
}

export function formatPct(value: number) {
  if (!Number.isFinite(value)) return '0%';
  return `${value.toFixed(value >= 10 ? 0 : 1)}%`;
}

export function formatMbps(value: number) {
  if (!Number.isFinite(value) || value <= 0) return 'N/A';
  return `${value.toFixed(value >= 10 ? 1 : 2)} Mbps`;
}

export function formatMs(value: unknown) {
  const number = typeof value === 'number' && Number.isFinite(value) ? value : 0;
  return number > 0 ? `${number.toFixed(1)} ms` : 'N/A';
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

export function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

export function asNumber(value: unknown) {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0;
}

export function asString(value: unknown) {
  return typeof value === 'string' ? value : '';
}

