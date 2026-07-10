import { describe, expect, it } from 'vitest';
import { commentText } from './commentText';

describe('commentText', () => {
  it('extracts text from a paragraph of text leaves', () => {
    const body = [{ type: 'paragraph', content: [{ type: 'text', text: 'Corrigir a seção 3.' }] }];
    expect(commentText(body)).toBe('Corrigir a seção 3.');
  });

  it('concatenates nested + multi-node content', () => {
    const body = [
      { type: 'paragraph', content: [{ type: 'text', text: 'Olá ' }, { type: 'text', text: 'João' }] },
      { type: 'paragraph', content: [{ type: 'text', text: 'segunda linha' }] },
    ];
    expect(commentText(body)).toBe('Olá João segunda linha');
  });

  it('returns empty string for a body with no text leaves (fail-closed, no crash)', () => {
    expect(commentText([{ type: 'horizontalRule' }])).toBe('');
    expect(commentText([])).toBe('');
    expect(commentText(undefined)).toBe('');
    expect(commentText('not-an-array')).toBe('');
  });
});
