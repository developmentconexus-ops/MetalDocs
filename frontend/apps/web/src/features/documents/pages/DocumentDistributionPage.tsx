import { Icon } from '../../../components/ui/Icon';
import { DocumentHero } from '../components/DocumentHero';
import { DocRefCard } from '../components/distribution/DocRefCard';
import { KPIStrip } from '../components/distribution/KPIStrip';
import { DonutCard } from '../components/distribution/DonutCard';
import { DistributionFacts } from '../components/distribution/DistributionFacts';
import { CoverageByArea } from '../components/distribution/CoverageByArea';
import { TimelineCard } from '../components/distribution/TimelineCard';
import { RecipientsCard } from '../components/distribution/RecipientsCard';
import styles from './DocumentDistributionPage.module.css';

function SectionHeader({
  kicker,
  title,
  aside,
}: {
  kicker: string;
  title: string;
  aside?: React.ReactNode;
}) {
  return (
    <div className={styles.sectionHeader}>
      <div>
        <div className={styles.sectionKicker}>{kicker}</div>
        <h2 className={styles.sectionTitle}>{title}</h2>
      </div>
      {aside && <div className={styles.sectionAside}>{aside}</div>}
    </div>
  );
}

export function DocumentDistributionPage() {
  return (
    <div className={styles.page}>
      <DocumentHero
        breadcrumbItems={[
          { label: 'Biblioteca', href: '#' },
          { label: 'SSMA', href: '#' },
          { label: 'PR-EHS-014', href: '#' },
          { label: 'Distribuição' },
        ]}
        docCard={<DocRefCard />}
        badges={
          <>
            <span className={styles.codeBadge}>PR-EHS-014 · v3.2</span>
            <span className={styles.deadlineBadge}>
              <span className={styles.deadlineDot} />
              Vence em 8 dias
            </span>
            <span className={styles.sectionBadge}>§03.05 · Fanout</span>
          </>
        }
        title="Distribuição & cobertura de leitura"
        subtitle="Procedimento de Bloqueio e Etiquetagem (LOTO) · publicado em 12 mar 2026 · 14:32. Reconhecimento obrigatório por assinatura para todas as 248 pessoas alcançadas."
        actions={
          <>
            <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
              <Icon name="mail" size={15} />
              Lembrete em massa
            </button>
            <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
              <Icon name="download" size={13} />
              Exportar relatório
            </button>
            <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
              <Icon name="users" size={13} />
              Adicionar destinatários
            </button>
            <button type="button" aria-disabled="true" title="Em breve" className={styles.ctaDisabled}>
              <Icon name="cog" size={13} />
              Política de fanout
            </button>
          </>
        }
      />

      {/* Main content */}
      <main className={styles.main}>
        <KPIStrip />

        {/* §01 — Cobertura geral + detalhes */}
        <section className={styles.section}>
          <SectionHeader
            kicker="01 · Status"
            title="Cobertura geral e detalhes da distribuição"
          />
          <div className={styles.twoCol}>
            <DonutCard />
            <DistributionFacts />
          </div>
        </section>

        {/* §02 — Por área */}
        <section className={styles.section}>
          <SectionHeader
            kicker="02 · Por área"
            title="Onde está a pendência"
            aside={
              <a href="#" className={styles.sectionAside}>
                Lembrar todas as áreas críticas →
              </a>
            }
          />
          <CoverageByArea />
        </section>

        {/* §03 — Linha do tempo */}
        <section className={styles.section}>
          <SectionHeader
            kicker="03 · Linha do tempo"
            title="Curva de adoção desde a publicação"
            aside={<span>últimos 8 dias</span>}
          />
          <TimelineCard />
        </section>

        {/* §04 — Destinatários */}
        <section className={styles.section}>
          <SectionHeader
            kicker="04 · Destinatários"
            title="Lista detalhada"
            aside={<span>selecione para ações em massa</span>}
          />
          <RecipientsCard />
        </section>
      </main>
    </div>
  );
}
