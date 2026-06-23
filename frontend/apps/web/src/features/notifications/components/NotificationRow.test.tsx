import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

import { NotificationRow } from './NotificationRow';
import type { Notification } from '../api/notifications';

function makeNotification(overrides: Partial<Notification> = {}): Notification {
  return {
    id: 'n-1',
    recipient_user_id: 'u-1',
    event_type: 'document.published',
    resource_type: 'document',
    resource_id: 'd-1',
    title: 'Documento publicado',
    message: 'O documento ABC foi publicado.',
    status: 'PENDING',
    created_at: '2026-06-22T11:59:30.000Z',
    read_at: null,
    ...overrides,
  };
}

describe('NotificationRow', () => {
  it('renders the event chip, title, and message', () => {
    render(<NotificationRow notification={makeNotification()} />);
    expect(screen.getByText('Publicado')).toBeInTheDocument();
    expect(screen.getByText('Documento publicado')).toBeInTheDocument();
    expect(screen.getByText('O documento ABC foi publicado.')).toBeInTheDocument();
  });

  it('calls onMarkRead with the id when an unread row is clicked', () => {
    const onMarkRead = vi.fn();
    render(<NotificationRow notification={makeNotification()} onMarkRead={onMarkRead} />);
    fireEvent.click(screen.getByRole('button', { name: /marcar como lida/i }));
    expect(onMarkRead).toHaveBeenCalledWith('n-1');
  });

  it('does not render a click target for a read row', () => {
    const onMarkRead = vi.fn();
    render(
      <NotificationRow
        notification={makeNotification({ status: 'READ', read_at: '2026-06-22T12:00:00.000Z' })}
        onMarkRead={onMarkRead}
      />,
    );
    expect(screen.queryByRole('button', { name: /marcar como lida/i })).not.toBeInTheDocument();
  });
});
