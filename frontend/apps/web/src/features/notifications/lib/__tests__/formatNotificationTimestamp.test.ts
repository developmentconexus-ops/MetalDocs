import { describe, expect, it } from 'vitest';

import { formatNotificationTimestamp } from '../formatNotificationTimestamp';

// Fixed reference instant so the relative math is deterministic.
const NOW = Date.parse('2026-06-22T12:00:00.000Z');
const ago = (ms: number) => new Date(NOW - ms).toISOString();

describe('formatNotificationTimestamp', () => {
  it('renders sub-minute as "agora"', () => {
    expect(formatNotificationTimestamp(ago(30_000), NOW)).toBe('agora');
  });

  it('renders minutes', () => {
    expect(formatNotificationTimestamp(ago(5 * 60_000), NOW)).toBe('há 5 min');
  });

  it('renders hours', () => {
    expect(formatNotificationTimestamp(ago(2 * 60 * 60_000), NOW)).toBe('há 2 h');
  });

  it('renders days under a week', () => {
    expect(formatNotificationTimestamp(ago(3 * 24 * 60 * 60_000), NOW)).toBe('há 3 d');
  });

  it('falls back to an absolute date past a week', () => {
    expect(formatNotificationTimestamp(ago(10 * 24 * 60 * 60_000), NOW)).toBe('12/06/2026');
  });

  it('returns a dash for an unparseable instant', () => {
    expect(formatNotificationTimestamp('not-a-date', NOW)).toBe('—');
  });
});
