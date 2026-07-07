import { describe, expect, it } from 'vitest';
import { formatDueRelative } from '../dates';

describe('formatDueRelative', () => {
  const NOW = new Date('2026-04-15T12:00:00.000Z').getTime();

  it('null due date -> em-dash, not overdue', () => {
    expect(formatDueRelative(null, NOW)).toEqual({ label: '—', overdue: false });
  });

  it('invalid due date -> em-dash, not overdue', () => {
    expect(formatDueRelative('not-a-date', NOW)).toEqual({ label: '—', overdue: false });
  });

  it('due later today -> "vence hoje"', () => {
    const dueAt = '2026-04-15T20:00:00.000Z';
    expect(formatDueRelative(dueAt, NOW)).toEqual({ label: 'vence hoje', overdue: false });
  });

  it('due in the future (multiple days) -> "vence em N dia(s)"', () => {
    const dueAt = '2026-04-18T12:00:00.000Z';
    expect(formatDueRelative(dueAt, NOW)).toEqual({ label: 'vence em 3 dias', overdue: false });
  });

  it('due in exactly one future day -> singular "dia"', () => {
    const dueAt = '2026-04-16T12:00:00.000Z';
    expect(formatDueRelative(dueAt, NOW)).toEqual({ label: 'vence em 1 dia', overdue: false });
  });

  it('due in the past -> "atrasado há N dia(s)", overdue true', () => {
    const dueAt = '2026-04-12T12:00:00.000Z';
    expect(formatDueRelative(dueAt, NOW)).toEqual({ label: 'atrasado há 3 dias', overdue: true });
  });

  it('due 1 day in the past -> singular "dia", overdue true', () => {
    const dueAt = '2026-04-14T12:00:00.000Z';
    expect(formatDueRelative(dueAt, NOW)).toEqual({ label: 'atrasado há 1 dia', overdue: true });
  });

  it('defaults now to Date.now() when omitted', () => {
    const result = formatDueRelative(null);
    expect(result).toEqual({ label: '—', overdue: false });
  });
});
