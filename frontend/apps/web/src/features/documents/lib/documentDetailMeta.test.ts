import { describe, expect, it } from 'vitest';
import { displayRevisionTitle, formatRevisionCode } from './documentDetailMeta';

describe('formatRevisionCode', () => {
  it('formats governed revision numbers as zero-based REV labels', () => {
    expect(formatRevisionCode(0)).toBe('REV00');
    expect(formatRevisionCode(1)).toBe('REV01');
  });

  it('falls back to REV00 when revision number is missing or invalid', () => {
    expect(formatRevisionCode(undefined)).toBe('REV00');
    expect(formatRevisionCode(Number.NaN)).toBe('REV00');
  });
});

describe('displayRevisionTitle', () => {
  it('defaults the first governed revision title', () => {
    expect(displayRevisionTitle('', 'REV00')).toBe('Criacao do documento');
  });

  it('keeps user-authored titles for later revisions', () => {
    expect(displayRevisionTitle('Ajuste operacional', 'REV01')).toBe('Ajuste operacional');
  });
});
