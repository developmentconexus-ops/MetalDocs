import { useParams } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { ArtifactHero } from '../../shared/controlled-artifact/ArtifactHero';
import { DocRefCard } from '../components/distribution/DocRefCard';
import { KPIStrip } from '../components/distribution/KPIStrip';
import { DonutCard } from '../components/distribution/DonutCard';
import { DistributionFacts } from '../components/distribution/DistributionFacts';
import { CoverageByArea } from '../components/distribution/CoverageByArea';
import { TimelineCard } from '../components/distribution/TimelineCard';
import { RecipientsCard } from '../components/distribution/RecipientsCard';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useDistributionSummaryQuery } from '../queries/useDistributionSummaryQuery';
import { useDistributionRecipientsQuery } from '../queries/useDistributionRecipientsQuery';
import { useDistributionCoverageQuery } from '../queries/useDistributionCoverageQuery';
import { formatRevisionCode } from '../lib/documentDetailMeta';
import styles from './DocumentDistributionPage.module.css';

const EM_DASH = '—';

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
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const documentId = rawDocumentId ?? '';

  const docQuery = useDocumentDetailQuery(documentId);
  const summaryQuery = useDistributionSummaryQuery(documentId);
  const recipientsQuery = useDistributionRecipientsQuery(documentId);
  const coverageQuery = useDistributionCoverageQuery(documentId);

  if (docQuery.isLoading) {
    return (
      <div className={styles.stateLoading} role="status" aria-live="polite">
        <Icon name="docs" size={24} className={styles.stateIcon} />
        <span>Carregando documento…</span>
      </div>
    );
  }

  if (docQuery.isError || !docQuery.data) {
    return (
      <div className={styles.stateError} role="alert">
        <Icon name="x" size={20} className={styles.stateIcon} />
        <span>Documento não encontrado ou sem permissão de acesso.</span>
        <button className="btn btn-sm" type="button" onClick={() => docQuery.refetch()}>
          Tentar novamente
        </button>
      </div>
    );
  }

  const doc = docQuery.data;
  const code = doc.code ?? EM_DASH;
  const docName = doc.name;
  const versionLabel = formatRevisionCode(doc.revision_number);

  // Show EM_DASH when summary errored so we never display a fabricated 0
  const totalTargets = summaryQuery.isError
    ? EM_DASH
    : (summaryQuery.data?.total_targets ?? EM_DASH);
  const summaryLoading = summaryQuery.isLoading;

  // Flatten all pages of recipients into a single list
  const allRecipients = recipientsQuery.data?.pages.flatMap((p) => p.items) ?? [];
  const hasMoreRecipients = recipientsQuery.hasNextPage ?? false;
  const recipientsLoading = recipientsQuery.isLoading || recipientsQuery.isFetchingNextPage;

  const coverageRows = coverageQuery.data ?? [];
  const coverageLoading = coverageQuery.isLoading;

  return (
    <div className={styles.page}>
      <ArtifactHero
        breadcrumb={[
          { label: 'Biblioteca', href: '/documents' },
          { label: EM_DASH },
          { label: code, href: `/documents/${documentId}/details` },
          { label: 'Distribuição' },
        ]}
        docCard={
          <DocRefCard
            areaLabel={EM_DASH}
            code={code}
            typeLabel={EM_DASH}
            versionLabel={versionLabel}
          />
        }
        badges={
          <>
            <span className={styles.codeBadge}>
              {code} · {versionLabel}
            </span>
          </>
        }
        title="Distribuição & cobertura de leitura"
        subtitle={<span>{docName ?? code}</span>}
        actions={
          <>
            <button
              type="button"
              aria-disabled="true"
              title="Ação disponível na missão de distribuição (wiki/backlog/document-distribution-mission.md)"
              className={styles.ctaDisabled}
            >
              <Icon name="mail" size={15} />
              Lembrete em massa
            </button>
            <button
              type="button"
              aria-disabled="true"
              title="Ação disponível na missão de distribuição (wiki/backlog/document-distribution-mission.md)"
              className={styles.ctaDisabled}
            >
              <Icon name="download" size={13} />
              Exportar relatório
            </button>
            <button
              type="button"
              aria-disabled="true"
              title="Ação disponível na missão de distribuição (wiki/backlog/document-distribution-mission.md)"
              className={styles.ctaDisabled}
            >
              <Icon name="users" size={13} />
              Adicionar destinatários
            </button>
            <button
              type="button"
              aria-disabled="true"
              title="Ação disponível na missão de distribuição (wiki/backlog/document-distribution-mission.md)"
              className={styles.ctaDisabled}
            >
              <Icon name="cog" size={13} />
              Política de fanout
            </button>
          </>
        }
      />

      <main className={styles.main}>
        <KPIStrip totalTargets={totalTargets} loading={summaryLoading} />

        <section className={styles.section}>
          <SectionHeader kicker="01 · Status" title="Cobertura geral e detalhes da distribuição" />
          <div className={styles.twoCol}>
            <DonutCard />
            <DistributionFacts />
          </div>
        </section>

        <section className={styles.section}>
          <SectionHeader kicker="02 · Por área" title="Destinatários por área" />
          <CoverageByArea
            rows={coverageRows}
            loading={coverageLoading}
            error={coverageQuery.isError}
          />
        </section>

        <section className={styles.section}>
          <SectionHeader
            kicker="03 · Linha do tempo"
            title="Curva de adoção desde a publicação"
          />
          <TimelineCard />
        </section>

        <section className={styles.section}>
          <SectionHeader kicker="04 · Destinatários" title="Lista detalhada" />
          <RecipientsCard
            items={allRecipients}
            hasMore={hasMoreRecipients}
            onLoadMore={() => void recipientsQuery.fetchNextPage()}
            loading={recipientsLoading}
            error={recipientsQuery.isError}
          />
        </section>
      </main>
    </div>
  );
}
