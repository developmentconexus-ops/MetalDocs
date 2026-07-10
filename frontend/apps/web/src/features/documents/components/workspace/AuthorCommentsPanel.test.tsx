import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { EditorComment } from '@metaldocs/editor-ui';
import { AuthorCommentsPanel } from './AuthorCommentsPanel';

const txt = (t: string) => [{ type: 'paragraph', content: [{ type: 'text', text: t }] }];
const root: EditorComment = { id: 1, author: 'Revisor', body: txt('Ajuste a introdução'), resolved: false };
const reply: EditorComment = { id: 2, parentId: 1, author: 'Autor', body: txt('Feito'), resolved: false };

function renderPanel(comments: EditorComment[], over: Partial<Record<'onReply' | 'onResolve' | 'onReopen', ReturnType<typeof vi.fn>>> = {}) {
  const onReply = over.onReply ?? vi.fn();
  const onResolve = over.onResolve ?? vi.fn();
  const onReopen = over.onReopen ?? vi.fn();
  render(<AuthorCommentsPanel comments={comments} onReply={onReply} onResolve={onResolve} onReopen={onReopen} />);
  return { onReply, onResolve, onReopen };
}

describe('AuthorCommentsPanel', () => {
  it('renders the reviewer thread text and its reply', () => {
    renderPanel([root, reply]);
    expect(screen.getByText('Ajuste a introdução')).toBeInTheDocument();
    expect(screen.getByText('Feito')).toBeInTheDocument();
    expect(screen.getByText('Revisor')).toBeInTheDocument();
  });

  it('reply composer submits text against the root comment then clears', () => {
    const { onReply } = renderPanel([root]);
    const box = screen.getByLabelText(/responder ao comentário/i);
    fireEvent.change(box, { target: { value: 'Vou corrigir' } });
    fireEvent.click(screen.getByRole('button', { name: /responder/i }));
    expect(onReply).toHaveBeenCalledWith('Vou corrigir', root);
    expect((box as HTMLTextAreaElement).value).toBe('');
  });

  it('does not submit an empty/whitespace reply', () => {
    const { onReply } = renderPanel([root]);
    fireEvent.change(screen.getByLabelText(/responder ao comentário/i), { target: { value: '   ' } });
    fireEvent.click(screen.getByRole('button', { name: /responder/i }));
    expect(onReply).not.toHaveBeenCalled();
  });

  it('resolve button calls onResolve for an unresolved thread', () => {
    const { onResolve } = renderPanel([root]);
    fireEvent.click(screen.getByRole('button', { name: /resolver/i }));
    expect(onResolve).toHaveBeenCalledWith(root);
  });

  it('resolved thread shows reopen, calls onReopen with the resolved comment', () => {
    const { onReopen } = renderPanel([{ ...root, resolved: true }]);
    fireEvent.click(screen.getByRole('button', { name: /reabrir/i }));
    expect(onReopen).toHaveBeenCalledWith({ ...root, resolved: true });
  });

  it('renders an empty state when there are no comments', () => {
    renderPanel([]);
    expect(screen.getByText(/nenhum comentário/i)).toBeInTheDocument();
  });
});
