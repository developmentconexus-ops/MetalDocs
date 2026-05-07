import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAreasQuery } from '../useAreasQuery';

vi.mock('../../../taxonomy/api/taxonomy', () => ({
  fetchAreas: vi.fn().mockResolvedValue([
    { code: 'rh', name: 'Recursos Humanos' },
    { code: 'fin', name: 'Financeiro' },
  ]),
}));

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

describe('useAreasQuery', () => {
  it('returns ProcessArea[] (unwrapped, canonical shape)', async () => {
    const { result } = renderHook(() => useAreasQuery(), { wrapper: wrap() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
    expect(result.current.data?.[0]).toHaveProperty('code');
    expect(result.current.data?.[0]).not.toHaveProperty('items');
    expect(result.current.data).toHaveLength(2);
    expect(result.current.data?.[0]).toMatchObject({ code: 'rh', name: 'Recursos Humanos' });
  });
});
