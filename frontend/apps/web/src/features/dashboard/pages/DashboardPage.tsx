import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth.store';
import { useDashboardInboxQuery } from '../queries/useDashboardInboxQuery';
import { useLibraryStatsQuery } from '../../documents/queries/useLibraryStatsQuery';
import { useDashboardActivityQuery } from '../queries/useDashboardActivityQuery';
import { deriveDashboardStats } from '../lib/deriveDashboardStats';
import { deriveActivityItems } from '../lib/deriveActivity';
import type { InboxItem } from '../../approval/api/approvalTypes';
import styles from './DashboardPage.module.css';

// ─── Helpers ─────────────────────────────────────────────────────

function getFirstName(displayName: string): string {
  return displayName.split(' ')[0] ?? displayName;
}

function getKindFromCode(code: string): string {
  return code.split('-')[0] ?? '';
}

function formatRelative(isoString: string): string {
  const diffMs = Date.now() - new Date(isoString).getTime();
  const diffH = Math.floor(diffMs / 3_600_000);
  if (diffH < 1) {
    const diffM = Math.floor(diffMs / 60_000);
    return diffM < 2 ? 'agora' : `há ${diffM}min`;
  }
  if (diffH < 24) return `há ${diffH}h`;
  const diffD = Math.floor(diffH / 24);
  return diffD === 1 ? 'ontem' : `há ${diffD} dias`;
}

function isUrgent(isoString: string): boolean {
  const diffH = (Date.now() - new Date(isoString).getTime()) / 3_600_000;
  return diffH < 24;
}

function formatHeroDate(): string {
  const d = new Date();
  const dow = ['DOM', 'SEG', 'TER', 'QUA', 'QUI', 'SEX', 'SAB'][d.getDay()];
  const month = ['JAN', 'FEV', 'MAR', 'ABR', 'MAI', 'JUN', 'JUL', 'AGO', 'SET', 'OUT', 'NOV', 'DEZ'][d.getMonth()];
  return `ED. ${dow} ${String(d.getDate()).padStart(2, '0')} · ${month} · ${d.getFullYear()} · VOL.III`;
}

function pendingLabel(count: number): string {
  if (count === 0) return 'nenhuma decisão';
  if (count === 1) return 'uma decisão';
  return `${count} decisões`;
}

// ─── Sub-components ───────────────────────────────────────────────

function PendingRow({ item, rank }: { item: InboxItem; rank: number }) {
  const navigate = useNavigate();
  // Unit 4.2 subject-generic: document rows carry a controlled document;
  // template rows carry none (null) and identify by subject_ref (templateId).
  // F-QA4-8: the canonical human code (controlled_documents.code) is what the
  // reader recognizes — and what getKindFromCode can actually parse. The uuid
  // stays as the truthful fallback, and '—' when even that is absent (wire
  // drift): an honest non-value, never a blank that reads as a working row
  // (no-fallback-principle).
  const code =
    item.subject_kind === 'document'
      ? (item.controlled_document_code ?? item.controlled_document_id ?? '—')
      : item.subject_ref;
  const kind = getKindFromCode(code);
  const urgent = isUrgent(item.submitted_at);
  const submittedAgo = formatRelative(item.submitted_at);

  return (
    <div
      className={styles.pendingItem}
      role="button"
      tabIndex={0}
      onClick={() => navigate('/approvals')}
      onKeyDown={(e) => e.key === 'Enter' && navigate('/approvals')}
    >
      <div className={styles.itemRank}>
        <span className={styles.itemRankNum}>{String(rank).padStart(2, '0')}</span>
        <span className={styles.itemRankKind}>{kind}</span>
      </div>

      <div>
        <div className={styles.itemCode}>{code}</div>
        <h3 className={styles.itemTitle}>{item.subject_title}</h3>
        <div className={styles.itemMeta}>
          <span style={{ color: 'var(--text-soft)', fontWeight: 500 }}>{item.submitted_by}</span>
          <span className={styles.itemMetaSep}>—</span>
          <span style={{ fontStyle: 'italic' }}>{item.area_code}</span>
          {item.stage_label && (
            <>
              <span className={styles.itemMetaSep}>·</span>
              <span>{item.stage_label}</span>
            </>
          )}
          {item.quorum_progress && (
            <>
              <span className={styles.itemMetaSep}>·</span>
              <span style={{ fontFamily: 'var(--font-mono)' }}>{item.quorum_progress}</span>
            </>
          )}
        </div>
      </div>

      <div className={styles.itemDeadline}>
        <div
          className={`${styles.itemDeadlineLabel} ${urgent ? styles.itemDeadlineUrgent : styles.itemDeadlineNormal}`}
        >
          {urgent && <span className={styles.itemUrgentDot} />}
          {submittedAgo}
        </div>
        <button
          className={styles.reviewBtn}
          onClick={(e) => {
            e.stopPropagation();
            navigate('/approvals');
          }}
        >
          Revisar →
        </button>
      </div>
    </div>
  );
}

// ─── Main page ────────────────────────────────────────────────────

export function Component() {
  const user = useAuthStore((s) => s.user);
  const { data, isLoading, isError } = useDashboardInboxQuery();
  const {
    data: statsData,
    isLoading: statsLoading,
    isError: statsError,
  } = useLibraryStatsQuery();

  const firstName = user ? getFirstName(user.displayName) : 'você';
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

  // Static card scaffolding (labels/subs/colors) derived from an empty stats
  // shape; live values fill in once `statsData` arrives. Loading → '…',
  // error → '—' (honest non-value, never a fabricated number).
  const statCards = deriveDashboardStats(statsData ?? { by_status: {}, by_area: {} });
  const statValue = (i: number): string =>
    statsLoading ? '…' : statsError || !statsData ? '—' : statCards[i].value;

  const {
    data: activityData,
    isLoading: activityLoading,
    isError: activityError,
  } = useDashboardActivityQuery();
  const activityItems = deriveActivityItems(activityData?.items ?? []);

  return (
    <div className={styles.page}>
      {/* ── Hero ──────────────────────────────────────── */}
      <div className={styles.hero}>
        <div className={styles.heroBg} />
        <div className={styles.heroDecorNum} aria-hidden="true">
          {new Date().getDate()}
        </div>

        <div className={styles.heroContent}>
          <div className={styles.heroKicker}>
            <span>{formatHeroDate()}</span>
            <span className={styles.heroKickerDot} />
          </div>

          <h1 className={styles.heroTitle}>
            Bom dia,
            <br />
            <span className={styles.heroTitleName}>{firstName}.</span>
            <span className={styles.heroSubtitle}>
              Hoje você tem{' '}
              <span className={styles.heroSubtitleAccent}>
                {isLoading ? '…' : pendingLabel(total)}
                <span className={styles.heroSubtitleAccentUnderline} />
              </span>{' '}
              que param na sua mesa.
            </span>
          </h1>

          <div className={styles.heroStats}>
            <div className={styles.statPill}>
              <div className={styles.statPillLabel}>AGUARDANDO VOCÊ</div>
              <div className={styles.statPillValue} style={{ color: total > 0 ? 'var(--accent)' : 'white' }}>
                {isLoading ? '…' : total}
              </div>
              <div className={styles.statPillSub}>pendências</div>
            </div>
            {statCards.map((s, i) => (
              <div key={s.label} className={styles.statPill}>
                <div className={styles.statPillLabel}>{s.label.toUpperCase()}</div>
                <div className={styles.statPillValue} style={{ color: 'white' }}>
                  {statValue(i)}
                </div>
                <div className={styles.statPillSub}>{s.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* ── Body ──────────────────────────────────────── */}
      <div className={styles.body}>

        {/* Left: pending approvals */}
        <section className={styles.leadColumn}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionTitle}>§ AGUARDANDO SUA ASSINATURA</span>
            <span className={styles.sectionCount}>
              {!isLoading && !isError && `${items.length} de ${total} visíveis`}
            </span>
          </div>

          {isLoading && (
            <div className={styles.stateBox}>
              <span className={styles.emptyIcon}>⋯</span>
              Carregando pendências…
            </div>
          )}

          {isError && (
            <div className={`${styles.stateBox} ${styles.stateError}`} role="alert">
              Não foi possível carregar a caixa de entrada.
            </div>
          )}

          {!isLoading && !isError && items.length === 0 && (
            <div className={styles.stateBox}>
              <span className={styles.emptyIcon}>✓</span>
              Nenhuma pendência. Sua caixa está limpa.
            </div>
          )}

          {!isLoading && !isError && items.map((item, i) => (
            <PendingRow key={item.instance_id} item={item} rank={i + 1} />
          ))}
        </section>

        {/* Right: stats + activity */}
        <aside className={styles.sidebar}>

          {/* § SEU PULSO */}
          <div className={styles.sideSection}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>§ SEU PULSO</span>
              <span className={styles.sectionCount}>5 dias úteis</span>
            </div>
            {statCards.map((s, i) => (
              <div key={s.label} className={styles.statRow}>
                <div className={styles.statRowInfo}>
                  <div className={styles.statRowLabel}>{s.label.toUpperCase()}</div>
                  <div className={styles.statRowValue} style={{ color: s.color }}>
                    {statValue(i)}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* § MURMÚRIOS */}
          <div className={styles.sideSection}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>§ MURMÚRIOS</span>
            </div>
            <div className={styles.activityList}>
              <span className={styles.activityLine} />

              {activityLoading && (
                <div className={styles.activityItem}>
                  <span className={styles.activityDot} />
                  <div className={styles.activityMuted}>Carregando atividade…</div>
                </div>
              )}

              {!activityLoading && activityError && (
                <div className={styles.activityItem} role="alert">
                  <span className={styles.activityDot} />
                  <div className={styles.activityMuted}>
                    Atividade indisponível no momento.
                  </div>
                </div>
              )}

              {!activityLoading && !activityError && activityItems.length === 0 && (
                <div className={styles.activityItem}>
                  <span className={styles.activityDot} />
                  <div className={styles.activityMuted}>Nenhuma atividade recente.</div>
                </div>
              )}

              {!activityLoading && !activityError && activityItems.map((a) => (
                <div key={a.id} className={styles.activityItem}>
                  <span className={styles.activityDot} />
                  <div className={styles.activityMuted}>
                    <span className={styles.activityWho}>{a.who}</span>
                    {a.what && ` ${a.what} `}
                    {a.code && <span className={styles.activityCode}>{a.code}</span>}
                  </div>
                  <span className={styles.activityTime}>{formatRelative(a.occurredAt)}</span>
                </div>
              ))}
            </div>
          </div>

        </aside>
      </div>
    </div>
  );
}
