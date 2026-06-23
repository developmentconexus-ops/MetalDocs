import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { useNotificationsQuery, useUnreadCountQuery } from '../useNotificationQueries';
import { fetchNotifications, fetchUnreadCount } from '../../api/notifications';

vi.mock('../../api/notifications', () => ({
  fetchNotifications: vi.fn(),
  fetchUnreadCount: vi.fn(),
}));

const fetchNotificationsMock = vi.mocked(fetchNotifications);
const fetchUnreadCountMock = vi.mocked(fetchUnreadCount);

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useNotificationsQuery', () => {
  it('returns the list payload and forwards params to the fetcher', async () => {
    fetchNotificationsMock.mockResolvedValue({
      items: [],
      page: { has_more: false, next_cursor: null },
    });

    const { result } = renderHook(() => useNotificationsQuery({ status: 'PENDING', limit: 5 }), {
      wrapper: wrap(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(fetchNotificationsMock).toHaveBeenCalledWith({ status: 'PENDING', limit: 5 });
    expect(result.current.data?.items).toEqual([]);
  });
});

describe('useUnreadCountQuery', () => {
  it('returns the unread count number', async () => {
    fetchUnreadCountMock.mockResolvedValue(7);

    const { result } = renderHook(() => useUnreadCountQuery(), { wrapper: wrap() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(7);
  });
});
