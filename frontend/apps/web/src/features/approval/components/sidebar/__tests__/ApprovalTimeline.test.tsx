// Finding 1 (Task 10 review): the extend-SLA button gate
// (`stage.status === 'active' && canExtendSla` in ApprovalTimeline.tsx) was
// the one piece of the Task 10 brief — "offer the action only to someone who
// can perform it" — that had to be hand-verified by reading code, because no
// test asserted on it. These tests close that gap.
//
// IMPORTANT: this is a UX-promise guard, not a security control. The server's
// tier-2 `authz.Require` (ADR 0022) is the real authority — a viewer without
// `approval.sla_extend` who forges the request still gets a 403 there. This
// suite only proves the app does not *offer* the action to someone who can't
// use it; it must never be read as, or relied on as, the security boundary.
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApprovalInstance, StageInstance } from '../../../api/approvalTypes';
import { ApprovalTimeline } from '../ApprovalTimeline';

const hasCapMock = vi.fn();
vi.mock('../../../../iam/hooks/useHasCapability', () => ({
  useHasCapability: (cap: string) => hasCapMock(cap),
}));

const extendMutateAsyncMock = vi.fn();
vi.mock('../../../queries/useExtendSLA', () => ({
  useExtendSLA: () => ({ mutateAsync: extendMutateAsyncMock, isPending: false }),
}));

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    signoffs: [],
    actors: [],
    stage_kind: 'review',
    due_at: '2026-08-10T12:00:00.000Z',
    ...overrides,
  } as StageInstance;
}

function makeInstance(stages: StageInstance[]): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'route-1',
    tenant_id: 'tenant-1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-08-01T00:00:00.000Z',
    completed_at: null,
    stages,
    etag: '"v1"',
    frozen_content_hash: null,
  } as ApprovalInstance;
}

function renderTimeline(instance: ApprovalInstance) {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  }
  return render(<ApprovalTimeline instance={instance} loading={false} />, { wrapper: Wrapper });
}

const EXTEND_BUTTON_NAME = 'Prorrogar prazo';

describe('ApprovalTimeline extend-SLA gate', () => {
  beforeEach(() => {
    hasCapMock.mockReset();
    extendMutateAsyncMock.mockReset();
  });

  it('does not offer "Prorrogar prazo" when the viewer lacks approval.sla_extend', () => {
    hasCapMock.mockReturnValue(false);
    renderTimeline(makeInstance([makeStage({ status: 'active', due_at: '2026-08-10T12:00:00.000Z' })]));

    expect(hasCapMock).toHaveBeenCalledWith('approval.sla_extend');
    expect(screen.queryByRole('button', { name: EXTEND_BUTTON_NAME })).toBeNull();
  });

  it('offers "Prorrogar prazo" when the viewer holds approval.sla_extend and the stage is active with a due date', () => {
    hasCapMock.mockReturnValue(true);
    renderTimeline(makeInstance([makeStage({ status: 'active', due_at: '2026-08-10T12:00:00.000Z' })]));

    expect(screen.getByRole('button', { name: EXTEND_BUTTON_NAME })).toBeTruthy();
  });

  it('does not offer "Prorrogar prazo" for a non-active stage even when the viewer holds the capability', () => {
    hasCapMock.mockReturnValue(true);
    renderTimeline(makeInstance([makeStage({ status: 'pending', due_at: '2026-08-10T12:00:00.000Z' })]));

    expect(screen.queryByRole('button', { name: EXTEND_BUTTON_NAME })).toBeNull();
  });

  it('does not offer "Prorrogar prazo" when the active stage has no due_at (no-fallback: null means no deadline)', () => {
    hasCapMock.mockReturnValue(true);
    renderTimeline(makeInstance([makeStage({ status: 'active', due_at: null })]));

    expect(screen.queryByRole('button', { name: EXTEND_BUTTON_NAME })).toBeNull();
  });
});
