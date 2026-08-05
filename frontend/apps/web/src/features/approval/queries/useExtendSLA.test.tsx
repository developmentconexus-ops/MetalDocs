import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { toast } from 'sonner';

import { ApiError } from '../../../lib/api';
import { QK } from '../../../lib/queryKeys';
import * as approvalApi from '../api/approvalApi';
import { useExtendSLA } from './useExtendSLA';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../api/approvalApi', async (importOriginal) => {
  const actual = await importOriginal<typeof approvalApi>();
  return {
    ...actual,
    extendSLA: vi.fn(),
  };
});

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
  });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function makeWrapperWithClient(client: QueryClient) {
  return function ClientWrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

describe('useExtendSLA', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls extendSLA with a trimmed reason and invalidates the approval cache on success', async () => {
    vi.mocked(approvalApi.extendSLA).mockResolvedValueOnce({
      id: 'instance-1',
      document_id: 'doc-1',
      route_id: 'route-1',
      tenant_id: 'tenant-1',
      status: 'in_progress',
      submitted_by: 'user-1',
      submitted_at: '2026-08-01T00:00:00.000Z',
      completed_at: null,
      stages: [],
      etag: '"v1"',
      frozen_content_hash: null,
    });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
    });
    const invalidateSpy = vi.spyOn(client, 'invalidateQueries');

    const { result } = renderHook(() => useExtendSLA(), {
      wrapper: makeWrapperWithClient(client),
    });

    result.current.mutate({
      instanceId: 'instance-1',
      dueAt: '2026-09-01T00:00:00.000Z',
      reason: '  motivo válido  ',
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(approvalApi.extendSLA).toHaveBeenCalledTimes(1);
    expect(approvalApi.extendSLA).toHaveBeenCalledWith('instance-1', {
      due_at: '2026-09-01T00:00:00.000Z',
      reason: 'motivo válido',
    });
    expect(toast.success).toHaveBeenCalledWith('Prazo prorrogado.');
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: QK.approval.all });
  });

  it('refuses a blank reason before it reaches the network', async () => {
    const { result } = renderHook(() => useExtendSLA(), { wrapper });

    result.current.mutate({
      instanceId: 'instance-1',
      dueAt: '2026-09-01T00:00:00.000Z',
      reason: '   ',
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(approvalApi.extendSLA).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Informe o motivo da prorrogação do prazo.');
    expect(toast.success).not.toHaveBeenCalled();
  });

  it('toasts the mapped PT-BR message when the server rejects a backward extension', async () => {
    vi.mocked(approvalApi.extendSLA).mockRejectedValueOnce(
      new ApiError('validation.sla_extension_not_forward', 422, 'not forward'),
    );

    const { result } = renderHook(() => useExtendSLA(), { wrapper });

    result.current.mutate({
      instanceId: 'instance-1',
      dueAt: '2026-01-01T00:00:00.000Z',
      reason: 'motivo',
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(toast.error).toHaveBeenCalledWith(
      'A nova data limite deve ser posterior ao prazo atual.',
    );
    expect(toast.success).not.toHaveBeenCalled();
  });

  it('toasts the mapped PT-BR message when there is no active stage to extend', async () => {
    vi.mocked(approvalApi.extendSLA).mockRejectedValueOnce(
      new ApiError('state.sla_extension_no_active_stage', 409, 'no active stage'),
    );

    const { result } = renderHook(() => useExtendSLA(), { wrapper });

    result.current.mutate({
      instanceId: 'instance-1',
      dueAt: '2026-09-01T00:00:00.000Z',
      reason: 'motivo',
    });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(toast.error).toHaveBeenCalledWith(
      'Não há etapa ativa com prazo configurado para prorrogar.',
    );
  });
});
