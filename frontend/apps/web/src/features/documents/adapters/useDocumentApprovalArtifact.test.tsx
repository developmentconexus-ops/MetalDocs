import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApprovalInstance } from '../../approval/api/approvalTypes';

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (state: { user: { userId: string; displayName: string } }) => unknown) =>
    selector({ user: { userId: 'admin-user', displayName: 'Administrator' } }),
}));

vi.mock('../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));
vi.mock('../queries/useControlledDocumentActiveDocumentQuery', () => ({
  useControlledDocumentActiveDocumentQuery: vi.fn(),
}));
vi.mock('../../approval/api/approvalApi', () => ({
  getInstance: vi.fn(),
}));

import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { getInstance } from '../../approval/api/approvalApi';
import {
  useDocumentApprovalArtifact,
  type DocumentApprovalHandlers,
} from './useDocumentApprovalArtifact';

const BASE_DOC = {
  id: 'doc-1',
  code: 'POP-QUA-0148',
  name: 'POP Limpeza de Linha',
  status: 'under_review',
  revision_version: 3,
  revision_number: 3,
  controlled_document_id: 'cd-1',
  current_revision_id: 'rev-1',
  created_by: 'admin-user',
};

function makeInstance(): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'route-1',
    tenant_id: 'tenant-1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-04-15T10:00:00.000Z',
    stages: [
      {
        id: 'stage-0',
        stage_index: 0,
        label: 'Qualidade',
        status: 'passed',
        signoffs: [
          {
            id: 'so-0',
            actor_user_id: 'approver-1',
            decision: 'approve',
            signature_method: 'password_reauth',
            signed_at: '2026-04-15T11:00:00.000Z',
          },
        ],
        actors: [
          { user_id: 'approver-1', display_name: 'approver-1', status: 'approved', decision: 'approve' },
        ],
      },
    ],
    etag: 'v1',
  };
}

const handlers: DocumentApprovalHandlers = {
  openSubmit: vi.fn(),
  cancelInstance: vi.fn(),
  openPublish: vi.fn(),
};

function makeDecisionInputs() {
  return { defaultOptionKey: null, submit: vi.fn().mockResolvedValue(undefined) };
}

function mockContext(approvalState: string, overrides: Record<string, unknown> = {}) {
  vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
    data: {
      document_id: 'doc-1',
      approval_state: approvalState,
      content_hash: 'hash-abc',
      revision_version: 3,
      approval_instance_id: 'inst-1',
      ...overrides,
    },
    isLoading: false,
    isError: false,
  } as never);
}

describe('useDocumentApprovalArtifact', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: BASE_DOC,
      refetch: vi.fn(),
    } as never);
    vi.mocked(getInstance).mockResolvedValue(makeInstance());
  });

  it('emits ONLY the submit action in draft state', async () => {
    mockContext('draft', { approval_instance_id: undefined });
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    const keys = result.current.model?.actions.map((a) => a.key);
    expect(keys).toEqual(['submit']);
  });

  it('emits ONLY the cancel action in under_review (signing routes through the decision panel)', async () => {
    mockContext('under_review');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    const actions = result.current.model?.actions ?? [];
    expect(actions.map((a) => a.key)).toEqual(['cancel']);
    const cancel = actions.find((a) => a.key === 'cancel');
    expect(cancel?.variant).toBe('secondary');
  });

  it('emits ONLY the publish action in approved state', async () => {
    mockContext('approved');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    expect(result.current.model?.actions.map((a) => a.key)).toEqual(['publish']);
  });

  it('emits ONLY the publish action in published state', async () => {
    mockContext('published');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    expect(result.current.model?.actions.map((a) => a.key)).toEqual(['publish']);
  });

  it.each(['scheduled', 'superseded', 'rejected', 'obsolete', 'cancelled'])(
    'emits NO actions in read-only / terminal state %s',
    async (state) => {
      mockContext(state);
      const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

      await waitFor(() => expect(result.current.instance).not.toBeNull());

      expect(result.current.model?.actions).toEqual([]);
    },
  );

  it('binds each action run to the supplied route handlers', async () => {
    mockContext('under_review');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    const cancel = result.current.model?.actions.find((a) => a.key === 'cancel');
    cancel?.run();
    expect(handlers.cancelInstance).toHaveBeenCalledTimes(1);
  });

  it('maps the approval instance stages/signoffs into approvalChain', async () => {
    mockContext('under_review');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    const chain = result.current.model?.approvalChain;
    expect(chain).not.toBeNull();
    expect(chain).toHaveLength(1);
    expect(chain?.[0]).toMatchObject({
      stageIndex: 0,
      label: 'Qualidade',
      status: 'passed',
      actorUserId: 'approver-1',
      actorDisplay: 'approver-1',
      decision: 'approve',
      signedAt: '2026-04-15T11:00:00.000Z',
    });
  });

  it('reports contextError when the active-document context query errors', () => {
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    } as never);

    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));
    expect(result.current.contextError).toBe(true);
  });

  it('reports noActiveContext when the document has no active approval context', () => {
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as never);

    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));
    expect(result.current.noActiveContext).toBe(true);
  });

  // ── FE-02: model.decision is the single ArtifactDecisionModel construction path
  // (previously built inline in SignoffDetailPage) ──────────────────────────────

  it('under_review with a resolved context + locked instance → model.decision offers approve/reject with password+legal+signer', async () => {
    mockContext('under_review');
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    const { decision } = result.current.model!;
    expect(decision).toBeDefined();
    expect(decision?.options.map((o) => o.key)).toEqual(['approve', 'reject']);
    expect(decision?.options.find((o) => o.key === 'reject')?.requiresReason).toBe(true);
    expect(decision?.password).toEqual({ label: 'Senha' });
    expect(decision?.legal).not.toBeNull();
    expect(decision?.signer).toEqual({
      name: 'Administrator',
      detail: undefined,
      note: 'Assinatura digital gerada no ato da confirmação.',
    });
  });

  it('draft state (signoff not allowed by policy) → model.decision is undefined', async () => {
    mockContext('draft', { approval_instance_id: undefined });
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    expect(result.current.model?.decision).toBeUndefined();
  });

  it('no active context → model.decision is undefined (sidebar not ready)', () => {
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as never);

    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, makeDecisionInputs()));
    expect(result.current.model?.decision).toBeUndefined();
  });

  it('decision.submit delegates to the injected decisionInputs.submit with (optionKey, reason, password)', async () => {
    mockContext('under_review');
    const decisionInputs = makeDecisionInputs();
    const { result } = renderHook(() => useDocumentApprovalArtifact('doc-1', handlers, decisionInputs));

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    await result.current.model?.decision?.submit({ optionKey: 'reject', reason: 'Falha', password: 'secret' });
    expect(decisionInputs.submit).toHaveBeenCalledWith({ optionKey: 'reject', reason: 'Falha', password: 'secret' });
  });

  it('passes defaultOptionKey through from decisionInputs (e.g. from a ?decision= deep-link)', async () => {
    mockContext('under_review');
    const { result } = renderHook(() =>
      useDocumentApprovalArtifact('doc-1', handlers, { defaultOptionKey: 'reject', submit: vi.fn() }),
    );

    await waitFor(() => expect(result.current.instance).not.toBeNull());

    expect(result.current.model?.decision?.defaultOptionKey).toBe('reject');
  });
});
