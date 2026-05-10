import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth.store';
import { useDashboardInboxQuery } from '../queries/useDashboardInboxQuery';
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
  const code = item.controlled_document_id;
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
        <h3 className={styles.itemTitle}>{item.document_title}</h3>
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

// TODO: replace with real API data when /api/v2/approval/stats endpoint is available
const MOCK_STATS = [
  { label: 'Aprovados', value: '—', sub: 'esta semana', color: 'var(--success)' as const },
  { label: 'Devolvidos', value: '—', sub: 'aguardando', color: 'var(--warning)' as const },
  { label: 'Tempo médio', value: '—', sub: 'por decisão', color: 'var(--brand)' as const },
];

// TODO: replace with real API data when /api/v2/activity or audit endpoint supports it
const MOCK_ACTIVITY = [
  { who: 'Sistema', what: 'Nenhuma atividade recente carregada.', code: '', time: '' },
];

// ─── Main page ────────────────────────────────────────────────────

export function Component() {
  const user = useAuthStore((s) => s.user);
  const { data, isLoading, isError } = useDashboardInboxQuery();

  const firstName = user ? getFirstName(user.displayName) : 'você';
  const items = data?.items ?? [];
  const total = data?.total ?? 0;

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
            {MOCK_STATS.map((s) => (
              <div key={s.label} className={styles.statPill}>
                <div className={styles.statPillLabel}>{s.label.toUpperCase()}</div>
                <div className={styles.statPillValue} style={{ color: 'white' }}>
                  {s.value}
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
            {MOCK_STATS.map((s) => (
              <div key={s.label} className={styles.statRow}>
                <div className={styles.statRowInfo}>
                  <div className={styles.statRowLabel}>{s.label.toUpperCase()}</div>
                  <div className={styles.statRowValue} style={{ color: s.color }}>
                    {s.value}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* § MURMÚRIOS */}
          <div className={styles.sideSection}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>§ MURMÚRIOS</span>
              <div className={styles.liveBadge}>
                <span className={styles.liveDot} />
                ao vivo
              </div>
            </div>
            <div className={styles.activityList}>
              <span className={styles.activityLine} />
              {MOCK_ACTIVITY.map((a, i) => (
                <div key={i} className={styles.activityItem}>
                  <span className={styles.activityDot} />
                  <div style={{ color: 'var(--text-muted)' }}>
                    <span className={styles.activityWho}>{a.who}</span>
                    {a.what && ` ${a.what} `}
                    {a.code && <span className={styles.activityCode}>{a.code}</span>}
                  </div>
                  {a.time && <span className={styles.activityTime}>{a.time}</span>}
                </div>
              ))}
            </div>
          </div>

        </aside>
      </div>
    </div>
  );
}
