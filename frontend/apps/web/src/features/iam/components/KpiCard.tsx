interface KpiCardProps {
  label?: string;
  value?: string | number;
}

export default function KpiCard(_props: KpiCardProps) {
  return <div data-testid="kpi-card" />;
}
