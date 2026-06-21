import { describe, expect, it } from 'vitest';
import type { DocumentStatsResponse } from '../../../documents/api/library';
import { deriveDashboardStats } from '../deriveDashboardStats';

function statsWith(byStatus: Record<string, number>): DocumentStatsResponse {
  return { by_status: byStatus, by_area: {} };
}

describe('deriveDashboardStats', () => {
  it('maps by_status counts to the three re-scoped cards', () => {
    const items = deriveDashboardStats(
      statsWith({ approved: 12, under_review: 3, rejected: 2, published: 40, draft: 7 }),
    );
    expect(items).toHaveLength(3);

    const [aprovados, revisao, publicados] = items;
    expect(aprovados.label).toBe('Aprovados');
    expect(aprovados.value).toBe('12');

    // Em revisão = under_review (3) + rejected/devolvidos (2) = 5
    expect(revisao.label).toBe('Em revisão');
    expect(revisao.value).toBe('5');

    expect(publicados.label).toBe('Publicados');
    expect(publicados.value).toBe('40');
  });

  it('treats missing status keys as 0 — never NaN or a fabricated value', () => {
    const items = deriveDashboardStats(statsWith({}));
    expect(items.map((i) => i.value)).toEqual(['0', '0', '0']);
  });

  it('sums only under_review and rejected into the review card', () => {
    const items = deriveDashboardStats(statsWith({ under_review: 4 }));
    expect(items[1].value).toBe('4');
  });
});
