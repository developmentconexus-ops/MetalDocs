import { api } from '../../../lib/api/client';
import { ApiError } from '../../../lib/api/errors';
import type { components, operations } from '../../../lib/api-types';

// Types come from the generated contract (F3.1/F3.4) — we never hand-roll the
// shape (frontend-structure §3 rule 4). Mirrors `dashboard/api/audit.ts`.
export type Notification = components['schemas']['Notification'];
export type NotificationsListResponse = components['schemas']['NotificationsListResponse'];
export type UnreadCountResponse = components['schemas']['UnreadCountResponse'];
export type MarkAllReadResponse = components['schemas']['MarkAllReadResponse'];
export type NotificationsListParams = NonNullable<
  operations['listNotifications']['parameters']['query']
>;

function asApiError(error: unknown, fallbackCode: string): ApiError {
  if (error instanceof ApiError) return error;
  const env = error as { code?: string; message?: string; details?: unknown } | null;
  return new ApiError(env?.code ?? fallbackCode, 0, env?.message ?? 'Erro inesperado', env?.details);
}

// fetchNotifications lists the caller's notifications (self-scope, newest first).
export async function fetchNotifications(
  params: NotificationsListParams = {},
): Promise<NotificationsListResponse> {
  const { data, error } = await api.GET('/notifications', { params: { query: params } });
  if (error) throw asApiError(error, 'notifications.list_failed');
  if (!data) throw new ApiError('notifications.empty_response', 0, 'Resposta vazia ao carregar notificações.');
  return data;
}

// fetchUnreadCount returns the caller's unread (PENDING/SENT) count.
export async function fetchUnreadCount(): Promise<number> {
  const { data, error } = await api.GET('/notifications/unread-count');
  if (error) throw asApiError(error, 'notifications.unread_count_failed');
  if (!data) throw new ApiError('notifications.empty_response', 0, 'Resposta vazia ao carregar contador.');
  return data.count;
}

// markNotificationRead marks one of the caller's notifications read (idempotent).
export async function markNotificationRead(id: string): Promise<void> {
  const { error } = await api.POST('/notifications/{id}/read', { params: { path: { id } } });
  if (error) throw asApiError(error, 'notifications.mark_read_failed');
}

// markAllNotificationsRead marks every unread notification of the caller read
// and returns how many were transitioned (idempotent, self-scope; F3.4 backend).
export async function markAllNotificationsRead(): Promise<number> {
  const { data, error } = await api.POST('/notifications/read-all');
  if (error) throw asApiError(error, 'notifications.mark_all_read_failed');
  return data?.updated ?? 0;
}
