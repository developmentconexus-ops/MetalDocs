import { useParams } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { DocumentHero } from '../components/DocumentHero';
import { DocRefCard } from '../components/distribution/DocRefCard';
import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { formatRevisionCode } from '../lib/documentDetailMeta';
import styles from './DocumentDistributionPage.module.css';

const EM_DASH = '—';

const PLANNED_CAPABILITIES = [
  { title: 'Cobertura de leitura', desc: 'Percentual de destinatários que leram e reconheceram o documento.' },
  { title: 'Cobertura por área', desc: 'Onde está a pendência, ranqueado pela menor adesão.' },
  { title: 'Curva de adoção', desc: 'Linha do tempo de leituras e reconhecimentos desde a publicação.' },
  { title: 'Lista de destinatários', desc: 'Quem leu, reconheceu ou está em atraso, com ações de lembrete.' },
];

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
        docCard={<DocRefCard areaLabel={EM_DASH} code={code} typeLabel={EM_DASH} versionLabel={versionLabel} />}
        badges={
          <>
            <span className={styles.codeBadge}>{code} · {versionLabel}</span>
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
        <section className={styles.empty} aria-labelledby="distribution-soon-title">
          <Icon name="users" size={28} className={styles.emptyIcon} />
          <h2 id="distribution-soon-title" className={styles.emptyTitle}>
            Distribuição e cobertura de leitura — em breve
          </h2>
          <p className={styles.emptyText}>
            O rastreamento de leitura e o fanout de distribuição ainda não estão
            disponíveis. Quando o backend de distribuição entrar no ar, esta página
            exibirá métricas reais para <strong>{docName ?? code}</strong>.
          </p>
          <ul className={styles.emptyList}>
            {PLANNED_CAPABILITIES.map((cap) => (
              <li key={cap.title} className={styles.emptyItem}>
                <span className={styles.emptyItemTitle}>{cap.title}</span>
                <span className={styles.emptyItemDesc}>{cap.desc}</span>
              </li>
            ))}
          </ul>
        </section>
      </main>
    </div>
  );
}
