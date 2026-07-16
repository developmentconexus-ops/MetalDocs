import { describe, expect, it } from 'vitest';
import { sortByDueAsc } from './sortByDue';
import type { InboxItem } from '../../features/approval/api/approvalTypes';

function makeItem(overrides: Partial<InboxItem> = {}): InboxItem {
  return {
    instance_id: 'inst-1',
    subject_kind: 'document',
    subject_key: 'doc-1',
    subject_title: 'Doc',
    subject_ref: 'doc-1',
    controlled_document_id: 'cd-1',
    area_code: 'QUA',
    submitted_by: 'maria',
    submitted_at: '2026-04-14T10:00:00.000Z',
    stage_kind: 'review',
    stage_label: 'Revisão',
    quorum_progress: '0/1',
    due_at: null,
    ...overrides,
  };
}

describe('sortByDueAsc', () => {
  it('orders ascending by due_at', () => {
    const later = makeItem({ instance_id: 'later', due_at: '2026-05-01T00:00:00.000Z' });
    const sooner = makeItem({ instance_id: 'sooner', due_at: '2026-04-20T00:00:00.000Z' });
    const result = sortByDueAsc([later, sooner]);
    expect(result.map((i) => i.instance_id)).toEqual(['sooner', 'later']);
  });

  it('sorts null due_at last', () => {
    const noDue = makeItem({ instance_id: 'no-due', due_at: null });
    const withDue = makeItem({ instance_id: 'with-due', due_at: '2026-04-20T00:00:00.000Z' });
    const result = sortByDueAsc([noDue, withDue]);
    expect(result.map((i) => i.instance_id)).toEqual(['with-due', 'no-due']);
  });

  it('does not mutate the input array', () => {
    const items = [makeItem({ instance_id: 'a' }), makeItem({ instance_id: 'b' })];
    const original = [...items];
    sortByDueAsc(items);
    expect(items).toEqual(original);
  });
});
