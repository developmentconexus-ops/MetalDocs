import KpiCard from "./KpiCard";
import styles from "./MembershipKpiStrip.module.css";

interface MembershipKpiStripProps {
  totalActive: number | null;
  distinctAreas: number | null;
  distinctUsers: number | null;
  isLoading: boolean;
}

export default function MembershipKpiStrip({
  totalActive,
  distinctAreas,
  distinctUsers,
  isLoading,
}: MembershipKpiStripProps) {
  return (
    <section
      className={styles.strip}
      aria-label="Indicadores de memberships"
      data-testid="membership-kpi-strip"
    >
      <KpiCard
        label="Memberships ativas"
        value={totalActive}
        tone="info"
        isLoading={isLoading}
        to="/admin/memberships"
        ctaLabel="Ver lista"
      />
      <KpiCard
        label="Áreas com acesso"
        value={distinctAreas}
        tone="neutral"
        isLoading={isLoading}
        to="/admin/memberships?sort=areaCode"
        ctaLabel="Ver por área"
      />
      <KpiCard
        label="Usuários com acesso"
        value={distinctUsers}
        tone="neutral"
        isLoading={isLoading}
        to="/admin/people"
        ctaLabel="Ver pessoas"
      />
    </section>
  );
}
