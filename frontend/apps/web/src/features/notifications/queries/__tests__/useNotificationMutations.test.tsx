import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import {
  useMarkNotificationReadMutation,
  useMarkAllNotificationsReadMutation,
} from '../useNotificationMutations';
import { markNotificationRead, markAllNotificationsRead } from '../../api/notifications';

vi.mock('../../api/notifications', () => ({
  markNotificationRead: vi.fn(),
  markAllNotificationsRead: vi.fn(),
}));

const toastSuccess = vi.fn();
const toastError = vi.fn();
vi.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => toastSuccess(...args),
    error: (...args: unknown[]) => toastError(...args),
  },
}));

const markReadMock = vi.mocked(markNotificationRead);
const markAllMock = vi.mocked(markAllNotificationsRead);

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function wrapWith(qc: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useMarkNotificationReadMutation', () => {
  it('invalidates the unread-count and list keys on settle', async () => {
    markReadMock.mockResolvedValue(undefined);
    const qc = makeClient();
    const spy = vi.spyOn(qc, 'invalidateQueries');

    const { result } = renderHook(() => useMarkNotificationReadMutation(), {
      wrapper: wrapWith(qc),
    });

    await act(async () => {
      await result.current.mutateAsync('abc');
    });

    const keys = spy.mock.calls.map((c) => JSON.stringify(c[0]?.queryKey));
    expect(keys).toContain(JSON.stringify(['notifications', 'unread-count']));
    expect(keys).toContain(JSON.stringify(['notifications', 'list', {}]));
  });

  it('toasts on failure', async () => {
    markReadMock.mockRejectedValue(new Error('boom'));
    const qc = makeClient();

    const { result } = renderHook(() => useMarkNotificationReadMutation(), {
      wrapper: wrapWith(qc),
    });

    await act(async () => {
      await result.current.mutateAsync('abc').catch(() => undefined);
    });

    await waitFor(() => expect(toastError).toHaveBeenCalled());
  });
});

describe('useMarkAllNotificationsReadMutation', () => {
  it('toasts the transitioned count on success', async () => {
    markAllMock.mockResolvedValue(3);
    const qc = makeClient();

    const { result } = renderHook(() => useMarkAllNotificationsReadMutation(), {
      wrapper: wrapWith(qc),
    });

    await act(async () => {
      await result.current.mutateAsync();
    });

    expect(toastSuccess).toHaveBeenCalledWith('3 notificações marcadas como lidas.');
  });

  it('toasts an empty-state message when nothing was unread', async () => {
    markAllMock.mockResolvedValue(0);
    const qc = makeClient();

    const { result } = renderHook(() => useMarkAllNotificationsReadMutation(), {
      wrapper: wrapWith(qc),
    });

    await act(async () => {
      await result.current.mutateAsync();
    });

    expect(toastSuccess).toHaveBeenCalledWith('Nenhuma notificação pendente.');
  });
});
