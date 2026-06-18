type Metric = {
  label: string;
  value: string;
};

type Props = {
  title: string;
  metrics: Metric[];
};

export function MetricCard({ title, metrics }: Props) {
  return (
    <article className="metricCard">
      <h2>{title}</h2>
      <div className="metricGrid">
        {metrics.map((metric) => (
          <div key={metric.label}>
            <span>{metric.label}</span>
            <strong>{metric.value}</strong>
          </div>
        ))}
      </div>
    </article>
  );
}

