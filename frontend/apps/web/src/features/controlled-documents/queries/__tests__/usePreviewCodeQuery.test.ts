import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { usePreviewCodeQuery } from '../usePreviewCodeQuery';

vi.mock('../../api/controlledDocuments', () => ({
  previewCode: vi.fn().mockResolvedValue({
    profileCode: 'DC',
    areaCode: 'RH',
    nextSeq: 1,
    code: 'DC-RH-001',
  }),
}));

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

describe('usePreviewCodeQuery', () => {
  it('is disabled when profileCode is null', () => {
    const { result } = renderHook(() => usePreviewCodeQuery(null, 'RH'), {
      wrapper: wrap(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when areaCode is null', () => {
    const { result } = renderHook(() => usePreviewCodeQuery('DC', null), {
      wrapper: wrap(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches and returns code when both params are provided', async () => {
    const { result } = renderHook(() => usePreviewCodeQuery('DC', 'RH'), {
      wrapper: wrap(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.code).toBe('DC-RH-001');
  });
});
