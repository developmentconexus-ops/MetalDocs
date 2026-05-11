import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { EditorChrome } from './EditorChrome';
import { AutosaveStatus, type AutosaveState } from './parts/AutosaveStatus';
import { VersionBadge } from './parts/VersionBadge';

// --- EditorChrome slot truthy-collapse ---

describe('EditorChrome slot rendering', () => {
  it('renders children unconditionally', () => {
    render(<EditorChrome><span data-testid="child" /></EditorChrome>);
    expect(screen.getByTestId('child')).toBeTruthy();
  });

  it('omits left slot when not provided', () => {
    const { container } = render(
      <EditorChrome center={<span>center</span>}><div /></EditorChrome>
    );
    // overlayLeft div is not rendered when left prop is falsy
    expect(container.querySelector('[class*="overlayLeft"]')).toBeNull();
  });

  it('renders left slot when provided', () => {
    render(
      <EditorChrome left={<button>Back</button>}><div /></EditorChrome>
    );
    expect(screen.getByRole('button', { name: 'Back' })).toBeTruthy();
  });

  it('omits center slot when not provided', () => {
    const { container } = render(<EditorChrome><div /></EditorChrome>);
    expect(container.querySelector('[class*="overlayCenter"]')).toBeNull();
  });

  it('renders center slot when provided', () => {
    render(
      <EditorChrome center={<span>Title</span>}><div /></EditorChrome>
    );
    expect(screen.getByText('Title')).toBeTruthy();
  });

  it('omits right slot when not provided', () => {
    const { container } = render(<EditorChrome><div /></EditorChrome>);
    expect(container.querySelector('[class*="overlayRight"]')).toBeNull();
  });

  it('renders right slot when provided', () => {
    render(
      <EditorChrome right={<span>Actions</span>}><div /></EditorChrome>
    );
    expect(screen.getByText('Actions')).toBeTruthy();
  });

  it('omits alert slot when not provided', () => {
    const { container } = render(<EditorChrome><div /></EditorChrome>);
    expect(container.querySelector('[class*="overlayAlert"]')).toBeNull();
  });

  it('renders alert slot when provided', () => {
    render(
      <EditorChrome alert={<span role="alert">Error!</span>}><div /></EditorChrome>
    );
    expect(screen.getByRole('alert')).toBeTruthy();
  });

  it('appends className to wrapper', () => {
    const { container } = render(
      <EditorChrome className="extra"><div /></EditorChrome>
    );
    expect(container.firstElementChild?.className).toContain('extra');
  });
});

// --- AutosaveStatus — 7-state labels + aria-live ---

const CASES: Array<{ status: AutosaveState; label: string; assertive: boolean }> = [
  { status: 'idle',         label: 'Salvo',                  assertive: false },
  { status: 'dirty',        label: 'Editado',                assertive: false },
  { status: 'saving',       label: 'Salvando…',              assertive: false },
  { status: 'saved',        label: 'Salvo',                  assertive: false },
  { status: 'stale',        label: 'Atualização disponível', assertive: false },
  { status: 'session_lost', label: 'Sessão perdida',         assertive: true  },
  { status: 'error',        label: 'Erro ao salvar',         assertive: true  },
];

describe('AutosaveStatus', () => {
  for (const { status, label, assertive } of CASES) {
    it(`renders label for "${status}"`, () => {
      render(<AutosaveStatus status={status} />);
      expect(screen.getByRole('status').textContent).toContain(label);
    });

    it(`uses aria-live="${assertive ? 'assertive' : 'polite'}" for "${status}"`, () => {
      render(<AutosaveStatus status={status} />);
      expect(screen.getByRole('status').getAttribute('aria-live')).toBe(
        assertive ? 'assertive' : 'polite'
      );
    });
  }

  it('accepts label overrides', () => {
    render(<AutosaveStatus status="saving" labels={{ saving: 'Saving…' }} />);
    expect(screen.getByRole('status').textContent).toContain('Saving…');
  });

  it('merges custom className onto wrapper span', () => {
    render(<AutosaveStatus status="idle" className="extra" />);
    expect(screen.getByRole('status').className).toContain('extra');
  });
});

// --- VersionBadge passthrough ---

describe('VersionBadge', () => {
  it('renders children inside a span', () => {
    render(<VersionBadge>v5</VersionBadge>);
    expect(screen.getByText('v5').tagName).toBe('SPAN');
  });

  it('appends extra className', () => {
    render(<VersionBadge className="custom">REV05</VersionBadge>);
    expect(screen.getByText('REV05').className).toContain('custom');
  });
});
