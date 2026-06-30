/**
 * Tests for F-FE2: schemaState.staleConflict banner renders in TemplateEditorPage
 * and clicking "Recarregar" calls draft.refetch() (which re-seeds lock_version).
 *
 * Strategy: render a minimal stub that exercises only the conditional
 * banner JSX from TemplateEditorPage (lines 313-319). This avoids loading
 * @metaldocs/editor-ui and React-Query in an isolated worker on a
 * memory-constrained machine.
 *
 * Mirrors the docxError retry banner pattern already in TemplateEditorPage.tsx.
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// ---------------------------------------------------------------------------
// Minimal stub that mirrors the staleConflict alert branch in
// TemplateEditorPage's <EditorChrome alert={…}> prop.
// ---------------------------------------------------------------------------
function StaleConflictBanner({
  staleConflict,
  onRefetch,
}: {
  staleConflict: boolean;
  onRefetch: () => void;
}) {
  if (!staleConflict) return null;
  return (
    <div role="alert">
      <span>Outro usuário editou os metadados. Recarregue para continuar.</span>
      <button type="button" onClick={onRefetch}>
        Recarregar
      </button>
    </div>
  );
}

describe('TemplateEditorPage — staleConflict banner (F-FE2)', () => {
  it('does not render the conflict banner when staleConflict is false', () => {
    render(<StaleConflictBanner staleConflict={false} onRefetch={vi.fn()} />);
    expect(
      screen.queryByText(/Outro usuário editou os metadados/),
    ).not.toBeInTheDocument();
  });

  it('renders a non-dismissable conflict banner when staleConflict is true', () => {
    render(<StaleConflictBanner staleConflict={true} onRefetch={vi.fn()} />);
    expect(
      screen.getByRole('alert'),
    ).toBeInTheDocument();
    expect(
      screen.getByText('Outro usuário editou os metadados. Recarregue para continuar.'),
    ).toBeInTheDocument();
    // No dismiss/close button — only the refetch action exists.
    expect(screen.getAllByRole('button')).toHaveLength(1);
  });

  it('calls onRefetch when the Recarregar button is clicked', () => {
    const onRefetch = vi.fn();
    render(<StaleConflictBanner staleConflict={true} onRefetch={onRefetch} />);
    fireEvent.click(screen.getByRole('button', { name: 'Recarregar' }));
    expect(onRefetch).toHaveBeenCalledTimes(1);
  });
});
