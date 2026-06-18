import { useEffect, useState } from 'react';
import { getSystemStatus } from '../api';
import { MetricCard } from '../components/MetricCard';
import { ProgressRing } from '../components/ProgressRing';
import { SectionHeader } from '../components/SectionHeader';
import type { SystemSnapshot } from '../types';
import { formatBytes, formatDuration, formatMbps, formatPct } from '../utils/format';

type Props = {
  token: string;
};

export function SystemStatusPage({ token }: Props) {
  const [snapshot, setSnapshot] = useState<SystemSnapshot | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let active = true;
    let timer: number | undefined;

    async function load() {
      try {
        const data = await getSystemStatus(token);
        if (active) {
          setSnapshot(data);
          setError('');
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : '系统状态加载失败');
        }
      } finally {
        if (active) {
          timer = window.setTimeout(load, 2000);
        }
      }
    }

    void load();
    return () => {
      active = false;
      if (timer) window.clearTimeout(timer);
    };
  }, [token]);

  return (
    <div className="pageStack">
      <PageTitle title="系统状态" subtitle="实时资源、流量与连接状态" />
      {error && <div className="alert">{error}</div>}

      <section className="whiteStatusGrid">
        <article className="whiteStatusCard">
          <div>
            <span>负载</span>
            <strong>{snapshot?.health.label ?? 'N/A'}</strong>
          </div>
          <ProgressRing value={snapshot?.health.loadPct ?? 0} label="" caption="" />
        </article>
        <article className="whiteStatusCard">
          <div>
            <span>CPU</span>
            <strong>{snapshot ? `${snapshot.cpu.cores}核心` : 'N/A'}</strong>
          </div>
          <ProgressRing value={snapshot?.cpu.usagePct ?? 0} label="" caption="" />
        </article>
        <article className="whiteStatusCard">
          <div>
            <span>内存</span>
            <strong>
              {snapshot ? `${formatBytes(snapshot.memory.usedBytes)} / ${formatBytes(snapshot.memory.totalBytes)}` : 'N/A'}
            </strong>
          </div>
          <ProgressRing value={snapshot?.memory.usagePct ?? 0} label="" caption="" />
        </article>
        <article className="whiteStatusCard">
          <div>
            <span>{snapshot?.disk.mount ?? '/'}</span>
            <strong>{snapshot ? `${formatBytes(snapshot.disk.usedBytes)} / ${formatBytes(snapshot.disk.totalBytes)}` : 'N/A'}</strong>
          </div>
          <ProgressRing value={snapshot?.disk.usagePct ?? 0} label="" caption="" />
        </article>
      </section>

      <section className="darkGrid">
        <MetricCard
          title="系统正常运行时间"
          metrics={[
            { label: '面板', value: formatDuration(snapshot?.uptime.appSeconds ?? 0) },
            { label: 'OS', value: formatDuration(snapshot?.uptime.systemSeconds ?? 0) },
          ]}
        />
        <MetricCard
          title="使用情况"
          metrics={[
            { label: 'RAM', value: snapshot ? `${formatBytes(snapshot.memory.usedBytes)} / ${formatBytes(snapshot.memory.totalBytes)}` : 'N/A' },
            { label: '线程', value: snapshot ? String(snapshot.system.threads) : 'N/A' },
          ]}
        />
        <MetricCard
          title="整体速度"
          metrics={[
            { label: '上传', value: formatMbps(snapshot?.network.transmitMbps ?? 0) },
            { label: '下载', value: formatMbps(snapshot?.network.receiveMbps ?? 0) },
          ]}
        />
        <MetricCard
          title="总数据"
          metrics={[
            { label: '已发送', value: formatBytes(snapshot?.network.transmittedBytes ?? 0) },
            { label: '已接收', value: formatBytes(snapshot?.network.receivedBytes ?? 0) },
          ]}
        />
        <MetricCard
          title="IP地址"
          metrics={[
            { label: 'IPv4', value: snapshot?.network.ipv4 || 'N/A' },
            { label: '系统', value: snapshot?.system.os || 'Linux' },
          ]}
        />
        <MetricCard
          title="连接数"
          metrics={[
            { label: 'TCP', value: snapshot ? String(snapshot.connections.tcp) : 'N/A' },
            { label: 'UDP', value: snapshot ? String(snapshot.connections.udp) : 'N/A' },
          ]}
        />
      </section>

      <section className="panel">
        <SectionHeader title="状态" subtitle="资源使用率" />
        <div className="ringGrid">
          <ProgressRing value={snapshot?.cpu.usagePct ?? 0} label="CPU" caption={snapshot ? `${snapshot.cpu.cores} 核心` : 'N/A'} />
          <ProgressRing value={snapshot?.memory.usagePct ?? 0} label="内存" caption={snapshot ? formatBytes(snapshot.memory.totalBytes) : 'N/A'} />
          <ProgressRing value={snapshot?.disk.usagePct ?? 0} label={snapshot?.disk.mount ?? '/'} caption={snapshot ? formatBytes(snapshot.disk.totalBytes) : 'N/A'} />
          <ProgressRing value={snapshot?.health.loadPct ?? 0} label="负载" caption={snapshot ? formatPct(snapshot.health.loadPct) : 'N/A'} />
        </div>
      </section>
    </div>
  );
}

function PageTitle({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <header className="pageHeader">
      <div>
        <p className="eyebrow">VPS Inspector</p>
        <h1>{title}</h1>
        <span>{subtitle}</span>
      </div>
    </header>
  );
}

