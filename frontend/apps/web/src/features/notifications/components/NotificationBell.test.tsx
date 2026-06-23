import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

import { NotificationBell } from './NotificationBell';
import type { Notification } from '../api/notifications';
import { useNotificationsQuery, useUnreadCountQuery } from '../queries/useNotificationQueries';
import { useMarkNotificationReadMutation } from '../queries/useNotificationMutations';

vi.mock('../queries/useNotificationQueries', () => ({
  useNotificationsQuery: vi.fn(),
  useUnreadCountQuery: vi.fn(),
}));
vi.mock('../queries/useNotificationMutations', () => ({
  useMarkNotificationReadMutation: vi.fn(),
}));
// Stub the spotlight so the bell test stays focused on the bell + popover.
vi.mock('./NotificationsSpotlight', () => ({
  NotificationsSpotlight: ({ open }: { open: boolean }) =>
    open ? <div data-testid="spotlight" /> : null,
}));

const listMock = vi.mocked(useNotificationsQuery);
const unreadMock = vi.mocked(useUnreadCountQuery);
const markReadMock = vi.mocked(useMarkNotificationReadMutation);

function makeNotification(overrides: Partial<Notification> = {}): Notification {
  return {
    id: 'n-1',
    recipient_user_id: 'u-1',
    event_type: 'document.published',
    resource_type: 'document',
    resource_id: 'd-1',
    title: 'Documento publicado',
    message: 'msg',
    status: 'PENDING',
    created_at: '2026-06-22T11:59:30.000Z',
    read_at: null,
    ...overrides,
  };
}

function queryResult(over: Record<string, unknown>) {
  return { data: undefined, isLoading: false, isError: false, error: null, refetch: vi.fn(), ...over } as never;
}

beforeEach(() => {
  vi.clearAllMocks();
  markReadMock.mockReturnValue({ mutate: vi.fn(), isPending: false } as never);
  listMock.mockReturnValue(
    queryResult({ data: { items: [makeNotification()], page: { has_more: false, next_cursor: null } } }),
  );
  unreadMock.mockReturnValue(queryResult({ data: 0 }));
});

describe('NotificationBell', () => {
  it('hides the badge when there are no unread', () => {
    unreadMock.mockReturnValue(queryResult({ data: 0 }));
    render(<NotificationBell />);
    expect(screen.getByRole('button', { name: 'Notificações' })).toBeInTheDocument();
  });

  it('shows the unread count in the badge', () => {
    unreadMock.mockReturnValue(queryResult({ data: 3 }));
    render(<NotificationBell />);
    expect(screen.getByRole('button', { name: 'Notificações (3 não lidas)' })).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('caps the badge at 9+', () => {
    unreadMock.mockReturnValue(queryResult({ data: 42 }));
    render(<NotificationBell />);
    expect(screen.getByText('9+')).toBeInTheDocument();
  });

  it('opens the popover on click and lists recent notifications', () => {
    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: 'Notificações' }));
    expect(screen.getByRole('dialog', { name: 'Notificações recentes' })).toBeInTheDocument();
    expect(screen.getByText('Documento publicado')).toBeInTheDocument();
  });

  it('opens the spotlight from "Mostrar todas" and closes the popover', () => {
    render(<NotificationBell />);
    fireEvent.click(screen.getByRole('button', { name: 'Notificações' }));
    fireEvent.click(screen.getByRole('button', { name: 'Mostrar todas' }));
    expect(screen.getByTestId('spotlight')).toBeInTheDocument();
    expect(screen.queryByRole('dialog', { name: 'Notificações recentes' })).not.toBeInTheDocument();
  });
});
