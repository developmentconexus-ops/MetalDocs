import { describe, expect, it } from 'vitest';
import { commentPlainText } from './commentPlainText';

describe('commentPlainText', () => {
  it('flattens nested ProseMirror nodes to spaced text', () => {
    const content = [
      { type: 'paragraph', content: [{ type: 'text', text: 'Revisar' }, { type: 'text', text: 'seção 3' }] },
      { type: 'paragraph', content: [{ type: 'text', text: 'OK' }] },
    ];
    expect(commentPlainText(content)).toBe('Revisar seção 3 OK');
  });

  it('returns empty string for non-array or empty input', () => {
    expect(commentPlainText(undefined)).toBe('');
    expect(commentPlainText([])).toBe('');
    expect(commentPlainText([{ type: 'paragraph' }])).toBe('');
  });
});
