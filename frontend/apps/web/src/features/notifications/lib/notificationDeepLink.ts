import type { Notification } from '../api/notifications';

// Maps a notification to its in-app resource route, or null when none applies.
// Fail-closed: only document notifications with a concrete id resolve; anything
// unknown or missing yields no link rather than a guessed path.
export function notificationDeepLink(n: Notification): string | null {
  if (n.resource_type === 'document' && typeof n.resource_id === 'string' && n.resource_id) {
    return `/documents/${n.resource_id}`;
  }
  return null;
}
