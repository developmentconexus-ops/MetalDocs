import type React from "react";
import { Link } from "react-router-dom";
import { Icon } from "../../../components/ui/Icon";
import { Avatar } from "../../../components/ui/Avatar";
import { formatShortDate, formatSignedAt } from "../../../lib/format/dates";
import { formatFileSize } from "../../../lib/format/fileSize";
import { ArtifactHero } from "./ArtifactHero";
import { ArtifactHeroDocCard, ArtifactHeroBadges } from "./ArtifactHeroCard";
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

/**
 * Pick the item that speaks for a whole stage in the collapsed "Sign-offs desta
 * versão" row.
 *
 * `group[0]` is NOT that item: the backend emits a stage's actor roster with the
 * recorded signoffs first, ordered by ascending `signed_at`, and only then the
 * still-pending actors. On a stage that several people acted on, the head of the
 * bucket is therefore the EARLIEST decider — for a rejected stage with prior
 * approvals that is the first approver, so the row would name the wrong person
 * at the wrong time. Resolve the decisive actor explicitly instead:
 *   (a) a rejection terminates the stage, so it owns the stage outcome;
 *   (b) otherwise the LATEST recorded approval is the stage's standing decision;
 *   (c) otherwise nobody has decided — show whoever is on the clock;
 *   (d) otherwise fall back to the bucket head (single-slot / roster-less stage).
 */
function pickStageDecisiveItem(group: ApprovalChainItem[]): ApprovalChainItem {
  const rejected = group.find(
    (item) => item.flowState === "rejected" || item.status === "rejected",
  );
  if (rejected) return rejected;

  let latestApproved: ApprovalChainItem | null = null;
  let latestApprovedAt = Number.NEGATIVE_INFINITY;
  for (const item of group) {
    if (item.flowState !== "approved" || item.signedAt == null) continue;
    const signedAt = new Date(item.signedAt).getTime();
    // An unparseable timestamp still counts as an approval, it just never wins
    // the recency comparison against a real one.
    const rank = Number.isNaN(signedAt) ? Number.NEGATIVE_INFINITY : signedAt;
    if (latestApproved == null || rank > latestApprovedAt) {
      latestApproved = item;
      latestApprovedAt = rank;
    }
  }
  if (latestApproved) return latestApproved;

  const awaiting = group.find(
    (item) => item.flowState === "current" || item.flowState === "pending",
  );
  if (awaiting) return awaiting;

  return group[0];
}

function toVersionEntries(model: ArtifactViewModel): VersionEntry[] {
  return model.lineage.map((item) => ({
    v: item.revisionLabel ?? `REV${String(item.revisionNumber ?? 0).padStart(2, "0")}`,
    when: formatShortDate(item.createdAt) || EM_DASH,
    // VersionEntry.meta trails the date with the lifecycle status — no per-revision author field exists on the model yet.
    meta: item.status || EM_DASH,
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
        docCard={<ArtifactHeroDocCard model={model} />}
        badges={<ArtifactHeroBadges badges={model.hero.badges} />}
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
                  // One row per stage: the actor whose decision the stage stands on
                  // (see pickStageDecisiveItem — never blindly the bucket head).
                  const item = pickStageDecisiveItem(group);
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

        {/* Section: Artefatos relacionados — backend não disponível, defer rastreado em wiki/backlog/documento-publicado.md */}
        <section className={styles.section}>
          <div className={styles.sectionHead}>
            <div>
              <div className={styles.sectionKicker}>04 · Referências</div>
              <h2 className={styles.sectionTitle}>Artefatos relacionados</h2>
            </div>
            <span className={styles.sectionAside}>não disponível</span>
          </div>
          <div className={styles.signoffEmpty}>
            O modelo de relacionamentos entre artefatos controlados ainda não está disponível.
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
            Comentários de exibição ainda não estão disponíveis para este artefato.
          </div>
        </section>

      </div>
    </div>
  );
}
