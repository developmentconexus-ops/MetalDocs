import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { TaxonomyAdminPage } from './TaxonomyAdminPage';
import * as taxonomyApi from '../api/taxonomy';
import type { DocumentProfile } from '../types';
import { useAuthStore } from '../../../store/auth.store';

vi.mock('../api/taxonomy', () => ({
  fetchFamilies: vi.fn().mockResolvedValue([
    { code: 'sop', name: 'Standard Operating Procedure', description: '', isActive: true, createdAt: '' },
  ]),
  fetchProfiles: vi.fn().mockResolvedValue([]),
  fetchAreas: vi.fn().mockResolvedValue([]),
  createFamily: vi.fn(),
  updateFamily: vi.fn(),
  deactivateFamily: vi.fn(),
  createProfile: vi.fn(),
  updateProfile: vi.fn(),
  archiveProfile: vi.fn(),
  setDefaultTemplate: vi.fn(),
  createArea: vi.fn(),
  updateArea: vi.fn(),
  archiveArea: vi.fn(),
}));

function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      QueryClientProvider,
      { client: qc },
      React.createElement(MemoryRouter, null, children),
    );
}

function makeProfile(overrides: Partial<DocumentProfile> = {}): DocumentProfile {
  return {
    code: 'PRC',
    familyCode: 'sop',
    name: 'Procedimento',
    description: '',
    reviewIntervalDays: 365,
    defaultTemplateVersionId: null,
    ownerUserId: null,
    editableByRole: 'admin',
    governanceClass: 'controlado',
    hasActiveRoute: true,
    hasActiveTemplateRoute: true,
    archivedAt: null,
    createdAt: '',
    ...overrides,
  };
}

async function openProfilesTab() {
  render(React.createElement(TaxonomyAdminPage), { wrapper: wrap() });
  await waitFor(() => expect(screen.getByText('Standard Operating Procedure')).toBeTruthy());
  fireEvent.click(screen.getByRole('tab', { name: 'Perfis' }));
}

afterEach(() => {
  useAuthStore.setState({ user: null });
  // Module-level mockResolvedValue persists across tests in this file.
  vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([]);
});

describe('TaxonomyAdminPage', () => {
  it('renders the families tab with fetched data by default', async () => {
    render(React.createElement(TaxonomyAdminPage), { wrapper: wrap() });

    expect(screen.getByText('Tipos Documentais')).toBeTruthy();
    await waitFor(() => expect(screen.getByText('Standard Operating Procedure')).toBeTruthy());
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('switches to the profiles tab and shows the empty state', async () => {
    render(React.createElement(TaxonomyAdminPage), { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('Standard Operating Procedure')).toBeTruthy());
    fireEvent.click(screen.getByRole('tab', { name: 'Perfis' }));

    await waitFor(() => expect(screen.getByText('Nenhum perfil encontrado.')).toBeTruthy());
  });

  it('badges an active profile that has no active approval route as INCOMPLETO', async () => {
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({ code: 'PRC', name: 'Procedimento', hasActiveRoute: false }),
    ]);

    await openProfilesTab();

    const badge = await screen.findByText('INCOMPLETO');
    expect(badge).toHaveAttribute('title', 'Sem rota de aprovação ativa');
    // Orthogonal: the template-route badge must NOT appear just because the
    // document route is missing (ADR 0086).
    expect(screen.queryByText('SEM ROTA DE TEMPLATE')).toBeNull();
  });

  it('badges an active profile that has no active TEMPLATE route, independently of the document route', async () => {
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({
        code: 'PRC',
        name: 'Procedimento',
        hasActiveRoute: true,
        hasActiveTemplateRoute: false,
      }),
    ]);

    await openProfilesTab();

    const badge = await screen.findByText('SEM ROTA DE TEMPLATE');
    expect(badge).toHaveAttribute('title', 'Sem rota de aprovação de template ativa');
    expect(screen.queryByText('INCOMPLETO')).toBeNull();
  });

  it('badges both readiness gaps when neither route kind is active', async () => {
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({
        code: 'PRC',
        name: 'Procedimento',
        hasActiveRoute: false,
        hasActiveTemplateRoute: false,
      }),
    ]);

    await openProfilesTab();

    await screen.findByText('INCOMPLETO');
    expect(screen.getByText('SEM ROTA DE TEMPLATE')).toBeTruthy();
  });

  it('does not badge a profile whose approval routes are both active', async () => {
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({ code: 'PRC', hasActiveRoute: true, hasActiveTemplateRoute: true }),
    ]);

    await openProfilesTab();

    await waitFor(() => expect(screen.getByText('Procedimento')).toBeTruthy());
    expect(screen.queryByText('INCOMPLETO')).toBeNull();
    expect(screen.queryByText('SEM ROTA DE TEMPLATE')).toBeNull();
  });

  it('offers the route shortcut only to users holding route.manage', async () => {
    useAuthStore.setState({
      user: { userId: 'user-1', capabilities: ['route.manage'] } as never,
    });
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({ hasActiveRoute: false }),
    ]);

    await openProfilesTab();

    const link = await screen.findByRole('link', { name: 'Configurar rota' });
    expect(link).toHaveAttribute('href', '/approval-routes');
  });

  it('hides the route shortcut without route.manage', async () => {
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValue([
      makeProfile({ hasActiveRoute: false }),
    ]);

    await openProfilesTab();

    await screen.findByText('INCOMPLETO');
    expect(screen.queryByRole('link', { name: 'Configurar rota' })).toBeNull();
  });

  it('switches to the areas tab and shows the empty state', async () => {
    render(React.createElement(TaxonomyAdminPage), { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('Standard Operating Procedure')).toBeTruthy());
    fireEvent.click(screen.getByRole('tab', { name: 'Áreas' }));

    await waitFor(() => expect(screen.getByText('Nenhuma área encontrada.')).toBeTruthy());
  });
});
