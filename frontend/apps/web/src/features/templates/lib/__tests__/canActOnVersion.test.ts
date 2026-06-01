import { describe, expect, it } from 'vitest';
import {
  canApprove,
  canPublish,
  canReview,
  canSubmit,
  type ActorContext,
} from '../canActOnVersion';
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
    pending_reviewer_role: null,
    pending_approver_role: null,
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
  return { roles: [], capabilities: [], ...overrides };
}

describe('canSubmit', () => {
  it('allows when draft + has template.submit', () => {
    const gate = canSubmit(makeVersion({ status: 'draft' }), actor({ capabilities: ['template.submit'] }));
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != draft', () => {
    const gate = canSubmit(makeVersion({ status: 'in_review' }), actor({ capabilities: ['template.submit'] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('status');
  });

  it('denies when capability missing', () => {
    const gate = canSubmit(makeVersion({ status: 'draft' }), actor({ capabilities: [] }));
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.submit');
  });
});

describe('canReview', () => {
  it('allows when in_review + has template.review + role matches pending_reviewer_role', () => {
    const gate = canReview(
      makeVersion({ status: 'in_review', pending_reviewer_role: 'reviewer' }),
      actor({ capabilities: ['template.review'], roles: ['reviewer'] }),
    );
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != in_review', () => {
    const gate = canReview(
      makeVersion({ status: 'draft', pending_reviewer_role: 'reviewer' }),
      actor({ capabilities: ['template.review'], roles: ['reviewer'] }),
    );
    expect(gate.allowed).toBe(false);
  });

  it('denies when no pending_reviewer_role (no-reviewer flow)', () => {
    const gate = canReview(
      makeVersion({ status: 'in_review', pending_reviewer_role: null }),
      actor({ capabilities: ['template.review'], roles: ['reviewer'] }),
    );
    expect(gate.allowed).toBe(false);
  });

  it('denies when capability missing', () => {
    const gate = canReview(
      makeVersion({ status: 'in_review', pending_reviewer_role: 'reviewer' }),
      actor({ capabilities: [], roles: ['reviewer'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.review');
  });

  it('denies when role binding mismatched', () => {
    const gate = canReview(
      makeVersion({ status: 'in_review', pending_reviewer_role: 'reviewer' }),
      actor({ capabilities: ['template.review'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('reviewer');
  });
});

describe('canApprove', () => {
  it('allows when approved + has template.approve + role matches pending_approver_role', () => {
    const gate = canApprove(
      makeVersion({ status: 'approved', pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(true);
  });

  it('allows when binding is null (no role gate)', () => {
    const gate = canApprove(
      makeVersion({ status: 'approved', pending_approver_role: null }),
      actor({ capabilities: ['template.approve'], roles: [] }),
    );
    expect(gate.allowed).toBe(true);
  });

  it('denies when status != approved', () => {
    const gate = canApprove(
      makeVersion({ status: 'in_review', pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
  });

  it('denies when role binding mismatched', () => {
    const gate = canApprove(
      makeVersion({ status: 'approved', pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['author'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('approver');
  });

  it('denies when capability missing', () => {
    const gate = canApprove(
      makeVersion({ status: 'approved', pending_approver_role: 'approver' }),
      actor({ capabilities: [], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.approve');
  });
});

describe('canPublish', () => {
  // The "Publish" button on in_review without reviewer actually calls approveVersion
  // (POST /approve → Service.Approve accepts + publishes + spawns next draft in-tx).
  // Backend capability is template.approve, not template.publish.
  it('allows when in_review without reviewer + has template.approve + matches pending_approver_role', () => {
    const gate = canPublish(
      makeVersion({ status: 'in_review', pending_reviewer_role: null, pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(true);
  });

  it('denies when in_review WITH reviewer (must go through review first)', () => {
    const gate = canPublish(
      makeVersion({ status: 'in_review', pending_reviewer_role: 'reviewer', pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
  });

  it('denies when status != in_review', () => {
    const gate = canPublish(
      makeVersion({ status: 'approved', pending_reviewer_role: null, pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
  });

  it('denies when capability missing', () => {
    const gate = canPublish(
      makeVersion({ status: 'in_review', pending_reviewer_role: null, pending_approver_role: 'approver' }),
      actor({ capabilities: [], roles: ['approver'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('template.approve');
  });

  it('denies when role binding mismatched', () => {
    const gate = canPublish(
      makeVersion({ status: 'in_review', pending_reviewer_role: null, pending_approver_role: 'approver' }),
      actor({ capabilities: ['template.approve'], roles: ['author'] }),
    );
    expect(gate.allowed).toBe(false);
    if (!gate.allowed) expect(gate.reason).toContain('approver');
  });
});
