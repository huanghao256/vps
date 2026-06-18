export type PageKey = 'status' | 'detection' | 'ports';

export type CheckInfo = {
  id: string;
  name: string;
  description: string;
  category: string;
};

export type CheckResult = {
  checkId: string;
  status: 'pass' | 'warn' | 'fail' | 'skip';
  score: number;
  summary: string;
  details?: Record<string, unknown>;
  startedAt: string;
  endedAt: string;
  error?: string;
};

export type Run = {
  id: string;
  checkIds: string[];
  status: string;
  score: number;
  startedAt: string;
  endedAt?: string;
  results: CheckResult[];
};

export type SystemSnapshot = {
  capturedAt: string;
  health: { label: string; loadPct: number };
  cpu: { cores: number; usagePct: number; load1: number; load5: number; load15: number };
  memory: { totalBytes: number; usedBytes: number; availableBytes: number; usagePct: number };
  disk: { mount: string; totalBytes: number; usedBytes: number; freeBytes: number; usagePct: number };
  uptime: { systemSeconds: number; appSeconds: number };
  network: {
    ipv4: string;
    receivedBytes: number;
    transmittedBytes: number;
    receiveMbps: number;
    transmitMbps: number;
  };
  connections: { tcp: number; udp: number };
  system: { hostname: string; os: string; threads: number; processes: number };
};

export type FirewallRule = {
  port: number;
  protocol: string;
  action: string;
  source: string;
};

export type FirewallSnapshot = {
  backend: string;
  available: boolean;
  enabled: boolean;
  message: string;
  rules: FirewallRule[];
};

export type PortRuleInput = {
  port: number;
  protocol: 'tcp' | 'udp';
};

