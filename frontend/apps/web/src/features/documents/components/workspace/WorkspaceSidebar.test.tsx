import type { ComponentProps } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// F-QA4-4 — the IDENTIFICAÇÃO panel is fed by live queries; mocking the query
// hooks (the useDocumentArtifact.test.tsx pattern) lets each case pin exactly
// which wire fact produces which rendered row.
vi.mock('../../queries/useAreasQuery', () => ({ useAreasQuery: vi.fn() }));
vi.mock('../../../taxonomy/queries/useProfilesQuery', () => ({ useProfilesQuery: vi.fn() }));
vi.mock('../../../controlled-documents/queries/useControlledDocumentDetailQuery', () => ({
  useControlledDocumentDetailQuery: vi.fn(),
}));
vi.mock('../../queries/useDocumentRevisionHistoryQuery', () => ({
  useDocumentRevisionHistoryQuery: vi.fn(),
}));

import { WorkspaceSidebar } from './WorkspaceSidebar';
import { useAreasQuery } from '../../queries/useAreasQuery';
import { useProfilesQuery } from '../../../taxonomy/queries/useProfilesQuery';
import { useControlledDocumentDetailQuery } from '../../../controlled-documents/queries/useControlledDocumentDetailQuery';
import { useDocumentRevisionHistoryQuery } from '../../queries/useDocumentRevisionHistoryQuery';
import type { DocumentDetail } from '../../api/documents';
import type { ApprovalInstance, StageInstance } from '../../../approval/api/approvalTypes';

function asQuery<T>(value: unknown): T {
  return value as T;
}

const AREAS = [{ code: 'usinagem', name: 'Usinagem' }];
const PROFILES = [{ code: 'it', name: 'Instrução de Trabalho' }];
const VISIBILITY_COMPANY = { scope: 'company' as const, area_codes: [], user_ids: [] };

function mockIdentityQueries(
  overrides: {
    areas?: unknown;
    profiles?: unknown;
    controlledDocument?: unknown;
    revisionHistory?: unknown;
  } = {},
) {
  vi.mocked(useAreasQuery).mockReturnValue(
    asQuery(overrides.areas ?? { data: AREAS, isLoading: false, isError: false }),
  );
  vi.mocked(useProfilesQuery).mockReturnValue(
    asQuery(overrides.profiles ?? { data: PROFILES, isLoading: false, isError: false }),
  );
  vi.mocked(useControlledDocumentDetailQuery).mockReturnValue(
    asQuery(
      overrides.controlledDocument ?? {
        data: { visibility: VISIBILITY_COMPANY },
        isLoading: false,
        isError: false,
      },
    ),
  );
  vi.mocked(useDocumentRevisionHistoryQuery).mockReturnValue(
    asQuery(overrides.revisionHistory ?? { data: undefined, isLoading: false, isError: false }),
  );
}

function makeDoc(overrides: Partial<DocumentDetail> = {}): DocumentDetail {
  return {
    id: 'doc-1',
    code: 'POP-QUA-0148',
    name: 'POP Limpeza de Linha',
    status: 'under_review',
    revision_version: 3,
    revision_number: 3,
    controlled_document_id: 'cd-1',
    current_revision_id: 'rev-1',
    created_by: 'user-author-1',
    created_at: '2026-05-01T10:00:00.000Z',
    profile_code_snapshot: 'it',
    process_area_code_snapshot: 'usinagem',
    current_revision_page_count: 4,
    current_revision_file_size_bytes: 1304,
    ...overrides,
  } as DocumentDetail;
}

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    stage_kind: 'review',
    signoffs: [],
    actors: [],
    due_at: null,
    ...overrides,
  } as StageInstance;
}

function makeInstance(overrides: Partial<ApprovalInstance> = {}): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'r1',
    tenant_id: 't1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-05-01T10:00:00.000Z',
    completed_at: null,
    stages: [makeStage()],
    etag: '"v1"',
    frozen_content_hash: null,
    viewer: {
      is_author: false,
      eligible_for_active_stage: true,
      has_signed_active_stage: false,
      via_delegation_from: null,
    },
    verdicts: [],
    ...overrides,
  } as ApprovalInstance;
}

type Props = ComponentProps<typeof WorkspaceSidebar>;

function renderSidebar(overrides: Partial<Props> = {}) {
  const doc = overrides.doc ?? makeDoc();
  const instance = 'instance' in overrides ? (overrides.instance ?? null) : makeInstance();
  const props: Props = {
    documentId: 'doc-1',
    doc,
    instance,
    instanceLoading: false,
    instanceError: null,
    onRetryInstance: vi.fn(),
    activeStage: instance?.stages.find((s) => s.status === 'active'),
    onRefetchInstance: vi.fn(),
    ...overrides,
  };
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <WorkspaceSidebar {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('WorkspaceSidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockIdentityQueries();
  });

  it('renders the embedded meta panel, the timeline, and a link to the record view', () => {
    renderSidebar();
    expect(screen.getByText('Identificação')).toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
    const link = screen.getByRole('link', { name: /ficha completa/i });
    expect(link).toHaveAttribute('href', '/documents/doc-1/details');
  });

  it('renders verdict CTAs when the viewer is eligible on an active review stage', () => {
    renderSidebar();
    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Solicitar mudanças' })).toBeInTheDocument();
  });

  it('renders no verdict CTAs and no footer for an ineligible (observing) viewer', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({ instance, activeStage: instance.stages[0] });
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
  });

  it('renders no footer at all when there is no approval instance, but keeps the meta panel', () => {
    renderSidebar({ instance: null, activeStage: undefined });
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
    expect(screen.getByText('Identificação')).toBeInTheDocument();
  });

  it('renders a passed lifecycle action in the reviewing footer and fires run on click', () => {
    const run = vi.fn();
    renderSidebar({
      lifecycleActions: [
        { key: 'cancel', label: 'Cancelar instância', variant: 'secondary', available: true, run },
      ],
    });
    // Verdict CTAs (reviewing) AND the lifecycle action both render.
    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeInTheDocument();
    const cancelBtn = screen.getByRole('button', { name: 'Cancelar instância' });
    expect(cancelBtn).toBeInTheDocument();
    cancelBtn.click();
    expect(run).toHaveBeenCalledTimes(1);
  });

  it('renders a lifecycle action in the observing footer even with no verdict CTAs', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({
      instance,
      activeStage: instance.stages[0],
      lifecycleActions: [
        { key: 'cancel', label: 'Cancelar instância', variant: 'secondary', available: true, run: vi.fn() },
      ],
    });
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.getByTestId('approval-sidebar-footer')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancelar instância' })).toBeInTheDocument();
  });

  it('renders no lifecycle footer when the action list is empty and the viewer is observing', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({ instance, activeStage: instance.stages[0], lifecycleActions: [] });
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
  });

  // ---------------------------------------------------------------------
  // F-QA4-4 — IDENTIFICAÇÃO panel bound to real query data (no `---`).
  // ---------------------------------------------------------------------

  it('renders tipo / área / visibilidade from the document snapshots + controlled document', () => {
    renderSidebar();

    expect(screen.getByText('Instrução de Trabalho')).toBeInTheDocument();
    expect(screen.getByText('Usinagem')).toBeInTheDocument();
    expect(screen.getByText('Toda empresa')).toBeInTheDocument();
    expect(screen.getByText('4 páginas · 1,3 KB')).toBeInTheDocument();
    // The retired placeholder must never come back.
    expect(screen.queryByText('---')).not.toBeInTheDocument();
  });

  it('falls back to the raw snapshot codes while the taxonomy lists are unresolved', () => {
    mockIdentityQueries({
      areas: { data: undefined, isLoading: true, isError: false },
      profiles: { data: undefined, isLoading: true, isError: false },
    });
    renderSidebar();

    expect(screen.getByText('it')).toBeInTheDocument();
    expect(screen.getByText('usinagem')).toBeInTheDocument();
  });

  it('renders a restricted visibility label resolved through the area names', () => {
    mockIdentityQueries({
      controlledDocument: {
        data: { visibility: { scope: 'restricted', area_codes: ['usinagem'], user_ids: [] } },
        isLoading: false,
        isError: false,
      },
    });
    renderSidebar();

    expect(screen.getByText('Restrito a area Usinagem')).toBeInTheDocument();
  });

  it('explains the absent visibility when the revision has no controlled document yet', () => {
    mockIdentityQueries({
      controlledDocument: { data: undefined, isLoading: false, isError: false },
    });
    renderSidebar({ doc: makeDoc({ controlled_document_id: null }) });

    expect(screen.getByTitle(/ainda não está vinculada a um/i)).toHaveTextContent('—');
  });

  it('explains the absent visibility when the controlled-document read fails', () => {
    mockIdentityQueries({
      controlledDocument: { data: undefined, isLoading: false, isError: true },
    });
    renderSidebar();

    expect(screen.getByTitle(/Não foi possível carregar a visibilidade/i)).toHaveTextContent('—');
  });

  it('explains absent snapshots and page count instead of rendering a silent placeholder', () => {
    renderSidebar({
      doc: makeDoc({
        profile_code_snapshot: null,
        process_area_code_snapshot: null,
        current_revision_page_count: null,
        current_revision_file_size_bytes: null,
      }),
    });

    expect(screen.getByTitle(/não tem perfil registrado/i)).toHaveTextContent('—');
    expect(screen.getByTitle(/não tem área responsável registrada/i)).toHaveTextContent('—');
    expect(screen.getByTitle(/contagem de páginas/i)).toHaveTextContent('—');
    expect(screen.queryByText('---')).not.toBeInTheDocument();
  });

  it('renders the real revision lineage returned by the revision-history query', () => {
    mockIdentityQueries({
      revisionHistory: {
        data: {
          items: [
            {
              document_id: 'doc-1',
              revision_number: 3,
              revision_title: 'Ajuste de parâmetros de corte',
              status: 'under_review',
              created_at: '2026-05-01T10:00:00.000Z',
              is_current: true,
            },
            {
              document_id: 'doc-0',
              revision_number: 2,
              revision_title: 'Revisão periódica',
              status: 'published',
              created_at: '2026-03-01T10:00:00.000Z',
              is_current: false,
            },
          ],
        },
        isLoading: false,
        isError: false,
      },
    });
    renderSidebar();

    expect(screen.getByText('REV03')).toBeInTheDocument();
    expect(screen.getByText('Ajuste de parâmetros de corte')).toBeInTheDocument();
    expect(screen.getByText('REV02')).toBeInTheDocument();
    expect(screen.getByText('Revisão periódica')).toBeInTheDocument();
  });

  it('falls back to the document-derived current revision row when history is unavailable', () => {
    renderSidebar({ doc: makeDoc({ revision_title: 'Criação do documento' }) });

    expect(screen.getByText('REV03')).toBeInTheDocument();
    expect(screen.getByText('Criação do documento')).toBeInTheDocument();
  });

  it('surfaces an instance error in the timeline panel and wires the retry callback', () => {
    const onRetryInstance = vi.fn();
    renderSidebar({
      instance: null,
      activeStage: undefined,
      instanceError: 'Erro ao carregar dados de aprovação.',
      onRetryInstance,
    });
    expect(screen.getByRole('alert')).toHaveTextContent('Erro ao carregar dados de aprovação.');
    screen.getByRole('button', { name: 'Tentar novamente' }).click();
    expect(onRetryInstance).toHaveBeenCalledTimes(1);
  });
});
