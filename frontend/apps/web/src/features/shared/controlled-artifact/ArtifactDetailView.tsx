import type React from "react";
import { Link } from "react-router-dom";
import { Icon } from "../../../components/ui/Icon";
import { Avatar } from "../../../components/ui/Avatar";
import { CodeChip } from "../../../components/ui/CodeChip";
import { formatShortDate, formatSignedAt } from "../../../lib/format/dates";
import { formatFileSize } from "../../../lib/format/fileSize";
import { formatRevisionCode } from "../../../lib/labels/revisionCode";
import { ArtifactHero } from "./ArtifactHero";
import { VersionTimeline, type VersionEntry } from "./VersionTimeline";
import type { ApprovalChainItem, ArtifactViewModel } from "./types";
import styles from "./ArtifactDetailView.module.css";

const EM_DASH = "—";

interface ArtifactDetailViewProps {
  model: ArtifactViewModel;
  /** Rendered into the ArtifactHero `actions` area (document workflow buttons, etc.). */
  heroActions?: React.ReactNode;
  /** Rendered as the coverage side-card area in the About section. */
  aside?: React.ReactNode;
  /** Dialogs/composers/banners rendered by the caller (revision composer, publish dialog, obsolete banner). */
  extras?: React.ReactNode;
}

function groupByStageIndex(chain: ApprovalChainItem[]): ApprovalChainItem[][] {
  const byStage = new Map<number, ApprovalChainItem[]>();
  for (const item of chain) {
    const bucket = byStage.get(item.stageIndex);
    if (bucket) {
      bucket.push(item);
    } else {
      byStage.set(item.stageIndex, [item]);
    }
  }
  return [...byStage.entries()]
    .sort(([a], [b]) => a - b)
    .map(([, items]) => items);
}

function toVersionEntries(model: ArtifactViewModel): VersionEntry[] {
  return model.lineage.map((item) => ({
    v: item.revisionLabel ?? `REV${String(item.revisionNumber ?? 0).padStart(2, "0")}`,
    when: formatShortDate(item.createdAt) || EM_DASH,
    // VersionEntry.author shows the lifecycle status — no per-revision author field exists on the model yet.
    author: item.status || EM_DASH,
    current: item.isCurrent,
    summary: item.title?.trim() || "Sem título governado registrado.",
  }));
}

/**
 * Purely presentational detail surface for a controlled artifact. Consumes an
 * `ArtifactViewModel` and optional ReactNode slots — no data fetching, no
 * `useParams`, no `kind` branching. Every kind-specific difference is either
 * normalized into the model by the adapter or injected via a slot by the caller.
 */
export function ArtifactDetailView({ model, heroActions, aside, extras }: ArtifactDetailViewProps) {
  const code = model.code ?? EM_DASH;
  const versionLabel = model.revisionLabel ?? formatRevisionCode(model.versionNumber);
  const profileLabel = model.meta.profileLabel ?? EM_DASH;
  const areaLabel = model.meta.areaLabel ?? EM_DASH;
  const fileSizeLabel = formatFileSize(model.meta.fileSizeBytes);
  const effectiveFromLabel = model.meta.effectiveFrom ? formatShortDate(model.meta.effectiveFrom) : EM_DASH;
  const nextReviewLabel = model.meta.nextReviewAt ? formatShortDate(model.meta.nextReviewAt) : EM_DASH;
  const subtitle = model.hero.subtitle;

  const approvalGroups = model.approvalChain ? groupByStageIndex(model.approvalChain) : [];
  const stageCount = approvalGroups.length;
  // Connector spans from first pin center to last pin center.
  const connectorSide = stageCount > 1 ? `${(100 / (2 * stageCount)).toFixed(2)}%` : "50%";

  const versionEntries = toVersionEntries(model);

  return (
    <div className={styles.root}>
      {extras}

      <ArtifactHero
        breadcrumb={model.hero.breadcrumb}
        docCard={
          <div className={styles.docCard}>
            <div className={styles.docCardHeader}>{areaLabel}</div>
            <div className={styles.docCardBody}>
              <div className={styles.docCardCode}>{code}</div>
              <div className={styles.docCardType}>{profileLabel}</div>
              <div className={styles.docCardSpacer} />
              <div className={styles.docCardDivider} />
              <div className={styles.docCardFooter}>
                <span className={styles.docCardVersion}>{versionLabel}</span>
                <span className={styles.docCardDot} />
              </div>
            </div>
          </div>
        }
        badges={
          <>
            {model.hero.badges.map((badge) => {
              if (badge.variant === "code") {
                return (
                  <CodeChip key={badge.key} className={styles.codeChip}>
                    {badge.label}
                  </CodeChip>
                );
              }
              if (badge.variant === "status") {
                return (
                  <span key={badge.key} className={styles.vigenteBadge}>
                    <span className={styles.vigenteDot} />
                    {badge.label}
                  </span>
                );
              }
              return (
                <span key={badge.key} className={styles.typeLabel}>
                  {badge.label}
                </span>
              );
            })}
          </>
        }
        title={model.title}
        subtitle={subtitle ? <span>{subtitle}</span> : null}
        actions={heroActions}
      />

      {/* Content area */}
      <div className={styles.content}>

        {/* KPI strip — model-driven. Adapters compose all cells (including kind-specific rules
            such as the document scheduled→published-head current-version override and the live
            coverage denominator). The shared view renders them in order with zero kind awareness. */}
        <div className={styles.kpiStrip}>
          {model.kpis.map((cell) =>
            cell.href ? (
              <Link key={cell.key} to={cell.href} className={`${styles.kpiCell} ${styles.kpiCellLink}`}>
                <div className={styles.kpiLabel}>{cell.label}</div>
                <div className={styles.kpiValue}>{cell.value}</div>
                {cell.hint ? <div className={styles.kpiHint}>{cell.hint}</div> : null}
              </Link>
            ) : (
              <div key={cell.key} className={styles.kpiCell}>
                <div className={styles.kpiLabel}>{cell.label}</div>
                <div className={styles.kpiValue}>{cell.value}</div>
                {cell.hint ? <div className={styles.kpiHint}>{cell.hint}</div> : null}
              </div>
            ),
          )}
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
            <div className={styles.aboutCard}>
              <div className={styles.ownerBanner}>
                <Avatar name={model.meta.ownerName ?? model.title} size="sm" />
                <div className={styles.ownerInfo}>
                  <div className={styles.ownerName}>{model.meta.ownerName ?? model.title}</div>
                  <div className={styles.ownerMeta}>{model.meta.ownerDescriptor ?? subtitle ?? code}</div>
                </div>
              </div>
              <div className={styles.factsGrid}>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Tipo</div>
                    <div className={styles.factValue}>{profileLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="taxonomy" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Área</div>
                    <div className={styles.factValue}>{areaLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="calendar" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Vigente desde</div>
                    <div className={styles.factValue}>{effectiveFromLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="calendar" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Próxima revisão</div>
                    <div className={styles.factValue}>{nextReviewLabel}</div>
                  </div>
                </div>
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="docs" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Tamanho</div>
                    <div className={styles.factValue}>{fileSizeLabel}</div>
                  </div>
                </div>
                {/* DEFER: confidentiality classification — no field on the artifact model yet.
                    Tracked in wiki/backlog/documento-publicado.md (trigger: classification field). */}
                <div className={styles.factCell}>
                  <div className={styles.factIcon}>
                    <Icon name="shield" size={14} />
                  </div>
                  <div className={styles.factContent}>
                    <div className={styles.factLabel}>Classificação</div>
                    <div className={styles.factValue}>{EM_DASH}</div>
                  </div>
                </div>
              </div>
            </div>

            {/* Coverage side card — artifact-specific (distribution denominator). Injected by the
                caller via the `aside` slot so the shared view never reaches into distribution data. */}
            {aside}
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

          {stageCount === 0 ? (
            <div className={styles.signoffEmpty}>
              Nenhum registro de aprovação para esta versão.
            </div>
          ) : (
            <div className={styles.signoffCard}>
              <div
                className={styles.signoffGrid}
                style={{ gridTemplateColumns: `repeat(${stageCount}, 1fr)` }}
              >
                <div
                  className={styles.signoffConnector}
                  style={{ left: connectorSide, right: connectorSide }}
                />
                {approvalGroups.map((group) => {
                  // Show first signoff slot per stage (one actor per stage in current workflow).
                  const item = group[0];
                  const signed = item.signedAt != null;
                  return (
                    <div key={item.stageIndex} className={styles.signoffStage}>
                      <div className={styles.signoffPin}>
                        <Icon name="check" size={16} />
                      </div>
                      <div className={styles.signoffStageName}>{item.label}</div>
                      {signed && (
                        <>
                          <div className={styles.signoffActor}>
                            <Avatar name={item.actorDisplay ?? item.actorUserId ?? ""} size="sm" />
                            <span className={styles.signoffActorName}>
                              {item.actorDisplay ?? item.actorUserId}
                            </span>
                          </div>
                          <div className={styles.signoffWhen}>{formatSignedAt(item.signedAt)}</div>
                        </>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </section>

        {/* Section: Histórico de versões */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>03 · Linhagem</div>
              <h2 className={styles.sectionTitle}>Histórico de versões</h2>
            </div>
            {versionEntries.length > 0 ? (
              <span className={styles.sectionAside}>passe o mouse para detalhar</span>
            ) : null}
          </div>
          {versionEntries.length > 0 ? (
            <VersionTimeline versions={versionEntries} />
          ) : (
            <div className={styles.signoffEmpty}>Nenhuma revisão governada encontrada.</div>
          )}
        </section>

        {/* Section: Documentos relacionados — backend não disponível, defer rastreado em wiki/backlog/documento-publicado.md */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>04 · Referências</div>
              <h2 className={styles.sectionTitle}>Documentos relacionados</h2>
            </div>
            <span className={styles.sectionAside}>não disponível</span>
          </div>
          <div className={styles.signoffEmpty}>
            O modelo de relacionamentos entre documentos ainda não está disponível.
          </div>
        </section>

        {/* Section: Comentários — backend não disponível, defer rastreado em wiki/backlog/documento-publicado.md */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>05 · Discussão interna</div>
              <h2 className={styles.sectionTitle}>Comentários</h2>
            </div>
            <span className={styles.sectionAside}>não disponível</span>
          </div>
          <div className={styles.signoffEmpty}>
            Comentários de exibição ainda não estão disponíveis para este documento.
          </div>
        </section>

      </div>
    </div>
  );
}
