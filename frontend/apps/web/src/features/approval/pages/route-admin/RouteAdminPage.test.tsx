import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ApprovalError } from '../../api/mutationClient';
import * as routeAdminApi from '../../api/routeAdminApi';
import type { ListRoutesResponse, RouteSummary } from '../../api/routeAdminApi';
import { RouteAdminPage } from './RouteAdminPage';

vi.mock('../../api/routeAdminApi');
vi.mock('../../../taxonomy/api/taxonomy', () => ({
  fetchAreas: vi.fn().mockResolvedValue([{ code: 'AREA-01', name: 'Área 01' }]),
}));

function makeRoute(overrides: Partial<RouteSummary> = {}): RouteSummary {
  return {
    id: 'route-1',
    name: 'Rota Jurídica',
    tenant_id: 'tenant-1',
    profile_code: 'JUR',
    active: true,
    version: 3,
    stages: [
      {
        label: 'Revisão',
        required_role: 'approver',
        required_capability: 'doc.signoff',
        area_code: 'AREA-01',
        quorum_kind: 'any_1_of',
        quorum_m: null,
        drift_policy: 'reduce_quorum',
      },
      {
        label: 'Aprovação',
        required_role: 'approver',
        required_capability: 'doc.signoff',
        area_code: 'AREA-01',
        quorum_kind: 'all_of',
        quorum_m: null,
        drift_policy: 'reduce_quorum',
      },
    ],
    created_at: '2026-04-20T10:00:00.000Z',
    updated_at: '2026-04-20T10:00:00.000Z',
    ...overrides,
  };
}

function renderWithProviders(ui: ReactElement) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
      mutations: { retry: false },
    },
  });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

function listResponse(routes: RouteSummary[]): ListRoutesResponse {
  return { routes, total: routes.length };
}

async function flushAsync() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe('RouteAdminPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders rows fetched from useRoutesQuery', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([makeRoute()]));

    renderWithProviders(<RouteAdminPage />);

    expect(await screen.findByText('Rota Jurídica')).toBeTruthy();
    expect(screen.getByText('2 etapa(s)')).toBeTruthy();
    expect(screen.getByLabelText('Status: ativa')).toBeTruthy();
  });

  it('opens the editor dialog when "Nova rota" is clicked', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);

    await screen.findByRole('heading', { name: 'Administração de Rotas' });
    fireEvent.click(screen.getByRole('button', { name: 'Nova rota' }));

    expect(await screen.findByRole('dialog', { name: 'Criar rota' })).toBeTruthy();
  });

  it('disables edit for inactive routes with cause-based tooltip copy', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(
      listResponse([makeRoute({ id: 'route-2', name: 'Rota Inativa', active: false })]),
    );

    renderWithProviders(<RouteAdminPage />);

    const editButton = await screen.findByRole('button', { name: 'Editar Rota Inativa' });
    expect((editButton as HTMLButtonElement).disabled).toBe(true);
    expect(editButton.getAttribute('title')).toBe(
      'Rota inativa — desativada e somente leitura.',
    );
  });

  it('cascades validation: missing stage label fires before role/area errors', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Nova Rota' },
    });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() =>
      expect(within(dialog).getByText('Toda etapa deve ter nome.')).toBeTruthy(),
    );

    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 1'), {
      target: { value: 'Jurídico' },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() =>
      expect(
        within(dialog).getByText('A etapa "Jurídico" deve ter uma role definida.'),
      ).toBeTruthy(),
    );

    expect(vi.mocked(routeAdminApi.createRoute)).not.toHaveBeenCalled();
  });

  it('m_of_n with M=0 shows distinct error and does not call createRoute', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota M' },
    });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'OPS' },
    });
    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 1'), {
      target: { value: 'Operação' },
    });
    await within(dialog).findByRole('option', { name: 'Aprovador' });
    await within(dialog).findByRole('option', { name: 'Área 01 (AREA-01)' });
    fireEvent.change(within(dialog).getByLabelText('Role requerida da etapa 1'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área da etapa 1'), {
      target: { value: 'AREA-01' },
    });
    fireEvent.change(within(dialog).getByLabelText('Quórum da etapa 1'), {
      target: { value: 'm_of_n' },
    });
    const mInput = within(dialog).getByLabelText('M da etapa 1') as HTMLInputElement;
    fireEvent.change(mInput, { target: { value: '0' } });

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText('Na etapa "Operação", informe um valor de M válido.'),
      ).toBeTruthy(),
    );
    expect(vi.mocked(routeAdminApi.createRoute)).not.toHaveBeenCalled();
  });

  it('create surfaces backend conflict via shared problem-code → PT-BR mapping', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));
    // `route.duplicate_profile` is the real backend code (errors.go:36); the
    // canonical PT-BR copy lives in `lib/api/errorMessages.ts`. The dialog
    // stays open with the mapped message inline; no synthetic optimistic row
    // is ever inserted, so there is nothing in the table to roll back.
    vi.mocked(routeAdminApi.createRoute).mockRejectedValueOnce(
      new ApprovalError('route.duplicate_profile', 409, 'duplicate profile'),
    );

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Conflito' },
    });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });
    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 1'), {
      target: { value: 'Jurídico' },
    });
    await within(dialog).findByRole('option', { name: 'Aprovador' });
    await within(dialog).findByRole('option', { name: 'Área 01 (AREA-01)' });
    fireEvent.change(within(dialog).getByLabelText('Role requerida da etapa 1'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área da etapa 1'), {
      target: { value: 'AREA-01' },
    });

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText(/Já existe uma rota para este perfil\./),
      ).toBeTruthy(),
    );
    // No synthetic row was inserted into the list — assert the table stays empty.
    expect(
      screen.queryByRole('cell', { name: 'Conflito' }),
    ).toBeNull();
  });

  it('deactivate disables submit when reason is too short', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([makeRoute()]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Desativar Rota Jurídica' }));

    const dialog = await screen.findByRole('dialog', { name: 'Confirmar desativação' });
    const submit = within(dialog).getByRole('button', { name: /Confirmar desativação/i });
    expect((submit as HTMLButtonElement).disabled).toBe(true);

    fireEvent.change(within(dialog).getByLabelText('Motivo da desativação'), {
      target: { value: 'ok' },
    });
    expect((submit as HTMLButtonElement).disabled).toBe(true);

    fireEvent.change(within(dialog).getByLabelText('Motivo da desativação'), {
      target: { value: 'erro recorrente' },
    });
    expect((submit as HTMLButtonElement).disabled).toBe(false);
  });

  it('Escape on the dialog closes the editor', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    // Native <dialog> closes on `cancel` event (fired by Esc). Dispatch it directly
    // so the test stays jsdom-friendly without depending on key event propagation
    // through showModal().
    await act(async () => {
      dialog.dispatchEvent(new Event('cancel'));
      await flushAsync();
    });

    await waitFor(() =>
      expect(screen.queryByRole('dialog', { name: 'Criar rota' })).toBeNull(),
    );
  });

  it('role select options come from the IAM roles query, not a frozen import', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    // Labels come from the labelled IAM catalogue (PT-BR), not raw role codes
    // like the legacy frozen `STAGE_ROLES` array used.
    expect(await within(dialog).findByRole('option', { name: 'Aprovador' })).toBeTruthy();
    expect(within(dialog).getByRole('option', { name: 'Editor' })).toBeTruthy();
    expect(within(dialog).getByRole('option', { name: 'Administrador do sistema' })).toBeTruthy();
  });
});
