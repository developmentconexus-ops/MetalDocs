import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

import { NotificationsSpotlight } from './NotificationsSpotlight';
import type { Notification } from '../api/notifications';
import { useNotificationsQuery } from '../queries/useNotificationQueries';
import {
  useMarkAllNotificationsReadMutation,
  useMarkNotificationReadMutation,
} from '../queries/useNotificationMutations';

vi.mock('../queries/useNotificationQueries', () => ({
  useNotificationsQuery: vi.fn(),
}));
vi.mock('../queries/useNotificationMutations', () => ({
  useMarkNotificationReadMutation: vi.fn(),
  useMarkAllNotificationsReadMutation: vi.fn(),
}));

const listMock = vi.mocked(useNotificationsQuery);
const markReadMock = vi.mocked(useMarkNotificationReadMutation);
const markAllMock = vi.mocked(useMarkAllNotificationsReadMutation);

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

// Minimal query/mutation result shims — only the fields the component reads.
function queryResult(over: Record<string, unknown>) {
  return { data: undefined, isLoading: false, isError: false, error: null, refetch: vi.fn(), ...over } as never;
}
function mutationResult(over: Record<string, unknown> = {}) {
  return { mutate: vi.fn(), isPending: false, ...over } as never;
}

beforeEach(() => {
  vi.clearAllMocks();
  markReadMock.mockReturnValue(mutationResult());
  markAllMock.mockReturnValue(mutationResult());
});

describe('NotificationsSpotlight', () => {
  it('shows the loading state', () => {
    listMock.mockReturnValue(queryResult({ isLoading: true }));
    render(<NotificationsSpotlight open onClose={vi.fn()} />);
    expect(screen.getByText('Carregando notificações…')).toBeInTheDocument();
  });

  it('shows the empty state', () => {
    listMock.mockReturnValue(queryResult({ data: { items: [], page: { has_more: false, next_cursor: null } } }));
    render(<NotificationsSpotlight open onClose={vi.fn()} />);
    expect(screen.getByText('Nenhuma notificação.')).toBeInTheDocument();
  });

  it('renders rows and an enabled mark-all when unread exist', () => {
    listMock.mockReturnValue(
      queryResult({ data: { items: [makeNotification()], page: { has_more: false, next_cursor: null } } }),
    );
    render(<NotificationsSpotlight open onClose={vi.fn()} />);
    expect(screen.getByText('Documento publicado')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Marcar todas como lidas' })).toBeEnabled();
  });

  it('disables mark-all when everything is already read', () => {
    listMock.mockReturnValue(
      queryResult({
        data: {
          items: [makeNotification({ status: 'READ', read_at: '2026-06-22T12:00:00.000Z' })],
          page: { has_more: false, next_cursor: null },
        },
      }),
    );
    render(<NotificationsSpotlight open onClose={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Marcar todas como lidas' })).toBeDisabled();
  });

  it('fires the bulk mutation when mark-all is clicked', () => {
    const mutate = vi.fn();
    markAllMock.mockReturnValue(mutationResult({ mutate }));
    listMock.mockReturnValue(
      queryResult({ data: { items: [makeNotification()], page: { has_more: false, next_cursor: null } } }),
    );
    render(<NotificationsSpotlight open onClose={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: 'Marcar todas como lidas' }));
    expect(mutate).toHaveBeenCalledTimes(1);
  });
});
