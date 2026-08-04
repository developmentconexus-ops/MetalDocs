import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ApprovalError } from '../../api/mutationClient';
import * as routeAdminApi from '../../api/routeAdminApi';
import type { ListRoutesResponse, RouteSummary } from '../../api/routeAdminApi';
import * as taxonomyApi from '../../../taxonomy/api/taxonomy';
import type { DocumentProfile } from '../../../taxonomy/types';
import { stageOrderViolationMessage } from './routeGovernance';
import { RouteAdminPage } from './RouteAdminPage';

vi.mock('../../api/routeAdminApi');
vi.mock('../../../taxonomy/api/taxonomy', () => ({
  fetchAreas: vi.fn().mockResolvedValue([{ code: 'AREA-01', name: 'Área 01' }]),
  fetchProfiles: vi.fn().mockResolvedValue([
    { code: 'JUR', name: 'Jurídico' },
    { code: 'OPS', name: 'Operações' },
  ]),
}));
vi.mock('../../../iam/queries/useUsersQuery', () => ({
  useUsersQuery: vi.fn(() => ({
    data: { items: [], page: { next_cursor: null, has_more: false } },
    isLoading: false,
    isError: false,
  })),
}));

function makeProfile(overrides: Partial<DocumentProfile> = {}): DocumentProfile {
  return {
    code: 'CTL',
    familyCode: 'FAM',
    name: 'Controlado',
    description: '',
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
        order: 1,
        name: 'Revisão',
        required_capability: 'document.signoff',
        selectors: [{ kind: 'role_in_fixed_area', role: 'approver', area_code: 'AREA-01' }],
        quorum: 'any_1_of',
        quorum_m: null,
        drift_policy: 'reduce_quorum',
      },
      {
        order: 2,
        name: 'Aprovação',
        required_capability: 'document.signoff',
        selectors: [{ kind: 'role_in_fixed_area', role: 'approver', area_code: 'AREA-01' }],
        quorum: 'all_of',
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
    expect(screen.getByText('2 etapas')).toBeTruthy();
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
    await within(dialog).findByRole('option', { name: 'Jurídico (JUR)' });
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
        within(dialog).getByText('Na etapa "Jurídico", selecione uma role.'),
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
    await within(dialog).findByRole('option', { name: 'Operações (OPS)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'OPS' },
    });
    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 1'), {
      target: { value: 'Operação' },
    });
    await within(dialog).findByRole('option', { name: 'Aprovador' });
    await within(dialog).findByRole('option', { name: 'Área 01 (AREA-01)' });
    fireEvent.change(within(dialog).getByLabelText('Role do seletor 1 da etapa 1'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área do seletor 1 da etapa 1'), {
      target: { value: 'AREA-01' },
    });
    fireEvent.click(within(dialog).getByRole('radio', { name: 'M de N' }));
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
      new ApprovalError('conflict.approval_route_duplicate_profile', 409, 'duplicate profile'),
    );

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Conflito' },
    });
    await within(dialog).findByRole('option', { name: 'Jurídico (JUR)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });
    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 1'), {
      target: { value: 'Jurídico' },
    });
    await within(dialog).findByRole('option', { name: 'Aprovador' });
    await within(dialog).findByRole('option', { name: 'Área 01 (AREA-01)' });
    fireEvent.change(within(dialog).getByLabelText('Role do seletor 1 da etapa 1'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área do seletor 1 da etapa 1'), {
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

  async function fillStage1Basics(dialog: HTMLElement, label: string) {
    fireEvent.change(within(dialog).getByLabelText(`Nome da etapa 1`), {
      target: { value: label },
    });
    await within(dialog).findByRole('option', { name: 'Aprovador' });
    await within(dialog).findByRole('option', { name: 'Área 01 (AREA-01)' });
    fireEvent.change(within(dialog).getByLabelText('Role do seletor 1 da etapa 1'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área do seletor 1 da etapa 1'), {
      target: { value: 'AREA-01' },
    });
  }

  it('selecting a controlado profile shows the "obrigatório" badge and blocks save without an approval stage', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValueOnce([
      makeProfile({ code: 'CTL', name: 'Controlado', governanceClass: 'controlado' }),
    ]);

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Controlada' },
    });
    await within(dialog).findByRole('option', { name: 'Controlado (CTL)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'CTL' },
    });

    expect(await within(dialog).findByText('Obrigatório ≥1 assinatura')).toBeTruthy();

    await fillStage1Basics(dialog, 'Revisão inicial');
    // Only a review stage is present (no approval stage) — controlado requires one.
    fireEvent.click(within(dialog).getByRole('radio', { name: 'Revisão' }));

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() =>
      expect(
        within(dialog).getByText(
          'Perfil controlado exige ao menos uma etapa de assinatura (aprovação).',
        ),
      ).toBeTruthy(),
    );
    expect(vi.mocked(routeAdminApi.createRoute)).not.toHaveBeenCalled();
  });

  it('selecting a simples profile shows the "opcional" badge and allows a review-only route', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValueOnce([
      makeProfile({ code: 'SMP', name: 'Simples', governanceClass: 'simples' }),
    ]);

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Simples' },
    });
    await within(dialog).findByRole('option', { name: 'Simples (SMP)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'SMP' },
    });

    expect(await within(dialog).findByText('Assinatura opcional')).toBeTruthy();

    await fillStage1Basics(dialog, 'Conferência');
    fireEvent.click(within(dialog).getByRole('radio', { name: 'Revisão' }));

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() => expect(vi.mocked(routeAdminApi.createRoute)).toHaveBeenCalled());
    const body = vi.mocked(routeAdminApi.createRoute).mock.calls[0][0];
    expect(body.stages[0].stage_kind).toBe('review');
  });

  it('selecting a livre profile shows the auto-approval badge, disables "Adicionar etapa", auto-drops the untouched seed stage, and saves with zero stages (ADR 0087)', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValueOnce([
      makeProfile({ code: 'LIV', name: 'Livre', governanceClass: 'livre' }),
    ]);

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Livre' },
    });
    await within(dialog).findByRole('option', { name: 'Livre (LIV)' });
    // The dialog opens with a pristine default stage (Etapa 1) that the user
    // has not touched yet — selecting a livre profile auto-drops it instead
    // of forcing a manual "Remover" click.
    expect(within(dialog).getByLabelText('Nome da etapa 1')).toBeTruthy();
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'LIV' },
    });

    expect(
      await within(dialog).findByText('Rota livre — aprovação automática, sem etapas'),
    ).toBeTruthy();
    expect(within(dialog).queryByLabelText('Nome da etapa 1')).toBeNull();
    expect(
      within(dialog).getByText('Rota livre: aprovação automática, sem etapas.'),
    ).toBeTruthy();
    expect(
      (within(dialog).getByRole('button', { name: 'Adicionar etapa' }) as HTMLButtonElement)
        .disabled,
    ).toBe(true);

    const saveButton = within(dialog).getByRole('button', { name: /Salvar rota/i });
    expect((saveButton as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(saveButton);

    await waitFor(() => expect(vi.mocked(routeAdminApi.createRoute)).toHaveBeenCalled());
    const body = vi.mocked(routeAdminApi.createRoute).mock.calls[0][0];
    expect(body.stages).toEqual([]);
  });

  it('keeps a user-edited seed stage when switching to a livre profile, and flags it on save instead of silently discarding the edit', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));
    vi.mocked(taxonomyApi.fetchProfiles).mockResolvedValueOnce([
      makeProfile({ code: 'LIV', name: 'Livre', governanceClass: 'livre' }),
    ]);

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Livre Editada' },
    });
    // Touch the seed stage before switching profile — it is no longer
    // "pristine", so the auto-drop must NOT fire. Fill it out fully so the
    // save attempt below reaches the livre-specific stage-count error
    // instead of tripping an earlier per-field validation first.
    await fillStage1Basics(dialog, 'Etapa preenchida');

    await within(dialog).findByRole('option', { name: 'Livre (LIV)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'LIV' },
    });

    // The edited stage survives the profile switch.
    expect(within(dialog).getByLabelText('Nome da etapa 1')).toHaveValue('Etapa preenchida');

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() =>
      expect(
        within(dialog).getByText(
          'Perfil livre não admite etapas — a rota livre é aprovada automaticamente.',
        ),
      ).toBeTruthy(),
    );
    expect(vi.mocked(routeAdminApi.createRoute)).not.toHaveBeenCalled();
  });

  it('renders stage-kind and quorum pills and updates the draft on click', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    const approvalRadio = within(dialog).getByRole('radio', { name: 'Assinatura' });
    expect(approvalRadio).toBeChecked();
    fireEvent.click(within(dialog).getByRole('radio', { name: 'Revisão' }));
    expect(within(dialog).getByRole('radio', { name: 'Revisão' })).toBeChecked();

    const anyOfRadio = within(dialog).getByRole('radio', { name: 'Qualquer um (1 de N)' });
    expect(anyOfRadio).toBeChecked();
    fireEvent.click(within(dialog).getByRole('radio', { name: 'Todos (N de N)' }));
    expect(within(dialog).getByRole('radio', { name: 'Todos (N de N)' })).toBeChecked();
  });

  it('submit payload includes stage_kind and the canonical capability per stage', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Padrão' },
    });
    await within(dialog).findByRole('option', { name: 'Jurídico (JUR)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });
    await fillStage1Basics(dialog, 'Assinatura única');

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() => expect(vi.mocked(routeAdminApi.createRoute)).toHaveBeenCalled());
    const body = vi.mocked(routeAdminApi.createRoute).mock.calls[0][0];
    expect(body.stages[0].stage_kind).toBe('approval');
    // Literal on purpose: pins the default-stage seed to the Go registry
    // capability so drift in the SIGNOFF_CAPABILITY constant cannot self-green.
    expect(body.stages[0].required_capability).toBe('document.signoff');
  });

  it('rejects a review stage that comes after an approval stage', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota Fora de Ordem' },
    });
    await within(dialog).findByRole('option', { name: 'Jurídico (JUR)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });
    await fillStage1Basics(dialog, 'Assinatura');

    fireEvent.click(within(dialog).getByRole('button', { name: 'Adicionar etapa' }));
    fireEvent.change(within(dialog).getByLabelText('Nome da etapa 2'), {
      target: { value: 'Revisão tardia' },
    });
    fireEvent.change(within(dialog).getByLabelText('Role do seletor 1 da etapa 2'), {
      target: { value: 'approver' },
    });
    fireEvent.change(within(dialog).getByLabelText('Área do seletor 1 da etapa 2'), {
      target: { value: 'AREA-01' },
    });
    // Stage 1 stays 'approval' (default); flip stage 2 (second "Revisão" radio) to review.
    const reviewRadios = within(dialog).getAllByRole('radio', { name: 'Revisão' });
    fireEvent.click(reviewRadios[1]);

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() =>
      expect(within(dialog).getByText(stageOrderViolationMessage())).toBeTruthy(),
    );
    expect(vi.mocked(routeAdminApi.createRoute)).not.toHaveBeenCalled();
  });

  it('lists the subject kind per route (Documentos / Templates)', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(
      listResponse([
        makeRoute({ id: 'route-doc', name: 'Rota Doc', subject_kind: 'document' }),
        makeRoute({
          id: 'route-tpl',
          name: 'Rota Tpl',
          subject_kind: 'template',
          profile_code: 'JUR',
        }),
        // Legacy row written before subject kinds existed — reads as a
        // document route, never as a blank cell.
        makeRoute({ id: 'route-legacy', name: 'Rota Legada' }),
      ]),
    );

    renderWithProviders(<RouteAdminPage />);

    await screen.findByText('Rota Doc');
    expect(screen.getAllByLabelText('Governa: Documentos')).toHaveLength(2);
    expect(screen.getByLabelText('Governa: Templates')).toBeTruthy();
  });

  it('create sends subject_kind=template with profile_code and no subject_key (ADR 0086)', async () => {
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'Nova rota' }));
    const dialog = await screen.findByRole('dialog', { name: 'Criar rota' });

    const subjectKind = within(dialog).getByLabelText('O que a rota governa') as HTMLSelectElement;
    // Default is the document subject, preserving the legacy create path.
    expect(subjectKind.value).toBe('document');
    fireEvent.change(subjectKind, { target: { value: 'template' } });

    fireEvent.change(within(dialog).getByLabelText('Nome da rota'), {
      target: { value: 'Rota de Templates' },
    });
    await within(dialog).findByRole('option', { name: 'Jurídico (JUR)' });
    fireEvent.change(within(dialog).getByLabelText('Código do perfil'), {
      target: { value: 'JUR' },
    });
    await fillStage1Basics(dialog, 'Aprovação de template');

    fireEvent.click(within(dialog).getByRole('button', { name: /Salvar rota/i }));
    await waitFor(() => expect(vi.mocked(routeAdminApi.createRoute)).toHaveBeenCalled());

    const body = vi.mocked(routeAdminApi.createRoute).mock.calls[0][0];
    expect(body.subject_kind).toBe('template');
    // The profile keys BOTH subject kinds; subject_key is server-derived.
    expect(body.profile_code).toBe('JUR');
    expect('subject_key' in body).toBe(false);
  });

  it('subject kind is immutable on edit', async () => {
    const route = makeRoute({ subject_kind: 'template', profile_code: 'JUR' });
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([route]));

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: `Editar ${route.name}` }));
    const dialog = await screen.findByRole('dialog', { name: 'Editar rota' });

    const subjectKind = within(dialog).getByLabelText('O que a rota governa') as HTMLSelectElement;
    expect(subjectKind.value).toBe('template');
    expect(subjectKind.disabled).toBe(true);
    expect(within(dialog).getByText('O tipo de assunto é imutável após criação.')).toBeTruthy();
  });

  it('editing a route whose profile was archived out of the active list skips client-side signature-policy gating', async () => {
    // `fetchProfiles()` (mocked at module top) only returns JUR/OPS — this
    // route's profile_code is neither, simulating an archived/retired
    // profile. `governanceClass` resolves to null (S3 wiring req #2): no
    // badge, a neutral note instead, Save stays enabled, and the client never
    // fabricates a controlado signature-policy error — the backend is the
    // sole enforcer for this branch.
    const route = makeRoute({ profile_code: 'ARCH-XYZ' });
    vi.mocked(routeAdminApi.listRoutes).mockResolvedValue(listResponse([route]));
    vi.mocked(routeAdminApi.updateRoute).mockResolvedValueOnce({
      route_id: route.id,
      new_version: 4,
    });

    renderWithProviders(<RouteAdminPage />);
    fireEvent.click(await screen.findByRole('button', { name: `Editar ${route.name}` }));
    const dialog = await screen.findByRole('dialog', { name: 'Editar rota' });

    // (a) no policy badge for any known governance class renders.
    expect(within(dialog).queryByText('Obrigatório ≥1 assinatura')).toBeNull();
    expect(within(dialog).queryByText('Assinatura opcional')).toBeNull();
    expect(
      within(dialog).queryByText('Rota livre — aprovação automática, sem etapas'),
    ).toBeNull();

    // (b) the neutral "unclassified profile" note shows instead.
    expect(within(dialog).getByText('Perfil não classificado.')).toBeTruthy();

    // (c) Save is not falsely blocked by client-side gating.
    const saveButton = within(dialog).getByRole('button', { name: /Salvar rota/i });
    expect((saveButton as HTMLButtonElement).disabled).toBe(false);

    // (d) submitting does not surface the controlado signature-policy error;
    // validation is skipped and the update proceeds to the backend.
    fireEvent.click(saveButton);
    await waitFor(() => expect(vi.mocked(routeAdminApi.updateRoute)).toHaveBeenCalled());
    expect(
      screen.queryByText('Perfil controlado exige ao menos uma etapa de assinatura (aprovação).'),
    ).toBeNull();
  });
});
