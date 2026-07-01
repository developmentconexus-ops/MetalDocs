import { describe, it, expect } from 'vitest';

import { buildTemplateApprovalChain } from './templateApprovalChain';
import type { VersionDTO } from '../api/templates';

// Field-to-field guard for the template → shared ApprovalChainItem[] mapping.
// Parity with documents/lib/approvalWorkflow.test.ts: templates are NOT
// signoff-less, so this asserts the same three-step submit→review→approve chain
// the shared cockpit flow-viz renders, with honest-omit (no fabricated names).

function makeVersion(overrides: Partial<VersionDTO> = {}): VersionDTO {
  return {
    id: 'v1',
    template_id: 'tpl-1',
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
  } as VersionDTO;
}

describe('buildTemplateApprovalChain', () => {
  it('always produces a stable three-step submit→review→approve chain', () => {
    const chain = buildTemplateApprovalChain(makeVersion());
    expect(chain).toHaveLength(3);
    expect(chain.map((s) => s.stageIndex)).toEqual([0, 1, 2]);
    expect(chain.map((s) => s.label)).toEqual(['Autoria', 'Revisão técnica', 'Aprovação']);
    expect(chain.map((s) => s.roleLabel)).toEqual(['Autoria', 'Revisão técnica', 'Aprovação']);
  });

  it('draft: authoring is current, downstream stages pending, no decisions/timestamps', () => {
    const [submit, review, approve] = buildTemplateApprovalChain(makeVersion({ status: 'draft' }));

    expect(submit.flowState).toBe('current');
    expect(submit.status).toBe('active');
    expect(submit.decision).toBeNull();
    expect(submit.signedAt).toBeNull();
    expect(submit.actorUserId).toBe('author-1');

    expect(review.flowState).toBe('pending');
    expect(review.decision).toBeNull();
    expect(approve.flowState).toBe('pending');
    expect(approve.decision).toBeNull();
  });

  it('under_review: submit sent, review current, approve pending; pending role fills actorDisplay', () => {
    const [submit, review, approve] = buildTemplateApprovalChain(
      makeVersion({
        status: 'under_review',
        submitted_at: '2026-01-02T10:00:00Z',
        pending_reviewer_role: 'quality-reviewer',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(submit.status).toBe('sent');
    expect(submit.decision).toBe('submitted');
    expect(submit.signedAt).toBe('2026-01-02T10:00:00Z');

    expect(review.flowState).toBe('current');
    // Honest-omit: no reviewer_id yet, so the pending role stands in for the actor.
    expect(review.actorUserId).toBeNull();
    expect(review.actorDisplay).toBe('quality-reviewer');
    expect(review.signedAt).toBeNull();

    expect(approve.flowState).toBe('pending');
  });

  it('approved: submit sent, review approved (done), approve current', () => {
    const [submit, review, approve] = buildTemplateApprovalChain(
      makeVersion({
        status: 'approved',
        submitted_at: '2026-01-02T10:00:00Z',
        reviewed_at: '2026-01-03T10:00:00Z',
        reviewer_id: 'reviewer-1',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(review.flowState).toBe('approved');
    expect(review.decision).toBe('approved');
    expect(review.actorUserId).toBe('reviewer-1');
    expect(review.actorDisplay).toBe('reviewer-1');
    expect(review.signedAt).toBe('2026-01-03T10:00:00Z');
    expect(approve.flowState).toBe('current');
  });

  it('published: all three stages resolved as done', () => {
    const [submit, review, approve] = buildTemplateApprovalChain(
      makeVersion({
        status: 'published',
        submitted_at: '2026-01-02T10:00:00Z',
        reviewed_at: '2026-01-03T10:00:00Z',
        approved_at: '2026-01-04T10:00:00Z',
        reviewer_id: 'reviewer-1',
        approver_id: 'approver-1',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(review.flowState).toBe('approved');
    expect(approve.flowState).toBe('approved');
    expect(approve.decision).toBe('approved');
    expect(approve.actorDisplay).toBe('approver-1');
    expect(approve.signedAt).toBe('2026-01-04T10:00:00Z');
  });
});
