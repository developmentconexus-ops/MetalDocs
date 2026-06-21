import { describe, expect, it } from 'vitest';
import type { AuditEventItem } from '../../api/audit';
import { deriveActivityItems } from '../deriveActivity';

function event(over: Partial<AuditEventItem>): AuditEventItem {
  return {
    id: 'evt-1',
    occurred_at: '2026-06-21T12:00:00Z',
    actor_id: 'editor-1',
    action: 'workflow.transitioned',
    resource_type: 'document',
    resource_id: 'DOC-1',
    payload: {},
    trace_id: 'trace-1',
    ...over,
  };
}

describe('deriveActivityItems', () => {
  it('maps actor/resource/occurred_at into who/code/occurredAt', () => {
    const [item] = deriveActivityItems([event({ actor_id: 'ana', resource_id: 'POP-7' })]);
    expect(item.who).toBe('ana');
    expect(item.code).toBe('POP-7');
    expect(item.occurredAt).toBe('2026-06-21T12:00:00Z');
  });

  it('humanizes a known action into PT text', () => {
    const [item] = deriveActivityItems([event({ action: 'workflow.transitioned' })]);
    expect(item.what.length).toBeGreaterThan(0);
    expect(item.what).not.toBe('workflow.transitioned');
  });

  it('falls back to a readable form for an unknown action — never blank', () => {
    const [item] = deriveActivityItems([event({ action: 'gizmo.frobnicated' })]);
    expect(item.what.length).toBeGreaterThan(0);
  });

  it('returns an empty array for no events', () => {
    expect(deriveActivityItems([])).toEqual([]);
  });
});
