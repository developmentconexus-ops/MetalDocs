import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as api from '../api/tokens';
import { useTokenMutations } from './useTokenMutations';

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
afterEach(() => vi.restoreAllMocks());

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
}

describe('useTokenMutations', () => {
  it('create calls the api and resolves', async () => {
    const spy = vi.spyOn(api, 'createToken').mockResolvedValue({ id: '1' } as never);
    const { result } = renderHook(() => useTokenMutations(), { wrapper });
    result.current.create.mutate({ name: 'slogan', value: 'v', label: 'L' });
    await waitFor(() => expect(result.current.create.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledOnce();
  });

  it('remove calls deleteToken', async () => {
    const spy = vi.spyOn(api, 'deleteToken').mockResolvedValue(undefined as never);
    const { result } = renderHook(() => useTokenMutations(), { wrapper });
    result.current.remove.mutate('1');
    await waitFor(() => expect(result.current.remove.isSuccess).toBe(true));
    expect(spy).toHaveBeenCalledWith('1');
  });
});
