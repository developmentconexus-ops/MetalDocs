import { describe, expect, it } from 'vitest';

import { eventTypeLabel } from '../eventTypeLabel';

describe('eventTypeLabel', () => {
  it.each([
    ['document.published', 'Publicado'],
    ['document.superseded', 'Substituído'],
    ['document.obsoleted', 'Obsoleto'],
    ['document.approved', 'Aprovado'],
    ['document.rejected', 'Reprovado'],
    ['approval.pending', 'Pendente'],
    ['approval.overdue', 'Vencida'],
  ])('maps known event %s to %s', (eventType, label) => {
    expect(eventTypeLabel(eventType)).toBe(label);
  });

  it('humanizes the last segment of an unknown dotted event', () => {
    expect(eventTypeLabel('document.archived')).toBe('Archived');
  });

  it('capitalizes an unknown event with no dot', () => {
    expect(eventTypeLabel('reminder')).toBe('Reminder');
  });

  it('falls back to a generic label for an empty event type', () => {
    expect(eventTypeLabel('')).toBe('Notificação');
  });
});
