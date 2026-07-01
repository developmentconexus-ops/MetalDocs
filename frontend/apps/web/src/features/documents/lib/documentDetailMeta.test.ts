import { describe, expect, it } from 'vitest';
import { displayRevisionTitle, formatRevisionCode, formatFileSize, formatPageCount } from './documentDetailMeta';

describe('formatFileSize', () => {
  it('formats binary units with pt-BR text', () => {
    expect(formatFileSize(0)).toBe('0 B');
    expect(formatFileSize(512)).toBe('512 B');
    expect(formatFileSize(1024)).toBe('1 KB');
    expect(formatFileSize(1536)).toBe('1,5 KB');
    expect(formatFileSize(1024 * 1024)).toBe('1 MB');
    expect(formatFileSize(5 * 1024 * 1024)).toBe('5 MB');
  });

  it('returns em-dash for null/undefined/negative/NaN', () => {
    expect(formatFileSize(null)).toBe('—');
    expect(formatFileSize(undefined)).toBe('—');
    expect(formatFileSize(-1)).toBe('—');
    expect(formatFileSize(Number.NaN)).toBe('—');
  });
});

describe('formatPageCount', () => {
  it('renders the integer when present', () => {
    expect(formatPageCount(1)).toBe('1');
    expect(formatPageCount(42)).toBe('42');
  });

  it('returns em-dash for null/undefined/negative/NaN', () => {
    expect(formatPageCount(null)).toBe('—');
    expect(formatPageCount(undefined)).toBe('—');
    expect(formatPageCount(-3)).toBe('—');
    expect(formatPageCount(Number.NaN)).toBe('—');
  });
});

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
    expect(displayRevisionTitle('', 'REV00')).toBe('Criação do documento');
  });

  it('keeps user-authored titles for later revisions', () => {
    expect(displayRevisionTitle('Ajuste operacional', 'REV01')).toBe('Ajuste operacional');
  });
});
