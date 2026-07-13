import { describe, it, expect } from 'vitest';

import { buildTemplateApprovalChain } from './templateApprovalChain';
import type { VersionDTO } from '../api/templates';

// Field-to-field guard for the template → shared ApprovalChainItem[] mapping.
// Kernel rebuild (unit 3.1a S3): three steps — Submissão → Aprovação →
// Publicação — driven solely by status + submitted_at/approved_at/published_at.
// The Aprovação/Publicação steps honest-omit the actor: the kernel completion
// writer never populates reviewer_id/approver_id, and VersionDTO carries no
// publisher-id field at all, so showing them would fabricate a signer.

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
  it('always produces a stable three-step Submissão→Aprovação→Publicação chain', () => {
    const chain = buildTemplateApprovalChain(makeVersion());
    expect(chain).toHaveLength(3);
    expect(chain.map((s) => s.stageIndex)).toEqual([0, 1, 2]);
    expect(chain.map((s) => s.label)).toEqual(['Submissão', 'Aprovação', 'Publicação']);
    expect(chain.map((s) => s.roleLabel)).toEqual(['Submissão', 'Aprovação', 'Publicação']);
  });

  it('draft: submission is current, downstream stages pending/waiting, no decisions/timestamps', () => {
    const [submit, approve, publish] = buildTemplateApprovalChain(makeVersion({ status: 'draft' }));

    expect(submit.flowState).toBe('current');
    expect(submit.status).toBe('active');
    expect(submit.decision).toBeNull();
    expect(submit.signedAt).toBeNull();
    expect(submit.actorUserId).toBe('author-1');
    expect(submit.actorDisplay).toBe('author-1');

    expect(approve.flowState).toBe('pending');
    expect(approve.status).toBe('waiting');
    expect(approve.decision).toBeNull();
    expect(publish.flowState).toBe('pending');
    expect(publish.status).toBe('waiting');
    expect(publish.decision).toBeNull();
  });

  it('under_review: submit sent, approve current, publish pending — honest-omit actor', () => {
    const [submit, approve, publish] = buildTemplateApprovalChain(
      makeVersion({
        status: 'under_review',
        submitted_at: '2026-01-02T10:00:00Z',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(submit.status).toBe('sent');
    expect(submit.decision).toBe('submitted');
    expect(submit.signedAt).toBe('2026-01-02T10:00:00Z');

    expect(approve.flowState).toBe('current');
    expect(approve.status).toBe('active');
    // Honest-omit: the kernel completion writer never populates reviewer_id/
    // approver_id, so there is no real signer identity to show yet.
    expect(approve.actorUserId).toBeNull();
    expect(approve.actorDisplay).toBeNull();
    expect(approve.signedAt).toBeNull();
    expect(approve.decision).toBeNull();

    expect(publish.flowState).toBe('pending');
  });

  it('approved: submit sent, approve done, publish current — approve step still honest-omits the actor', () => {
    const [submit, approve, publish] = buildTemplateApprovalChain(
      makeVersion({
        status: 'approved',
        submitted_at: '2026-01-02T10:00:00Z',
        approved_at: '2026-01-03T10:00:00Z',
        reviewer_id: 'reviewer-1',
        approver_id: 'approver-1',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(approve.flowState).toBe('approved');
    expect(approve.status).toBe('approved');
    expect(approve.decision).toBe('approved');
    // Even though reviewer_id/approver_id are set on the fixture, the mapper
    // never reads them — the kernel writer that produced this real state
    // never actually populates those columns (see module header).
    expect(approve.actorUserId).toBeNull();
    expect(approve.actorDisplay).toBeNull();
    expect(approve.signedAt).toBe('2026-01-03T10:00:00Z');

    expect(publish.flowState).toBe('current');
    expect(publish.status).toBe('active');
    expect(publish.decision).toBeNull();
  });

  it('published: all three stages resolved as done', () => {
    const [submit, approve, publish] = buildTemplateApprovalChain(
      makeVersion({
        status: 'published',
        submitted_at: '2026-01-02T10:00:00Z',
        approved_at: '2026-01-03T10:00:00Z',
        published_at: '2026-01-04T10:00:00Z',
      }),
    );

    expect(submit.flowState).toBe('sent');
    expect(approve.flowState).toBe('approved');
    expect(publish.flowState).toBe('approved');
    expect(publish.status).toBe('approved');
    expect(publish.decision).toBe('published');
    expect(publish.actorUserId).toBeNull();
    expect(publish.actorDisplay).toBeNull();
    expect(publish.signedAt).toBe('2026-01-04T10:00:00Z');
  });
});
