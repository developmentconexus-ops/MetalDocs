import { renderHook } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// ---------------------------------------------------------------------------
// Mocks — must be hoisted before any imports that pull the real modules
// ---------------------------------------------------------------------------

const authState = vi.hoisted(() => ({
  capabilities: ['template.submit', 'template.approve', 'template.publish'] as string[],
}));

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (
    selector: (state: {
      user: { userId: string; displayName: string; username: string; email: string | null; capabilities: string[] };
    }) => unknown,
  ) =>
    selector({
      user: {
        userId: 'actor-1',
        displayName: 'Test Actor',
        username: 'test.actor',
        email: 'test.actor@example.com',
        capabilities: authState.capabilities,
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
  latest_version: { id: 'v1', number: 1, revision_number: 0, status: 'draft' },
  published_version: null,
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
    runSignoff: vi.fn(),
    runPublish: vi.fn(),
  };
}

function makeDecisionSubmit() {
  return vi.fn().mockResolvedValue(undefined);
}

function mockDetail(status: string, overrides: Record<string, unknown> = {}) {
  vi.mocked(useTemplateDetailQuery).mockReturnValue({
    data: { template: BASE_TEMPLATE, latest_version: makeVersion({ status, ...overrides }) },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  } as never);
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useTemplateApprovalArtifact', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    authState.capabilities = ['template.submit', 'template.approve', 'template.publish'];
    mockDetail('draft');
  });

  it('returns model.kind === "template" and a 3-step Submissão→Aprovação→Publicação chain', () => {
    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );
    expect(result.current.model.kind).toBe('template');
    // Templates are not signoff-less: the adapter builds the same ApprovalChainItem[]
    // shape documents use, so the shared cockpit flow-viz renders for both kinds.
    const chain = result.current.model.approvalChain;
    expect(chain).not.toBeNull();
    expect(chain).toHaveLength(3);
    expect(chain?.map((s) => s.label)).toEqual(['Submissão', 'Aprovação', 'Publicação']);
    // Draft: submission is the current step, downstream stages still pending.
    expect(chain?.[0].flowState).toBe('current');
    expect(chain?.[1].flowState).toBe('pending');
    expect(chain?.[2].flowState).toBe('pending');
  });

  it('for under_review + capable actor → decision is offered and actions[] is suppressed (empty)', () => {
    mockDetail('under_review');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    // ArtifactApprovalScreen renders an "Outras ações" band for whatever is left
    // in actions[] even while a decision is offered — without this suppression
    // under_review would show the accept/reject pair twice (decision panel +
    // plain-action band).
    expect(result.current.model.decision).toBeDefined();
    expect(result.current.model.actions).toEqual([]);
  });

  it('for under_review + actor WITHOUT template.approve → decision is undefined and the disabled accept/reject fallback surfaces', () => {
    authState.capabilities = ['template.submit'];
    mockDetail('under_review');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    expect(result.current.model.decision).toBeUndefined();
    const { actions } = result.current.model;
    expect(actions).toHaveLength(2);
    expect(actions[0].key).toBe('accept');
    expect(actions[0].available).toBe(false);
    expect(actions[1].key).toBe('reject');
    expect(actions[1].available).toBe(false);
  });

  it('for approved + capable actor → model.actions offers a single available "publish" action', () => {
    mockDetail('approved');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    const { actions } = result.current.model;
    expect(actions).toHaveLength(1);
    expect(actions[0].key).toBe('publish');
    expect(actions[0].label).toBe('Publicar');
    expect(actions[0].available).toBe(true);
  });

  it('for published version → model.actions is []', () => {
    mockDetail('published');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    expect(result.current.model.actions).toEqual([]);
  });

  it('returns the raw version and surfaces isLoading/isError from the base', () => {
    mockDetail('under_review');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
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
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
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
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    expect(result.current.isError).toBe(true);
  });

  // ── FE-02: model.decision is the single ArtifactDecisionModel construction path
  // (previously built inline in TemplateApprovalRoute) ─────────────────────────

  it('for under_review + capable actor → model.decision offers accept/reject with password + signer, no legal', () => {
    mockDetail('under_review');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    const { decision } = result.current.model;
    expect(decision).toBeDefined();
    expect(decision?.options.map((o) => o.key)).toEqual(['accept', 'reject']);
    expect(decision?.options.find((o) => o.key === 'accept')?.label).toBe('Assinar aprovação');
    expect(decision?.options.find((o) => o.key === 'reject')?.requiresReason).toBe(true);
    expect(decision?.password).toEqual({ label: 'Senha' });
    expect(decision?.legal).toBeNull();
    expect(decision?.signer).toEqual({
      name: 'Test Actor',
      detail: 'test.actor@example.com',
      note: 'Assinatura digital gerada no ato da confirmação.',
    });
  });

  it('for approved status → model.decision is undefined (publish has no signature ceremony)', () => {
    mockDetail('approved');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    expect(result.current.model.decision).toBeUndefined();
  });

  it('for published version → model.decision is undefined', () => {
    mockDetail('published');

    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), makeDecisionSubmit()),
    );

    expect(result.current.model.decision).toBeUndefined();
  });

  it('decision.submit calls the injected decisionSubmit with (accept, reason, password)', async () => {
    mockDetail('under_review');

    const decisionSubmit = makeDecisionSubmit();
    const { result } = renderHook(() =>
      useTemplateApprovalArtifact('tpl-1', makeHandlers(), decisionSubmit),
    );

    await result.current.model.decision?.submit({
      optionKey: 'reject',
      reason: 'Conteúdo incorreto',
      password: 'segredo123',
    });
    expect(decisionSubmit).toHaveBeenCalledWith(false, 'Conteúdo incorreto', 'segredo123');
  });
});
