import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import type { TrackedChange } from '@metaldocs/editor-ui';
import type { EditorComment } from '@metaldocs/editor-ui';
import { RequestedChangesPanel } from '../RequestedChangesPanel';

function makeTrackedChange(overrides: Partial<TrackedChange> = {}): TrackedChange {
  return {
    revisionId: 'rev-1',
    author: 'Maria Autora',
    type: 'insertion',
    excerpt: 'texto inserido de exemplo',
    ...overrides,
  };
}

function makeComment(overrides: Partial<EditorComment> = {}): EditorComment {
  return {
    id: 42,
    author: 'João Revisor',
    body: [],
    resolved: false,
    ...overrides,
  };
}

describe('RequestedChangesPanel', () => {
  it('lists tracked changes with author/type/excerpt', () => {
    render(
      <RequestedChangesPanel
        trackedChanges={[makeTrackedChange()]}
        comments={[]}
        onAcceptChange={vi.fn()}
        onRejectChange={vi.fn()}
        onResolveComment={vi.fn()}
      />,
    );

    expect(screen.getByText('Maria Autora')).toBeTruthy();
    expect(screen.getByText('texto inserido de exemplo')).toBeTruthy();
  });

  it('lists unresolved comments only', () => {
    render(
      <RequestedChangesPanel
        trackedChanges={[]}
        comments={[
          makeComment({ id: 1, author: 'Aberto', resolved: false }),
          makeComment({ id: 2, author: 'Fechado', resolved: true }),
        ]}
        onAcceptChange={vi.fn()}
        onRejectChange={vi.fn()}
        onResolveComment={vi.fn()}
      />,
    );

    expect(screen.getByText('Aberto')).toBeTruthy();
    expect(screen.queryByText('Fechado')).toBeNull();
  });

  it('calls onAcceptChange with the revisionId when Aceitar is clicked', () => {
    const onAcceptChange = vi.fn();
    render(
      <RequestedChangesPanel
        trackedChanges={[makeTrackedChange({ revisionId: 'rev-abc' })]}
        comments={[]}
        onAcceptChange={onAcceptChange}
        onRejectChange={vi.fn()}
        onResolveComment={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: /Aceitar/i }));
    expect(onAcceptChange).toHaveBeenCalledWith('rev-abc');
  });

  it('calls onRejectChange with the revisionId when Rejeitar is clicked', () => {
    const onRejectChange = vi.fn();
    render(
      <RequestedChangesPanel
        trackedChanges={[makeTrackedChange({ revisionId: 'rev-xyz' })]}
        comments={[]}
        onAcceptChange={vi.fn()}
        onRejectChange={onRejectChange}
        onResolveComment={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: /Rejeitar/i }));
    expect(onRejectChange).toHaveBeenCalledWith('rev-xyz');
  });

  it('calls onResolveComment with the comment when Resolver is clicked', () => {
    const onResolveComment = vi.fn();
    const comment = makeComment({ id: 7, resolved: false });
    render(
      <RequestedChangesPanel
        trackedChanges={[]}
        comments={[comment]}
        onAcceptChange={vi.fn()}
        onRejectChange={vi.fn()}
        onResolveComment={onResolveComment}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: /Resolver/i }));
    expect(onResolveComment).toHaveBeenCalledWith(comment);
  });

  it('shows the all-clean empty state when there are no tracked changes or unresolved comments', () => {
    render(
      <RequestedChangesPanel
        trackedChanges={[]}
        comments={[]}
        onAcceptChange={vi.fn()}
        onRejectChange={vi.fn()}
        onResolveComment={vi.fn()}
      />,
    );

    expect(screen.getByText(/Nenhuma marcação pendente/i)).toBeTruthy();
  });
});
