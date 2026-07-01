import { renderHook } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// ---------------------------------------------------------------------------
// Mocks — must be hoisted before any imports that pull the real modules
// ---------------------------------------------------------------------------

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (
    selector: (state: { user: { userId: string; displayName: string; roles: string[]; capabilities: string[] } }) => unknown,
  ) =>
    selector({
      user: {
        userId: 'actor-1',
        displayName: 'Test Actor',
        roles: ['reviewer', 'approver'],
        capabilities: ['template.submit', 'template.review', 'template.approve'],
      },
    }),
}));

vi.mock('../queries/useTemplateDetailQuery', () => ({
  useTemplateDetailQuery: vi.fn(),
}));

// ---------------------------------------------------------------------------
// Imports — after vi.mock calls
// ---------------------------------------------------------------------------

import { useTemplateDetailQuery } from '../queries/useTemplateDetailQuery';
import { useTemplateApprovalArtifact, type TemplateApprovalHandlers } from './useTemplateApprovalArtifact';

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const BASE_TEMPLATE = {
  id: 'tpl-1',
  tenant_id: 'tenant-1',
  key: 'TPL-001',
  name: 'Test Template',
  doc_type_code: 'POP',
  description: null,
  created_by: 'actor-1',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  published_version_id: null,
  current_revision_number: null,
  archived_at: null,
};

function makeVersion(overrides: Record<string, unknown> = {}) {
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
    author_id: 'actor-1',
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

function makeHandlers(): TemplateApprovalHandlers {
  return {
    runSubmit: vi.fn(),
    runReview: vi.fn(),
    runApprove: vi.fn(),
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useTemplateApprovalArtifact', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: { template: BASE_TEMPLATE, latest_version: makeVersion() },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);
  });

  it('returns model.kind === "template" and a 3-step submit→review→approve chain', () => {
    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );
    expect(result.current.model.kind).toBe('template');
    // Templates are not signoff-less: the adapter builds the same ApprovalChainItem[]
    // shape documents use, so the shared cockpit flow-viz renders for both kinds.
    const chain = result.current.model.approvalChain;
    expect(chain).not.toBeNull();
    expect(chain).toHaveLength(3);
    expect(chain?.map((s) => s.label)).toEqual(['Autoria', 'Revisão técnica', 'Aprovação']);
    // Draft: authoring is the current step, downstream stages still pending.
    expect(chain?.[0].flowState).toBe('current');
    expect(chain?.[1].flowState).toBe('pending');
    expect(chain?.[2].flowState).toBe('pending');
  });

  it('for under_review WITH reviewer + fully-capable actor → model.actions has 2 items, both available:true', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: {
        template: BASE_TEMPLATE,
        latest_version: makeVersion({
          status: 'under_review',
          pending_reviewer_role: 'reviewer',
          pending_approver_role: 'approver',
        }),
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );

    const { actions } = result.current.model;
    expect(actions).toHaveLength(2);
    expect(actions[0].key).toBe('accept');
    expect(actions[0].available).toBe(true);
    expect(actions[1].key).toBe('reject');
    expect(actions[1].available).toBe(true);
  });

  it('for published version → model.actions is []', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: {
        template: BASE_TEMPLATE,
        latest_version: makeVersion({ status: 'published' }),
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );

    expect(result.current.model.actions).toEqual([]);
  });

  it('returns the raw version and surfaces isLoading/isError from the base', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: {
        template: BASE_TEMPLATE,
        latest_version: makeVersion({ status: 'under_review', pending_reviewer_role: 'reviewer' }),
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );

    expect(result.current.version?.status).toBe('under_review');
    expect(result.current.isLoading).toBe(false);
    expect(result.current.isError).toBe(false);
  });

  it('isLoading is true when the detail query is loading', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      refetch: vi.fn(),
    } as never);

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );

    expect(result.current.isLoading).toBe(true);
  });

  it('isError is true when the detail query has no data', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
    } as never);

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers()),
    );

    expect(result.current.isError).toBe(true);
  });
});
