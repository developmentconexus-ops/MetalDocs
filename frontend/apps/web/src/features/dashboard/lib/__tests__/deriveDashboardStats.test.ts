import { describe, expect, it } from 'vitest';
import type { DocumentStatsResponse } from '../../../documents/api/library';
import { deriveDashboardStats } from '../deriveDashboardStats';

function statsWith(byStatus: Record<string, number>): DocumentStatsResponse {
  return { by_status: byStatus, by_area: {} };
}

describe('deriveDashboardStats', () => {
  it('maps by_status counts to the three re-scoped cards', () => {
    const items = deriveDashboardStats(
      statsWith({ approved: 12, under_review: 3, published: 40, draft: 7 }),
    );
    expect(items).toHaveLength(3);

    const [aprovados, revisao, publicados] = items;
    expect(aprovados.label).toBe('Aprovados');
    expect(aprovados.value).toBe('12');

    expect(revisao.label).toBe('Em revisão');
    expect(revisao.value).toBe('3');

    expect(publicados.label).toBe('Publicados');
    expect(publicados.value).toBe('40');
  });

  it('treats missing status keys as 0 — never NaN or a fabricated value', () => {
    const items = deriveDashboardStats(statsWith({}));
    expect(items.map((i) => i.value)).toEqual(['0', '0', '0']);
  });

  it('ignores a stray document-status "rejected" count (M4 F4.1: dead, no producer)', () => {
    const items = deriveDashboardStats(statsWith({ under_review: 4, rejected: 2 }));
    expect(items[1].value).toBe('4');
  });
});
