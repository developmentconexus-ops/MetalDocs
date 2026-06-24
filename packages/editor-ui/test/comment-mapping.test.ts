import { describe, it, expect } from 'vitest';
import { toEigenpalComment, fromEigenpalComment, type EditorComment } from '../src/comment-mapping';

describe('comment-mapping', () => {
  const ec: EditorComment = {
    id: 7,
    parentId: 3,
    author: 'Ada Lovelace',
    createdAt: '2026-06-23T10:00:00Z',
    body: [{ type: 'paragraph' }],
    resolved: true,
  };

  it('toEigenpalComment renames fields and derives initials', () => {
    const c = toEigenpalComment(ec);
    expect(c).toEqual({
      id: 7,
      parentId: 3,
      author: 'Ada Lovelace',
      date: '2026-06-23T10:00:00Z',
      content: [{ type: 'paragraph' }],
      done: true,
      initials: 'AL',
    });
  });

  it('fromEigenpalComment renames fields, coerces resolved, drops initials', () => {
    const ec2 = fromEigenpalComment({
      id: 7,
      parentId: 3,
      author: 'Ada Lovelace',
      date: '2026-06-23T10:00:00Z',
      initials: 'AL',
      content: [{ type: 'paragraph' }] as never,
      done: undefined,
    });
    expect(ec2).toEqual({
      id: 7,
      parentId: 3,
      author: 'Ada Lovelace',
      createdAt: '2026-06-23T10:00:00Z',
      body: [{ type: 'paragraph' }],
      resolved: false,
    });
  });

  it('round-trips EditorComment -> eigenpal -> EditorComment (minus derived initials)', () => {
    expect(fromEigenpalComment(toEigenpalComment(ec))).toEqual(ec);
  });

  it('derives single initial for single-word author', () => {
    expect(toEigenpalComment({ ...ec, author: 'Plato' }).initials).toBe('P');
  });
});
