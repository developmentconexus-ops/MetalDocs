import { describe, expect, it } from 'vitest';
import { canPublish, canSignoff, canSubmit, type ActorContext } from '../canActOnVersion';
import type { VersionDTO } from '../../api/templates';

function makeVersion(overrides: Partial<VersionDTO> = {}): VersionDTO {
  return {
    id: 'v1',
    template_id: 't1',
    version_number: 1,
    revision_number: 0,
    status: 'draft',
    docx_storage_key: null,
    content_hash: null,
    metadata_schema: null,
    placeholder_schema: null,
    author_id: 'author-1',
    reviewer_id: null,
    approver_id: null,
    submitted_at: null,
    reviewed_at: null,
    approved_at: null,
    published_at: null,
    obsoleted_at: null,
    created_at: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

function actor(overrides: Partial<ActorContext> = {}): ActorContext {
  return { capabilities: [], ...overrides };
}

describe('canSubmit', () => {
  it('allows when draft + has template.submit', () => {
    const gate = canSubmit(makeVersion({ status: 'draft' }), actor({ capabilities: ['template.submit'] }));
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != draft', () => {
    const gate = canSubmit(makeVersion({ status: 'under_review' }), actor({ capabilities: ['template.submit'] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('status');
  });

  it('denies when capability missing', () => {
    const gate = canSubmit(makeVersion({ status: 'draft' }), actor({ capabilities: [] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.submit');
  });
});

describe('canSignoff', () => {
  it('allows when under_review + has template.approve', () => {
    const gate = canSignoff(makeVersion({ status: 'under_review' }), actor({ capabilities: ['template.approve'] }));
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != under_review', () => {
    const gate = canSignoff(makeVersion({ status: 'draft' }), actor({ capabilities: ['template.approve'] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('status');
  });

  it('denies when capability missing', () => {
    const gate = canSignoff(makeVersion({ status: 'under_review' }), actor({ capabilities: [] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.approve');
  });
});

describe('canPublish', () => {
  it('allows when approved + has template.publish', () => {
    const gate = canPublish(makeVersion({ status: 'approved' }), actor({ capabilities: ['template.publish'] }));
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != approved', () => {
    const gate = canPublish(makeVersion({ status: 'under_review' }), actor({ capabilities: ['template.publish'] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('status');
  });

  it('denies when capability missing', () => {
    const gate = canPublish(makeVersion({ status: 'approved' }), actor({ capabilities: [] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.publish');
  });

  it('denies template.approve alone — publish requires template.publish specifically', () => {
    const gate = canPublish(makeVersion({ status: 'approved' }), actor({ capabilities: ['template.approve'] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.publish');
  });
});
