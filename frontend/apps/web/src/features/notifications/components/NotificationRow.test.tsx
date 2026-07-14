import type { ReactElement } from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

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

function renderRow(ui: ReactElement) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe('NotificationRow', () => {
  it('renders the event chip, title, and message', () => {
    renderRow(<NotificationRow notification={makeNotification()} />);
    expect(screen.getByText('Publicado')).toBeInTheDocument();
    expect(screen.getByText('Documento publicado')).toBeInTheDocument();
    expect(screen.getByText('O documento ABC foi publicado.')).toBeInTheDocument();
  });

  it('deep-links an unread targeted row and marks it read on activation', () => {
    const onMarkRead = vi.fn();
    renderRow(<NotificationRow notification={makeNotification()} onMarkRead={onMarkRead} />);
    const link = screen.getByRole('link', { name: /abrir/i });
    expect(link).toHaveAttribute('href', '/documents/d-1');
    fireEvent.click(link);
    expect(onMarkRead).toHaveBeenCalledWith('n-1');
  });

  it('deep-links a read targeted row but does not re-mark it read', () => {
    const onMarkRead = vi.fn();
    renderRow(
      <NotificationRow
        notification={makeNotification({ status: 'READ', read_at: '2026-06-22T12:00:00.000Z' })}
        onMarkRead={onMarkRead}
      />,
    );
    const link = screen.getByRole('link', { name: /abrir/i });
    expect(link).toHaveAttribute('href', '/documents/d-1');
    fireEvent.click(link);
    expect(onMarkRead).not.toHaveBeenCalled();
  });

  it('keeps the mark-read overlay for an unread non-routable row', () => {
    const onMarkRead = vi.fn();
    renderRow(
      <NotificationRow
        notification={makeNotification({ resource_type: '', resource_id: '' })}
        onMarkRead={onMarkRead}
      />,
    );
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /marcar como lida/i }));
    expect(onMarkRead).toHaveBeenCalledWith('n-1');
  });

  it('has no click target for a read non-routable row', () => {
    const onMarkRead = vi.fn();
    renderRow(
      <NotificationRow
        notification={makeNotification({
          resource_type: '',
          resource_id: '',
          status: 'READ',
          read_at: '2026-06-22T12:00:00.000Z',
        })}
        onMarkRead={onMarkRead}
      />,
    );
    expect(screen.queryByRole('link')).not.toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
