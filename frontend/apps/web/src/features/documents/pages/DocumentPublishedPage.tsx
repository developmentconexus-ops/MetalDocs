import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { CodeChip } from '../../../components/ui/CodeChip';
import styles from './DocumentPublishedPage.module.css';

export function DocumentPublishedPage() {
  // Phase 3c wires: const { documentId } = useParams();
  // For now, placeholder data
  return (
    <div className={styles.root}>
      {/* ObsoleteBanner positioned overlay — shown when status === 'obsolete' */}
      <div className={styles.obsoleteBanner}>
        <div className={styles.obsoleteStamp}>OBSOLETO</div>
      </div>

      {/* Hero */}
      <header className={styles.hero}>
        <div className={styles.heroBg} />

        {/* Breadcrumb */}
        <nav className={styles.breadcrumb}>
          <a className={styles.breadcrumbLink}>Biblioteca</a>
          <Icon name="chevron" size={10} className={styles.breadcrumbSep} />
          <a className={styles.breadcrumbLink}>SSMA</a>
          <Icon name="chevron" size={10} className={styles.breadcrumbSep} />
          <a className={styles.breadcrumbLink}>Procedimentos</a>
          <Icon name="chevron" size={10} className={styles.breadcrumbSep} />
          <span className={styles.breadcrumbCurrent}>PR-EHS-014</span>
        </nav>

        {/* Hero grid: DocCardMini + content */}
        <div className={styles.heroGrid}>
          <div className={styles.heroCardWrap}>
            <div className={styles.docCard}>
              <div className={styles.docCardHeader}>SSMA</div>
              <div className={styles.docCardBody}>
                <div className={styles.docCardCode}>PR-EHS-014</div>
                <div className={styles.docCardType}>Procedimento</div>
                <div className={styles.docCardSpacer} />
                <div className={styles.docCardDivider} />
                <div className={styles.docCardFooter}>
                  <span className={styles.docCardVersion}>v3.2</span>
                  <span className={styles.docCardDot} />
                </div>
              </div>
            </div>
          </div>

          <div className={styles.heroContent}>
            <div className={styles.heroBadges}>
              <CodeChip className={styles.codeChip}>PR-EHS-014</CodeChip>
              <span className={styles.vigeenteBadge}>
                <span className={styles.vigeenteDot} />
                v3.2 · vigente
              </span>
              <span className={styles.typeLabel}>Procedimento</span>
            </div>

            <h1 className={styles.heroTitle}>Procedimento de Bloqueio e Etiquetagem (LOTO)</h1>

            <div className={styles.heroActions}>
              <button className="btn btn-primary btn-lg" type="button">
                <Icon name="eye" size={15} />
                Visualizar documento
              </button>
              <button className="btn" type="button">
                <Icon name="arrow" size={13} />
                Iniciar revisão
              </button>
              <button className="btn btn-ghost" type="button">
                <Icon name="link" size={13} />
                Copiar link
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Content area */}
      <div className={styles.content}>

        {/* KPI strip — only Versão atual */}
        <div className={styles.kpiStrip}>
          <div className={styles.kpiCell}>
            <div className={styles.kpiLabel}>Versão atual</div>
            <div className={styles.kpiValue}>v3.2</div>
            <div className={styles.kpiHint}>desde 12 mar</div>
          </div>
        </div>

        {/* Section: Sobre */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>01 · Sobre</div>
              <h2 className={styles.sectionTitle}>Identificação e responsabilidade</h2>
            </div>
          </div>
          <div className={styles.aboutLayout}>
            {/* AboutCard */}
            <div className={styles.aboutCard}>
              {/* Owner banner */}
              <div className={styles.ownerBanner}>
                <Avatar name="Carolina Mendes" size="sm" />
                <div className={styles.ownerInfo}>
                  <div className={styles.ownerName}>Carolina Mendes</div>
                  <div className={styles.ownerMeta}>publicou em 12 mar 2026 · 14:32</div>
                </div>
              </div>
              {/* Facts grid */}
              <div className={styles.factsGrid}>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Tipo</div>
                    <div className={styles.factValue}>Procedimento operacional</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="taxonomy" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Área</div>
                    <div className={styles.factValue}>Saúde, Segurança &amp; Meio Ambiente</div>
                  </div>
                </div>
              </div>
            </div>
            {/* Deferred: CoverageCard + AuditCard */}
          </div>
        </section>

        {/* Section: Cadeia de aprovação */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>02 · Cadeia de aprovação</div>
              <h2 className={styles.sectionTitle}>Sign-offs desta versão</h2>
            </div>
          </div>
          {/* SignoffPipeline */}
          <div className={styles.signoffCard}>
            <div className={styles.signoffGrid}>
              {/* Connector line */}
              <div className={styles.signoffConnector} />
              {/* Stage items (3 placeholder stages) */}
              {['Revisão técnica', 'Aprovação gerencial', 'Anuência de diretoria'].map((stage) => (
                <div key={stage} className={styles.signoffStage}>
                  <div className={styles.signoffPin}>
                    <Icon name="check" size={16} />
                  </div>
                  <div className={styles.signoffStageName}>{stage}</div>
                  <div className={styles.signoffActor}>
                    <Avatar name="Renata Souza" size="sm" />
                    <span className={styles.signoffActorName}>Renata Souza</span>
                  </div>
                  <div className={styles.signoffActorRole}>Coord. SSMA</div>
                  <div className={styles.signoffWhen}>11 mar · 18:04</div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Deferred sections — not rendered */}
        {/* <VersionTimeline /> — deferred: no revision list endpoint */}
        {/* <RelatedGrid /> — deferred: no relationship model */}
        {/* <CommentsCard /> — deferred: needs architecture brainstorm */}

      </div>

      {/* ObsoleteBanner overlay — deferred condition wiring to Phase 3c */}
    </div>
  );
}
