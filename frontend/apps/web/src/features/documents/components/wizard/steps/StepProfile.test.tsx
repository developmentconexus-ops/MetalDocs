import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { StepProfile } from './StepProfile';
import type { DocumentProfile } from '../../../../taxonomy/types';
import { useAuthStore } from '../../../../../store/auth.store';

function makeProfile(overrides: Partial<DocumentProfile> = {}): DocumentProfile {
  return {
    code: 'PRC',
    familyCode: 'FAM',
    name: 'Procedimento',
    description: 'Perfil de procedimento',
    reviewIntervalDays: 365,
    defaultTemplateVersionId: null,
    ownerUserId: null,
    editableByRole: 'admin',
    governanceClass: 'controlado',
    hasActiveRoute: true,
    archivedAt: null,
    createdAt: '2026-01-01T00:00:00.000Z',
    ...overrides,
  };
}

function renderStep(profiles: DocumentProfile[], onSelect = vi.fn()) {
  render(
    <MemoryRouter>
      <StepProfile
        profiles={profiles}
        isLoading={false}
        isError={false}
        error={null}
        selectedCode={null}
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

describe('StepProfile — approval-route readiness gate (D1)', () => {
  it('disables a profile that has no active approval route', () => {
    renderStep([makeProfile({ code: 'PRC', hasActiveRoute: false })]);

    const card = screen.getByRole('radio', { name: /PRC/ });
    expect(card).toBeDisabled();
    expect(card).toHaveAttribute('title', 'Sem rota de aprovação ativa');
  });

  it('shows the explanatory marker on an ineligible profile', () => {
    renderStep([makeProfile({ hasActiveRoute: false })]);

    expect(screen.getByText('Incompleto — sem rota de aprovação')).toBeInTheDocument();
  });

  it('never calls onSelect for an ineligible profile', () => {
    const onSelect = renderStep([makeProfile({ hasActiveRoute: false })]);

    fireEvent.click(screen.getByRole('radio', { name: /PRC/ }));

    expect(onSelect).not.toHaveBeenCalled();
  });

  it('leaves an eligible profile selectable', () => {
    const onSelect = renderStep([makeProfile({ hasActiveRoute: true })]);

    const card = screen.getByRole('radio', { name: /PRC/ });
    expect(card).not.toBeDisabled();

    fireEvent.click(card);
    expect(onSelect).toHaveBeenCalledWith('PRC');
  });

  it('offers "Configurar rota" only when the user holds route.manage', () => {
    setCapabilities(['route.manage']);
    renderStep([makeProfile({ hasActiveRoute: false })]);

    const link = screen.getByRole('link', { name: 'Configurar rota' });
    expect(link).toHaveAttribute('href', '/approval-routes');
  });

  it('shows the marker without a remediation link when route.manage is absent', () => {
    setCapabilities(['controlled_documents.create']);
    renderStep([makeProfile({ hasActiveRoute: false })]);

    expect(screen.getByText('Incompleto — sem rota de aprovação')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Configurar rota' })).toBeNull();
  });

  it('renders eligible and ineligible profiles side by side', () => {
    renderStep([
      makeProfile({ code: 'PRC', name: 'Procedimento', hasActiveRoute: true }),
      makeProfile({ code: 'INS', name: 'Instrução', hasActiveRoute: false }),
    ]);

    expect(screen.getByRole('radio', { name: /PRC/ })).not.toBeDisabled();
    expect(screen.getByRole('radio', { name: /INS/ })).toBeDisabled();
  });
});
