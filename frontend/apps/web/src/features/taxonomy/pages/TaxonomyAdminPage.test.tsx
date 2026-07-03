import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';

import { TaxonomyAdminPage } from './TaxonomyAdminPage';

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
    React.createElement(QueryClientProvider, { client: qc }, children);
}

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
    fireEvent.click(screen.getByRole('button', { name: 'Perfis' }));

    await waitFor(() => expect(screen.getByText('Nenhum perfil encontrado.')).toBeTruthy());
  });

  it('switches to the areas tab and shows the empty state', async () => {
    render(React.createElement(TaxonomyAdminPage), { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('Standard Operating Procedure')).toBeTruthy());
    fireEvent.click(screen.getByRole('button', { name: 'Áreas' }));

    await waitFor(() => expect(screen.getByText('Nenhuma área encontrada.')).toBeTruthy());
  });
});
