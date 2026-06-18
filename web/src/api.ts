import type { CheckInfo, FirewallSnapshot, PortRuleInput, Run, SystemSnapshot } from './types';

type RequestOptions = {
  token: string;
  method?: 'GET' | 'POST' | 'DELETE';
  body?: unknown;
};

export async function getSystemStatus(token: string): Promise<SystemSnapshot> {
  return request<SystemSnapshot>('/api/v1/status', { token });
}

export async function listChecks(token: string): Promise<CheckInfo[]> {
  const data = await request<{ checks: CheckInfo[] }>('/api/v1/checks', { token });
  return data.checks;
}

export async function listRuns(token: string): Promise<Run[]> {
  const data = await request<{ runs: Run[] }>('/api/v1/runs', { token });
  return data.runs;
}

export async function createRun(token: string, checkIds: string[]): Promise<Run> {
  return request<Run>('/api/v1/runs', { token, method: 'POST', body: { checkIds } });
}

export async function getFirewall(token: string): Promise<FirewallSnapshot> {
  return request<FirewallSnapshot>('/api/v1/firewall', { token });
}

export async function enableFirewall(token: string): Promise<FirewallSnapshot> {
  return request<FirewallSnapshot>('/api/v1/firewall/enable', { token, method: 'POST', body: {} });
}

export async function disableFirewall(token: string): Promise<FirewallSnapshot> {
  return request<FirewallSnapshot>('/api/v1/firewall/disable', { token, method: 'POST', body: {} });
}

export async function addFirewallRule(token: string, rule: PortRuleInput): Promise<FirewallSnapshot> {
  return request<FirewallSnapshot>('/api/v1/firewall/rules', { token, method: 'POST', body: rule });
}

export async function deleteFirewallRule(token: string, rule: PortRuleInput): Promise<FirewallSnapshot> {
  return request<FirewallSnapshot>('/api/v1/firewall/rules', { token, method: 'DELETE', body: rule });
}

async function request<T>(path: string, options: RequestOptions): Promise<T> {
  const headers = new Headers({ 'Content-Type': 'application/json' });
  if (options.token.trim() !== '') {
    headers.set('Authorization', `Bearer ${options.token.trim()}`);
  }

  const response = await fetch(path, {
    method: options.method ?? 'GET',
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error ?? 'Request failed');
  }
  return data as T;
}

