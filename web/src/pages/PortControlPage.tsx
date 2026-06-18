import { FormEvent, useEffect, useState } from 'react';
import { addFirewallRule, deleteFirewallRule, disableFirewall, enableFirewall, getFirewall } from '../api';
import { SectionHeader } from '../components/SectionHeader';
import type { FirewallSnapshot, PortRuleInput } from '../types';

type Props = {
  token: string;
};

export function PortControlPage({ token }: Props) {
  const [snapshot, setSnapshot] = useState<FirewallSnapshot | null>(null);
  const [port, setPort] = useState('443');
  const [protocol, setProtocol] = useState<'tcp' | 'udp'>('tcp');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    void refresh();
  }, [token]);

  async function refresh() {
    try {
      setSnapshot(await getFirewall(token));
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : '防火墙状态加载失败');
    }
  }

  async function mutate(action: () => Promise<FirewallSnapshot>) {
    setBusy(true);
    setError('');
    try {
      setSnapshot(await action());
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    } finally {
      setBusy(false);
    }
  }

  function ruleInput(): PortRuleInput {
    return { port: Number(port), protocol };
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    await mutate(() => addFirewallRule(token, ruleInput()));
  }

  return (
    <div className="pageStack">
      <header className="pageHeader">
        <div>
          <p className="eyebrow">Firewall</p>
          <h1>端口控制</h1>
          <span>{snapshot?.backend ? `当前后端：${snapshot.backend}` : '检测防火墙后端'}</span>
        </div>
        <div className="actions">
          <button className="secondaryButton" onClick={() => void refresh()}>
            刷新
          </button>
          <button
            className={snapshot?.enabled ? 'dangerButton' : 'primaryButton'}
            disabled={busy || !snapshot?.available}
            onClick={() => mutate(() => (snapshot?.enabled ? disableFirewall(token) : enableFirewall(token)))}
          >
            {snapshot?.enabled ? '关闭防火墙' : '开启防火墙'}
          </button>
        </div>
      </header>

      {error && <div className="alert">{error}</div>}

      <section className="firewallHero">
        <div>
          <span>防火墙状态</span>
          <strong>{snapshot?.enabled ? '已开启' : '未开启'}</strong>
          <p>{snapshot?.message || 'N/A'}</p>
        </div>
        <span className={`firewallBadge ${snapshot?.enabled ? 'enabled' : 'disabled'}`}>
          {snapshot?.available ? snapshot.backend : '不可用'}
        </span>
      </section>

      <section className="panel">
        <SectionHeader title="添加端口规则" subtitle="支持 TCP / UDP" />
        <form className="portForm" onSubmit={submit}>
          <label>
            <span>端口</span>
            <input value={port} onChange={(event) => setPort(event.target.value)} inputMode="numeric" />
          </label>
          <label>
            <span>协议</span>
            <select value={protocol} onChange={(event) => setProtocol(event.target.value as 'tcp' | 'udp')}>
              <option value="tcp">TCP</option>
              <option value="udp">UDP</option>
            </select>
          </label>
          <button className="primaryButton" disabled={busy || !snapshot?.available}>
            添加规则
          </button>
        </form>
      </section>

      <section className="panel">
        <SectionHeader title="端口规则" subtitle={`${snapshot?.rules.length ?? 0} 条`} />
        <div className="ruleTable">
          <div className="ruleHead">
            <span>端口</span>
            <span>协议</span>
            <span>动作</span>
            <span>来源</span>
            <span />
          </div>
          {snapshot?.rules.map((rule) => (
            <div className="ruleRow" key={`${rule.port}-${rule.protocol}-${rule.action}-${rule.source}`}>
              <strong>{rule.port}</strong>
              <span>{rule.protocol.toUpperCase()}</span>
              <span>{rule.action}</span>
              <span>{rule.source}</span>
              <button
                className="dangerTextButton"
                disabled={busy}
                onClick={() => mutate(() => deleteFirewallRule(token, { port: rule.port, protocol: rule.protocol as 'tcp' | 'udp' }))}
              >
                删除
              </button>
            </div>
          ))}
          {snapshot?.rules.length === 0 && <div className="emptyState">暂无端口规则</div>}
        </div>
      </section>
    </div>
  );
}

