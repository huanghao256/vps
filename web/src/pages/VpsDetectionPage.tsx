import { useEffect, useMemo, useState } from 'react';
import { createRun, listChecks, listRuns } from '../api';
import { ProgressRing } from '../components/ProgressRing';
import { SectionHeader } from '../components/SectionHeader';
import type { CheckInfo, CheckResult, Run } from '../types';
import { asNumber, asRecord, asString, formatDate, formatMbps, formatMs } from '../utils/format';

type Props = {
  token: string;
};

type Dimension = {
  id: string;
  title: string;
  score: number;
  status: string;
  summary: string;
  metrics: Array<{ label: string; value: string }>;
  steps: Array<{ label: string; state: 'done' | 'warn' | 'fail' | 'idle' }>;
};

export function VpsDetectionPage({ token }: Props) {
  const [checks, setChecks] = useState<CheckInfo[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [runs, setRuns] = useState<Run[]>([]);
  const [activeRun, setActiveRun] = useState<Run | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    void refresh();
  }, [token]);

  async function refresh() {
    try {
      const [checkList, runList] = await Promise.all([listChecks(token), listRuns(token)]);
      setChecks(checkList);
      setSelected(new Set(checkList.map((check) => check.id)));
      setRuns(runList);
      setActiveRun(runList[0] ?? null);
      setError('');
    } catch (err) {
      setError(err instanceof Error ? err.message : '检测数据加载失败');
    }
  }

  async function runDetection() {
    setLoading(true);
    setError('');
    try {
      const run = await createRun(token, [...selected]);
      setActiveRun(run);
      setRuns((current) => [run, ...current.filter((item) => item.id !== run.id)].slice(0, 20));
    } catch (err) {
      setError(err instanceof Error ? err.message : '检测失败');
    } finally {
      setLoading(false);
    }
  }

  const resultMap = useMemo(() => {
    const map = new Map<string, CheckResult>();
    activeRun?.results.forEach((result) => map.set(result.checkId, result));
    return map;
  }, [activeRun]);
  const dimensions = buildDimensions(resultMap);

  return (
    <div className="pageStack">
      <header className="pageHeader">
        <div>
          <p className="eyebrow">VPS Quality</p>
          <h1>VPS检测</h1>
          <span>{activeRun ? `最近检测：${formatDate(activeRun.endedAt ?? activeRun.startedAt)}` : '线路、延迟、带宽、稳定性、IP风控风险'}</span>
        </div>
        <div className="actions">
          <button className="secondaryButton" onClick={() => void refresh()}>
            刷新
          </button>
          <button className="primaryButton" disabled={loading || selected.size === 0} onClick={runDetection}>
            {loading ? '检测中...' : '开始检测'}
          </button>
        </div>
      </header>

      {error && <div className="alert">{error}</div>}

      <section className="scoreHero">
        <div>
          <span>综合评分</span>
          <strong>{activeRun?.score ?? '--'}</strong>
          <p>{activeRun ? scoreText(activeRun.score) : '等待检测'}</p>
        </div>
        <ProgressRing value={activeRun?.score ?? 0} label="总分" caption={`${activeRun?.results.length ?? 0} 项`} />
      </section>

      <section className="qualityGrid">
        {dimensions.map((dimension) => (
          <article className="qualityCard" key={dimension.id}>
            <div className="qualityHead">
              <div>
                <span>{dimension.title}</span>
                <strong>{dimension.score || '--'}</strong>
              </div>
              <span className={`statusDot ${dimension.status}`} />
            </div>
            <p>{dimension.summary}</p>
            <div className="metricRows">
              {dimension.metrics.map((metric) => (
                <div className="metricLine" key={metric.label}>
                  <span>{metric.label}</span>
                  <strong>{metric.value}</strong>
                </div>
              ))}
            </div>
            <div className="processTrack">
              {dimension.steps.map((step) => (
                <div className={`processStep ${step.state}`} key={step.label}>
                  <span />
                  <small>{step.label}</small>
                </div>
              ))}
            </div>
          </article>
        ))}
      </section>

      <section className="panel">
        <SectionHeader title="检测模块" subtitle="可按需组合检测项" />
        <div className="checkGrid">
          {checks.map((check) => (
            <label className="checkTile" key={check.id}>
              <input
                type="checkbox"
                checked={selected.has(check.id)}
                onChange={(event) => {
                  const next = new Set(selected);
                  if (event.target.checked) next.add(check.id);
                  else next.delete(check.id);
                  setSelected(next);
                }}
              />
              <span>
                <strong>{checkName(check)}</strong>
                <small>{checkDescription(check)}</small>
              </span>
            </label>
          ))}
        </div>
      </section>

      <section className="panel">
        <SectionHeader title="检测历史" subtitle={`${runs.length} 条记录`} />
        <div className="historyRows">
          {runs.map((run) => (
            <button className="historyRow" key={run.id} onClick={() => setActiveRun(run)}>
              <span>{run.id}</span>
              <strong>{run.score}</strong>
              <small>{formatDate(run.endedAt ?? run.startedAt)}</small>
            </button>
          ))}
        </div>
      </section>
    </div>
  );
}

function buildDimensions(results: Map<string, CheckResult>): Dimension[] {
  return [
    lineDimension(results.get('network.route_profile')),
    latencyDimension(results.get('network.tcp_latency')),
    bandwidthDimension(results.get('network.bandwidth')),
    stabilityDimension(results.get('network.stability')),
    riskDimension(results.get('risk.reputation')),
  ];
}

function lineDimension(result?: CheckResult): Dimension {
  const data = asRecord(result?.details);
  return baseDimension('line', '线路', result, [
    { label: '线路类型', value: asString(data.lineType) || 'N/A' },
    { label: '识别置信度', value: asString(data.confidence) || 'N/A' },
    { label: '平均延迟', value: formatMs(data.averageLatencyMs) },
  ]);
}

function latencyDimension(result?: CheckResult): Dimension {
  const data = asRecord(result?.details);
  return baseDimension('latency', '延迟', result, [
    { label: '中位延迟', value: formatMs(data.medianLatencyMs) },
    { label: '成功目标', value: `${asNumber(data.successes)}/3` },
  ]);
}

function bandwidthDimension(result?: CheckResult): Dimension {
  const data = asRecord(result?.details);
  return baseDimension('bandwidth', '带宽', result, [
    { label: '下载', value: formatMbps(asNumber(data.downloadMbps)) },
    { label: '上传', value: formatMbps(asNumber(data.uploadMbps)) },
  ]);
}

function stabilityDimension(result?: CheckResult): Dimension {
  const data = asRecord(result?.details);
  return baseDimension('stability', '稳定性', result, [
    { label: '丢包率', value: `${asNumber(data.lossPct).toFixed(1)}%` },
    { label: '网络抖动', value: formatMs(data.jitterMs) },
  ]);
}

function riskDimension(result?: CheckResult): Dimension {
  const data = asRecord(result?.details);
  return baseDimension('risk', 'IP风控风险', result, [
    { label: '出口IP', value: asString(data.ip) || 'N/A' },
    { label: '地区', value: asString(data.country) || 'N/A' },
    { label: 'Cloudflare', value: asString(data.cloudflareColo) || 'N/A' },
  ]);
}

function baseDimension(id: string, title: string, result: CheckResult | undefined, metrics: Dimension['metrics']): Dimension {
  return {
    id,
    title,
    score: result?.score ?? 0,
    status: result?.status ?? 'idle',
    summary: result?.summary ?? '尚未检测',
    metrics,
    steps: [
      { label: '初始化', state: result ? 'done' : 'idle' },
      { label: '探测', state: result ? stateFromStatus(result.status) : 'idle' },
      { label: '评分', state: result ? stateFromStatus(result.status) : 'idle' },
    ],
  };
}

function stateFromStatus(status: CheckResult['status']) {
  if (status === 'pass') return 'done';
  if (status === 'warn') return 'warn';
  if (status === 'fail') return 'fail';
  return 'idle';
}

function scoreText(score: number) {
  if (score >= 85) return '质量优秀，适合作为长期代理出口';
  if (score >= 70) return '质量良好，适合日常代理访问';
  if (score >= 50) return '可用但存在短板';
  return '质量偏弱，建议更换线路或机房';
}

function checkName(check: CheckInfo) {
  const names: Record<string, string> = {
    'system.info': '系统信息',
    'network.overview': '网络总览',
    'network.route_profile': '线路识别',
    'network.tcp_latency': '延迟检测',
    'network.bandwidth': '带宽检测',
    'network.stability': '稳定性检测',
    'risk.reputation': 'IP风控风险',
  };
  return names[check.id] ?? check.name;
}

function checkDescription(check: CheckInfo) {
  const descriptions: Record<string, string> = {
    'system.info': '采集基础系统信息',
    'network.overview': '采集流量与连接数',
    'network.route_profile': '识别三网线路画像',
    'network.tcp_latency': '检测 TCP 连接延迟',
    'network.bandwidth': '检测上传与下载速度',
    'network.stability': '检测丢包与抖动',
    'risk.reputation': '检测出口 IP 风控信号',
  };
  return descriptions[check.id] ?? check.description;
}

