import { describe, expect, it } from 'vitest';

import { notificationDeepLink } from '../notificationDeepLink';
import type { Notification } from '../../api/notifications';

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

describe('notificationDeepLink', () => {
  it('links a document notification to its workspace route', () => {
    expect(notificationDeepLink(makeNotification())).toBe('/documents/d-1');
  });

  it('returns null when resource_id is empty', () => {
    expect(notificationDeepLink(makeNotification({ resource_id: '' }))).toBeNull();
  });

  it('returns null for a non-document resource type', () => {
    expect(
      notificationDeepLink(makeNotification({ resource_type: 'approval_instance' })),
    ).toBeNull();
  });
});
