import { useParams } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { DocumentHero } from '../components/DocumentHero';
import { DocRefCard } from '../components/distribution/DocRefCard';
import { KPIStrip } from '../components/distribution/KPIStrip';
import { DonutCard } from '../components/distribution/DonutCard';
import { DistributionFacts } from '../components/distribution/DistributionFacts';
import { CoverageByArea } from '../components/distribution/CoverageByArea';
import { TimelineCard } from '../components/distribution/TimelineCard';
import { RecipientsCard } from '../components/distribution/RecipientsCard';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
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

function IllustrativeBlock({ children }: { children: React.ReactNode }) {
  return (
    <div className={styles.illustrative} aria-hidden="true">
      <div className={styles.illustrativeWatermark}>Dados ilustrativos · Em breve</div>
      <div className={styles.illustrativeBody}>{children}</div>
    </div>
  );
}

export function DocumentDistributionPage() {
  const { documentId: rawDocumentId } = useParams<{ documentId: string }>();
  const documentId = rawDocumentId ?? '';
  const docQuery = useDocumentDetailQuery(documentId);

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
  const code = doc.Code ?? EM_DASH;
  const docName = doc.Name;
  const versionLabel = formatRevisionCode(doc.RevisionNumber);

  return (
    <div className={styles.page}>
      <DocumentHero
        breadcrumbItems={[
          { label: 'Biblioteca', href: '/documents' },
          { label: EM_DASH },
          { label: code, href: `/documents/${documentId}` },
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
            <span className={styles.soonBadge}>Em breve</span>
          </>
        }
        title="Distribuição & cobertura de leitura"
        subtitle={<span>{docName ?? code}</span>}
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

      <main className={styles.main}>
        <div className={styles.banner} role="note">
          <Icon name="users" size={18} className={styles.bannerIcon} />
          <div className={styles.bannerBody}>
            <strong>Distribuição & cobertura de leitura — em breve.</strong>{' '}
            O rastreamento de leitura e o fanout ainda não estão disponíveis no
            backend. O layout abaixo é a previsão visual da tela; todos os números,
            áreas e pessoas exibidos são <em>ilustrativos</em> e não refletem dados
            reais de <strong>{docName ?? code}</strong>.
          </div>
        </div>

        <IllustrativeBlock>
          <KPIStrip />
        </IllustrativeBlock>

        <section className={styles.section}>
          <SectionHeader kicker="01 · Status" title="Cobertura geral e detalhes da distribuição" />
          <IllustrativeBlock>
            <div className={styles.twoCol}>
              <DonutCard />
              <DistributionFacts />
            </div>
          </IllustrativeBlock>
        </section>

        <section className={styles.section}>
          <SectionHeader kicker="02 · Por área" title="Onde está a pendência" />
          <IllustrativeBlock>
            <CoverageByArea />
          </IllustrativeBlock>
        </section>

        <section className={styles.section}>
          <SectionHeader
            kicker="03 · Linha do tempo"
            title="Curva de adoção desde a publicação"
          />
          <IllustrativeBlock>
            <TimelineCard />
          </IllustrativeBlock>
        </section>

        <section className={styles.section}>
          <SectionHeader kicker="04 · Destinatários" title="Lista detalhada" />
          <IllustrativeBlock>
            <RecipientsCard />
          </IllustrativeBlock>
        </section>
      </main>
    </div>
  );
}
