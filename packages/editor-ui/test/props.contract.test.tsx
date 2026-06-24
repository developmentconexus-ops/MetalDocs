import { describe, it, expect } from 'vitest';
import type { MetalDocsEditorProps } from '../src/types';
import type { EditorComment } from '../src/comment-mapping';

/**
 * Contract guard: MetalDocsEditorProps exposes only neutral (non-vendor) types.
 *
 * After ACL Phase 3B, the comment surface uses EditorComment — NOT the eigenpal
 * Comment type — so forward/reverse assignability with DocxEditorProps for comment
 * fields is intentionally broken at the type level. These tests verify the neutral
 * surface shape instead.
 */

describe('MetalDocsEditorProps contract', () => {
  it('accepts EditorComment[] for comments prop', () => {
    const comments: EditorComment[] = [
      { id: 1, author: 'Alice', body: null, resolved: false },
    ];
    // Type-level check: MetalDocsEditorProps.comments accepts EditorComment[].
    const _props: Pick<MetalDocsEditorProps, 'comments'> = { comments };
    expect(_props.comments).toHaveLength(1);
  });

  it('accepts onCommentsChange with EditorComment[] callback', () => {
    let received: EditorComment[] = [];
    const handler = (cs: EditorComment[]) => { received = cs; };
    const _props: Pick<MetalDocsEditorProps, 'onCommentsChange'> = { onCommentsChange: handler };
    _props.onCommentsChange?.([{ id: 2, author: 'Bob', body: null, resolved: true }]);
    expect(received).toHaveLength(1);
    expect(received[0]!.id).toBe(2);
  });

  it('accepts onCommentAdd with EditorComment callback', () => {
    let received: EditorComment | null = null;
    const handler = (c: EditorComment) => { received = c; };
    const _props: Pick<MetalDocsEditorProps, 'onCommentAdd'> = { onCommentAdd: handler };
    _props.onCommentAdd?.({ id: 3, author: 'Eve', body: null, resolved: false });
    expect(received).not.toBeNull();
  });
});
