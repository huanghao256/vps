import type { CSSProperties } from 'react';
import { formatPct } from '../utils/format';

type Props = {
  value: number;
  label: string;
  caption: string;
  tone?: 'green' | 'blue' | 'amber' | 'red';
};

export function ProgressRing({ value, label, caption, tone = 'green' }: Props) {
  const safeValue = Number.isFinite(value) ? Math.max(0, Math.min(value, 100)) : 0;
  return (
    <div className="ringItem">
      <div className={`ring ${tone}`} style={{ '--value': `${safeValue}%` } as CSSProperties}>
        <span>{formatPct(safeValue)}</span>
      </div>
      <strong>{label}</strong>
      <small>{caption}</small>
    </div>
  );
}

