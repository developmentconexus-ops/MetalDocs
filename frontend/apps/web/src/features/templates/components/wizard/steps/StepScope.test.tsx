import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { StepScope } from './StepScope';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { useAuthStore } from '../../../../../store/auth.store';

function makeProfile(overrides: Partial<DocumentProfile> = {}): DocumentProfile {
  return {
    code: 'POP',
    familyCode: 'FAM',
    name: 'Procedimento Operacional',
    description: 'Perfil de procedimento',
    reviewIntervalDays: 365,
    defaultTemplateVersionId: null,
    ownerUserId: null,
    editableByRole: 'admin',
    governanceClass: 'controlado',
    hasActiveRoute: true,
    hasActiveTemplateRoute: true,
    archivedAt: null,
    createdAt: '2026-01-01T00:00:00.000Z',
    ...overrides,
  };
}

function renderStep(
  profiles: DocumentProfile[],
  onSelect = vi.fn(),
  selectedCode: string | null = null,
) {
  render(
    <MemoryRouter>
      <StepScope
        profiles={profiles}
        isLoading={false}
        isError={false}
        error={null}
        selectedCode={selectedCode}
        onSelect={onSelect}
        onAdvance={() => {}}
        onCancel={() => {}}
        advanceDisabled={false}
        onRetry={() => {}}
      />
    </MemoryRouter>,
  );
  return onSelect;
}

function setCapabilities(capabilities: string[]) {
  useAuthStore.setState({
    user: {
      userId: 'user-1',
      capabilities,
    } as never,
  });
}

afterEach(() => {
  useAuthStore.setState({ user: null });
});

describe('StepScope — generic scope removed (ADR 0086)', () => {
  it('offers no scope-type choice at all', () => {
    renderStep([makeProfile()]);

    expect(screen.queryByText('Genérico')).toBeNull();
    expect(screen.queryByText('Para um perfil')).toBeNull();
    expect(screen.queryByRole('radio', { name: /Genérico/ })).toBeNull();
  });

  it('shows the profile picker immediately, with no scope gate in front of it', () => {
    renderStep([makeProfile()]);

    expect(screen.getByRole('heading', { name: 'Perfil do template' })).toBeInTheDocument();
    expect(screen.getByRole('radio', { name: /POP/ })).toBeInTheDocument();
  });

  it('requires a profile before advancing', () => {
    renderStep([makeProfile()]);

    expect(screen.getByText(/Selecione um perfil para continuar/)).toBeInTheDocument();
  });
});

describe('StepScope — template approval-route readiness gate (ADR 0086)', () => {
  it('disables a profile that has no active TEMPLATE approval route', () => {
    renderStep([makeProfile({ hasActiveTemplateRoute: false })]);

    const card = screen.getByRole('radio', { name: /POP/ });
    expect(card).toBeDisabled();
    expect(card).toHaveAttribute('title', 'Sem rota de aprovação de template ativa');
  });

  it('shows the explanatory marker on an ineligible profile', () => {
    renderStep([makeProfile({ hasActiveTemplateRoute: false })]);

    expect(
      screen.getByText('Incompleto — sem rota de aprovação de template'),
    ).toBeInTheDocument();
  });

  it('never calls onSelect for an ineligible profile', () => {
    const onSelect = renderStep([makeProfile({ hasActiveTemplateRoute: false })]);

    fireEvent.click(screen.getByRole('radio', { name: /POP/ }));

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('gates on the TEMPLATE route, not the document route', () => {
    // A profile ready for documents but not for templates must still be
    // blocked here — the two readiness flags are orthogonal.
    renderStep([makeProfile({ hasActiveRoute: true, hasActiveTemplateRoute: false })]);

    expect(screen.getByRole('radio', { name: /POP/ })).toBeDisabled();
  });

  it('leaves an eligible profile selectable', () => {
    const onSelect = renderStep([makeProfile({ hasActiveTemplateRoute: true })]);

    const card = screen.getByRole('radio', { name: /POP/ });
    expect(card).not.toBeDisabled();

    fireEvent.click(card);
    expect(onSelect).toHaveBeenCalledWith('POP');
  });

  it('offers "Configurar rota" only when the user holds route.manage', () => {
    setCapabilities(['route.manage']);
    renderStep([makeProfile({ hasActiveTemplateRoute: false })]);

    expect(screen.getByRole('link', { name: 'Configurar rota' })).toHaveAttribute(
      'href',
      '/approval-routes',
    );
  });

  it('shows the marker without a remediation link when route.manage is absent', () => {
    setCapabilities(['templates.create']);
    renderStep([makeProfile({ hasActiveTemplateRoute: false })]);

    expect(
      screen.getByText('Incompleto — sem rota de aprovação de template'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Configurar rota' })).toBeNull();
  });

  it('renders eligible and ineligible profiles side by side', () => {
    renderStep([
      makeProfile({ code: 'POP', name: 'Procedimento', hasActiveTemplateRoute: true }),
      makeProfile({ code: 'INS', name: 'Instrução', hasActiveTemplateRoute: false }),
    ]);

    expect(screen.getByRole('radio', { name: /POP/ })).not.toBeDisabled();
    expect(screen.getByRole('radio', { name: /INS/ })).toBeDisabled();
  });
});

// The roving-radio group moves selection by index, so Arrow/Home/End could
// otherwise land on a profile the click path refuses — the reducer unlocks the
// remaining steps on any non-null profileCode and submit would eat a guaranteed
// 409 APPROVAL_ROUTE_MISSING. Both input paths share one eligibility predicate.
describe('StepScope — readiness gate is fail-closed on the keyboard path too', () => {
  const ELIGIBLE = makeProfile({ code: 'POP', name: 'Procedimento', hasActiveTemplateRoute: true });
  const BLOCKED = makeProfile({ code: 'INS', name: 'Instrução', hasActiveTemplateRoute: false });

  it('does not select a blocked profile via ArrowRight', () => {
    const onSelect = renderStep([ELIGIBLE, BLOCKED], vi.fn(), 'POP');

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'ArrowRight' });

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('does not select a blocked profile via ArrowDown', () => {
    const onSelect = renderStep([ELIGIBLE, BLOCKED], vi.fn(), 'POP');

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'ArrowDown' });

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('does not select a blocked profile via End', () => {
    const onSelect = renderStep([ELIGIBLE, BLOCKED], vi.fn(), 'POP');

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'End' });

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('does not select a blocked profile via Home', () => {
    // Blocked profile sits first, so Home targets index 0 = the blocked card.
    const onSelect = renderStep([BLOCKED, ELIGIBLE], vi.fn(), 'POP');

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'Home' });

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('does not select the "em breve" profile via the keyboard either', () => {
    const onSelect = renderStep(
      [ELIGIBLE, makeProfile({ code: 'CHK', name: 'Checklist', hasActiveTemplateRoute: true })],
      vi.fn(),
      'POP',
    );

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'ArrowRight' });

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('still selects an eligible profile via the keyboard', () => {
    const other = makeProfile({ code: 'MAN', name: 'Manual', hasActiveTemplateRoute: true });
    const onSelect = renderStep([ELIGIBLE, other], vi.fn(), 'POP');

    fireEvent.keyDown(screen.getByRole('radio', { name: /POP/ }), { key: 'ArrowRight' });

    expect(onSelect).toHaveBeenCalledWith('MAN');
  });
});
