import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { toast } from 'sonner';

import { QK } from '../../../lib/queryKeys';
import { ApprovalError } from '../api/mutationClient';
import * as routeAdminApi from '../api/routeAdminApi';
import type { ListRoutesResponse } from '../api/routeAdminApi';
import { useCreateRoute, useUpdateRoute } from './useRouteAdminMutations';

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../api/routeAdminApi', async (importOriginal) => {
  const actual = await importOriginal<typeof routeAdminApi>();
  return {
    ...actual,
    createRoute: vi.fn(),
    updateRoute: vi.fn(),
    seedRouteEtag: vi.fn(),
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

const createStage = () => ({
  order: 1,
  name: 'Review',
  required_capability: 'doc.signoff',
  selectors: [{ kind: 'role_in_fixed_area' as const, role: 'approver', area_code: 'ops' }],
  quorum: 'any_1_of' as const,
  drift_policy: 'reduce_quorum' as const,
});

describe('useRouteAdminMutations toast behaviour', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('toasts the mapped PT-BR message on a non-412 create failure', async () => {
    vi.mocked(routeAdminApi.createRoute).mockRejectedValueOnce(
      new ApprovalError('validation.profile_unknown', 422, 'unknown profile'),
    );

    const { result } = renderHook(() => useCreateRoute(), { wrapper });
    result.current.mutate({ profile_code: 'qa-preview', name: 'QA', stages: [createStage()] });

    await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
    expect(vi.mocked(toast.error).mock.calls[0][0]).toMatch(/perfil/i);
    expect(toast.success).not.toHaveBeenCalled();
  });

  it('does not duplicate the 412 toast (already emitted by mutationClient)', async () => {
    vi.mocked(routeAdminApi.updateRoute).mockRejectedValueOnce(
      new ApprovalError('conflict.stale_revision', 412, 'stale'),
    );

    const { result } = renderHook(() => useUpdateRoute(), { wrapper });
    result.current.mutate({ routeId: 'route-a', body: { name: 'x', stages: [createStage()] } });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('toasts success on a completed create', async () => {
    vi.mocked(routeAdminApi.createRoute).mockResolvedValueOnce({ route_id: 'new-1' });

    const { result } = renderHook(() => useCreateRoute(), { wrapper });
    result.current.mutate({ profile_code: 'ops', name: 'Ops', stages: [createStage()] });

    await waitFor(() => expect(toast.success).toHaveBeenCalledWith('Rota criada.'));
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('preserves stage_kind in the optimistic cache patch before onSettled refetch', async () => {
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false }, queries: { retry: false } },
    });
    const queryKey = QK.approval.routes.list();
    const initial: ListRoutesResponse = {
      routes: [
        {
          id: 'route-a',
          name: 'Rota A',
          tenant_id: 'tenant-1',
          profile_code: 'JUR',
          active: true,
          version: 1,
          stages: [
            {
              order: 1,
              name: 'Revisão',
              required_capability: 'doc.signoff',
              selectors: [{ kind: 'role_in_fixed_area', role: 'approver', area_code: 'ops' }],
              quorum: 'any_1_of',
              quorum_m: null,
              drift_policy: 'reduce_quorum',
              stage_kind: 'review',
            },
          ],
          created_at: '2026-01-01T00:00:00.000Z',
          updated_at: '2026-01-01T00:00:00.000Z',
        },
      ],
      total: 1,
    };
    client.setQueryData(queryKey, initial);

    // Never-resolving promise: lets us inspect the cache mid-flight, before
    // onSuccess/onSettled would refetch and mask a broken optimistic patch.
    vi.mocked(routeAdminApi.updateRoute).mockReturnValueOnce(new Promise(() => {}));

    const { result } = renderHook(() => useUpdateRoute(), {
      wrapper: makeWrapperWithClient(client),
    });

    result.current.mutate({
      routeId: 'route-a',
      body: {
        name: 'Rota A',
        stages: [{ ...createStage(), stage_kind: 'approval' }],
      },
    });

    await waitFor(() => {
      const cached = client.getQueryData<ListRoutesResponse>(queryKey);
      expect(cached?.routes[0]?.stages[0]?.stage_kind).toBe('approval');
    });
  });
});
