import type { DocumentStatsResponse } from '../../documents/api/library';

export interface StatItem {
  label: string;
  value: string;
  sub: string;
  color: string;
}

// The three re-scoped "§ SEU PULSO" cards, derived from the live
// `GET /documents/stats` `by_status` counts. Honest counts only — no time
// window and no average (no producer exists for those). See F1.1 spec.md.
// Status enum source of truth: documents/domain/model.go / DocumentSummaryStatus.
export function deriveDashboardStats(stats: DocumentStatsResponse): StatItem[] {
  const by = stats.by_status ?? {};
  const count = (key: string): number => by[key] ?? 0;

  return [
    {
      label: 'Aprovados',
      value: String(count('approved')),
      sub: 'no acervo',
      color: 'var(--success)',
    },
    {
      // M4 F4.1: document-status 'rejected' removed (dead, no producer — the
      // live returned-to-author signal is the approval-decision 'rejected'
      // under features/approval/*, not a documents.status value here).
      label: 'Em revisão',
      value: String(count('under_review')),
      sub: 'aguardam ajuste',
      color: 'var(--warning)',
    },
    {
      label: 'Publicados',
      value: String(count('published')),
      sub: 'no acervo',
      color: 'var(--brand)',
    },
  ];
}
