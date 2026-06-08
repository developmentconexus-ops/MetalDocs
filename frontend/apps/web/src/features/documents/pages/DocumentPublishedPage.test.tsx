import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { DocumentPublishedPage } from './DocumentPublishedPage';

const mockNavigate = vi.fn();

vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({ documentId: 'doc-published-1' }),
}));

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (state: { user: { userId: string; displayName: string; roles: string[] } }) => unknown) =>
    selector({
      user: {
        userId: 'admin-user',
        displayName: 'Administrator',
        roles: ['system_admin'],
      },
    }),
}));

vi.mock('../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));

vi.mock('../queries/useApprovalInstanceQuery', () => ({
  useApprovalInstanceQuery: vi.fn(),
}));

vi.mock('../queries/useControlledDocumentActiveDocumentQuery', () => ({
  useControlledDocumentActiveDocumentQuery: vi.fn(),
}));

vi.mock('../queries/useDocumentRevisionHistoryQuery', () => ({
  useDocumentRevisionHistoryQuery: vi.fn(),
}));

vi.mock('../queries/useAreasQuery', () => ({
  useAreasQuery: vi.fn(() => ({ data: [{ code: 'general', name: 'Geral' }], isLoading: false })),
}));

vi.mock('../../taxonomy/queries/useProfilesQuery', () => ({
  useProfilesQuery: vi.fn(() => ({ data: [{ code: 'pop', name: 'Procedimento Operacional' }], isLoading: false })),
}));

vi.mock('../../controlled-documents/api/controlledDocuments', () => ({
  createRevision: vi.fn(),
}));

import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { createRevision } from '../../controlled-documents/api/controlledDocuments';

describe('DocumentPublishedPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'published',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
        current_revision_file_size_bytes: 1024,
        current_revision_page_count: 1,
        current_revision_page_count_source: 'server_renderer',
      },
      refetch: vi.fn(),
    } as never);
    vi.mocked(useApprovalInstanceQuery).mockReturnValue({
      data: {
        completed_at: '2026-05-19T23:39:00.000Z',
        stages: [],
      },
      isLoading: false,
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-published-1',
        content_hash: 'hash-1',
        approval_state: 'published',
      },
      isLoading: false,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useDocumentRevisionHistoryQuery).mockReturnValue({
      data: {
        items: [
          {
            document_id: 'doc-published-1',
            revision_number: 1,
            revision_title: 'E2E Approval Flow 2026-05-19',
            status: 'published',
            created_at: '2026-05-19T23:39:00.000Z',
            is_current: true,
          },
        ],
      },
      isLoading: false,
    } as never);
  });

  it('delegates scheduled lifecycle polling to the query layer', () => {
    const detailRefetch = vi.fn();
    const activeRefetch = vi.fn();
    const historyRefetch = vi.fn();

    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'scheduled',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: detailRefetch,
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-published-1',
        content_hash: 'hash-1',
        approval_state: 'scheduled',
      },
      isLoading: false,
      refetch: activeRefetch,
    } as never);
    vi.mocked(useDocumentRevisionHistoryQuery).mockReturnValue({
      data: {
        items: [
          {
            document_id: 'doc-published-1',
            revision_number: 1,
            revision_title: 'E2E Approval Flow 2026-05-19',
            status: 'scheduled',
            created_at: '2026-05-19T23:39:00.000Z',
            is_current: true,
          },
        ],
      },
      isLoading: false,
      refetch: historyRefetch,
    } as never);

    render(<DocumentPublishedPage />);

    expect(useDocumentDetailQuery).toHaveBeenLastCalledWith('doc-published-1', { pollScheduledLifecycle: true });
    expect(useControlledDocumentActiveDocumentQuery).toHaveBeenLastCalledWith('cd-1', { refetchInterval: 5_000 });
    expect(useDocumentRevisionHistoryQuery).toHaveBeenLastCalledWith('doc-published-1', { refetchInterval: 5_000 });
  });

  it('disables scheduled lifecycle polling once the governed detail is stable', () => {
    const detailRefetch = vi.fn();
    const activeRefetch = vi.fn();
    const historyRefetch = vi.fn();

    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'published',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: detailRefetch,
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-published-1',
        content_hash: 'hash-1',
        approval_state: 'published',
      },
      isLoading: false,
      refetch: activeRefetch,
    } as never);
    vi.mocked(useDocumentRevisionHistoryQuery).mockReturnValue({
      data: {
        items: [
          {
            document_id: 'doc-published-1',
            revision_number: 1,
            revision_title: 'E2E Approval Flow 2026-05-19',
            status: 'published',
            created_at: '2026-05-19T23:39:00.000Z',
            is_current: true,
          },
        ],
      },
      isLoading: false,
      refetch: historyRefetch,
    } as never);

    render(<DocumentPublishedPage />);

    expect(useDocumentDetailQuery).toHaveBeenLastCalledWith('doc-published-1', { pollScheduledLifecycle: true });
    expect(useControlledDocumentActiveDocumentQuery).toHaveBeenLastCalledWith('cd-1', { refetchInterval: false });
    expect(useDocumentRevisionHistoryQuery).toHaveBeenLastCalledWith('doc-published-1', { refetchInterval: false });
  });

  it('creates a new revision from the published document screen', async () => {
    vi.mocked(createRevision).mockResolvedValue({
      document: { id: 'doc-draft-2' },
    } as never);

    render(<DocumentPublishedPage />);

    fireEvent.click(screen.getByRole('button', { name: /Iniciar revis/i }));
    const nameInput = screen.getByLabelText(/Nome do documento/i) as HTMLInputElement;
    expect(nameInput.value).toBe('E2E Approval Flow 2026-05-19');
    fireEvent.change(nameInput, {
      target: { value: 'E2E Approval Flow 2026-05-19 REV01' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Gerar documento' }));

    await waitFor(() => {
      expect(createRevision).toHaveBeenCalledWith(
        'cd-1',
        { name: 'E2E Approval Flow 2026-05-19 REV01', form_data: {}, template_version_id: 'template-1' },
        expect.any(String),
      );
    });
    expect(mockNavigate).toHaveBeenCalledWith('/documents/doc-draft-2/edit');
  });

  it.each([
    { state: 'draft', ctaLabel: 'Continuar rascunho', expectedUrl: '/documents/doc-sibling-active/edit' },
    { state: 'under_review', ctaLabel: 'Acompanhar revisão', expectedUrl: '/documents/doc-sibling-active' },
    { state: 'approved', ctaLabel: 'Publicar revisão aprovada', expectedUrl: '/documents/doc-sibling-active' },
    { state: 'scheduled', ctaLabel: 'Ver publicação agendada', expectedUrl: '/documents/doc-sibling-active' },
    { state: 'rejected', ctaLabel: 'Retomar revisão rejeitada', expectedUrl: '/documents/doc-sibling-active/edit' },
  ])('sibling active: uses truthful destination for $state and blocks new revision', ({ state, ctaLabel, expectedUrl }) => {
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-sibling-active',
        published_document_id: 'doc-published-1',
        content_hash: 'hash-sibling-1',
        approval_state: state,
        revision_version: 0,
      },
      isLoading: false,
      refetch: vi.fn(),
    } as never);

    render(<DocumentPublishedPage />);

    expect(screen.queryByRole('button', { name: /Iniciar revis/i })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: ctaLabel }));

    expect(createRevision).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith(expectedUrl);
  });

  it('shows publish action instead of new revision for approved documents', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'approved',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: vi.fn(),
    } as never);

    render(<DocumentPublishedPage />);

    expect(screen.getByRole('button', { name: 'Publicar / Agendar' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: /Iniciar revis/i })).toBeNull();
  });

  it('blocks publish with visible inline guidance when active publish context is missing', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'approved',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: vi.fn(),
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-published-1',
        approval_state: 'approved',
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    } as never);

    render(<DocumentPublishedPage />);

    const publishButton = screen.getByRole('button', { name: 'Publicar / Agendar' }) as HTMLButtonElement;
    expect(publishButton).toBeTruthy();
    expect(publishButton.getAttribute('aria-disabled')).toBe('true');
    expect(screen.getByText('A publicação está bloqueada porque o contexto ativo desta revisão ainda não foi confirmado.')).toBeTruthy();
  });

  it('surfaces active-document lookup failure inline and keeps publish blocked', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'approved',
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: vi.fn(),
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: vi.fn(),
    } as never);

    render(<DocumentPublishedPage />);

    const publishButton = screen.getByRole('button', { name: 'Publicar / Agendar' }) as HTMLButtonElement;
    expect(publishButton.getAttribute('aria-disabled')).toBe('true');
    expect(screen.getByText('Não foi possível confirmar o contexto ativo de publicação. Atualize a página e tente novamente antes de publicar.')).toBeTruthy();
  });

  it('renders governed revision history instead of placeholder versions', () => {
    render(<DocumentPublishedPage />);

    expect(screen.getByRole('button', { name: /REV01/ })).toBeTruthy();
    expect(screen.getByText('passe o mouse para detalhar')).toBeTruthy();
    expect(screen.getByText('atual')).toBeTruthy();
    expect(screen.queryByText('v3.2')).toBeNull();
  });

  it('uses governed revision code in the hero instead of technical revision version', () => {
    render(<DocumentPublishedPage />);

    expect(screen.getAllByText('REV01').length).toBeGreaterThan(0);
    expect(screen.queryByText('v2 · vigente')).toBeNull();
    expect(screen.queryByText('v2 · aprovado')).toBeNull();
  });

  it.each([
    {
      status: 'approved',
      badgeLabel: 'REV01 · aprovado',
      subtitle: 'aprovado em 19 de maio de 2026',
      ownerMeta: 'aprovado em 19 de maio de 2026',
    },
    {
      status: 'scheduled',
      badgeLabel: 'REV01 · agendado',
      subtitle: 'aprovação concluída em 19 de maio de 2026',
      ownerMeta: 'publicação agendada · aprovado em 19 de maio de 2026',
    },
    {
      status: 'published',
      badgeLabel: 'REV01 · publicado',
      subtitle: 'publicado em 19 de maio de 2026',
      ownerMeta: 'publicado em 19 de maio de 2026',
    },
    {
      status: 'superseded',
      badgeLabel: 'REV01 · substituído',
      subtitle: 'publicado em 19 de maio de 2026',
      ownerMeta: 'substituído · publicado em 19 de maio de 2026',
    },
  ])('renders truthful governed status copy for $status documents', ({ status, badgeLabel, subtitle, ownerMeta }) => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-published-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: status,
        form_data_json: {},
        current_revision_id: 'rev-1',
        revision_version: 2,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-19T00:00:00.000Z',
        updated_at: '2026-05-19T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 1,
        controlled_document_id: 'cd-1',
        revision_title: 'E2E Approval Flow 2026-05-19',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: vi.fn(),
    } as never);

    render(<DocumentPublishedPage />);

    expect(screen.getByText(badgeLabel)).toBeTruthy();
    expect(screen.getAllByText(subtitle).length).toBeGreaterThan(0);
    expect(screen.getAllByText(ownerMeta).length).toBeGreaterThan(0);
    expect(screen.queryByText('REV01 · vigente')).toBeNull();
  });

  it('shows the published head as the KPI current version when the opened revision is scheduled', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        id: 'doc-scheduled-1',
        tenant_id: 'tenant-1',
        template_version_id: 'template-1',
        name: 'E2E Approval Flow 2026-05-19',
        status: 'scheduled',
        form_data_json: {},
        current_revision_id: 'rev-2',
        revision_version: 3,
        active_session_id: '',
        values_frozen_at: null,
        archived_at: null,
        created_at: '2026-05-20T00:00:00.000Z',
        updated_at: '2026-05-20T00:00:00.000Z',
        created_by: 'admin-user',
        revision_number: 9,
        controlled_document_id: 'cd-1',
        revision_title: 'S2 boundary QA revision',
        profile_code_snapshot: 'pop',
        process_area_code_snapshot: 'general',
        code: 'POP-GENERAL-014',
      },
      refetch: vi.fn(),
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: {
        document_id: 'doc-scheduled-1',
        published_document_id: 'doc-published-1',
        content_hash: 'hash-scheduled-1',
        approval_state: 'scheduled',
        revision_version: 3,
      },
      isLoading: false,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useDocumentRevisionHistoryQuery).mockReturnValue({
      data: {
        items: [
          {
            document_id: 'doc-scheduled-1',
            revision_number: 9,
            revision_title: 'S2 boundary QA revision',
            status: 'scheduled',
            created_at: '2026-05-20T16:12:00.000Z',
            is_current: true,
          },
          {
            document_id: 'doc-published-1',
            revision_number: 8,
            revision_title: 'QA Matrix runtime submission title',
            status: 'published',
            created_at: '2026-05-20T15:12:00.000Z',
            is_current: false,
          },
        ],
      },
      isLoading: false,
    } as never);

    render(<DocumentPublishedPage />);

    const currentVersionCell = screen.getByText(/Vers.*o atual/i).parentElement;
    expect(currentVersionCell?.textContent).toContain('REV08');
    expect(currentVersionCell?.textContent).not.toContain('REV09');
    expect(currentVersionCell?.textContent).not.toContain('desde');
  });

  it('keeps hook order stable when approved data arrives after initial loading', () => {
    const refetch = vi.fn();

    vi.mocked(useDocumentDetailQuery)
      .mockReturnValueOnce({
        isLoading: true,
        isError: false,
        data: undefined,
        refetch,
      } as never)
      .mockReturnValueOnce({
        isLoading: false,
        isError: false,
        data: {
          id: 'doc-published-1',
          tenant_id: 'tenant-1',
          template_version_id: 'template-1',
          name: 'E2E Approval Flow 2026-05-19',
          status: 'approved',
          form_data_json: {},
          current_revision_id: 'rev-1',
          revision_version: 2,
          active_session_id: '',
          values_frozen_at: null,
          archived_at: null,
          created_at: '2026-05-19T00:00:00.000Z',
          updated_at: '2026-05-19T00:00:00.000Z',
          created_by: 'admin-user',
          revision_number: 1,
          controlled_document_id: 'cd-1',
          revision_title: 'E2E Approval Flow 2026-05-19',
          profile_code_snapshot: 'pop',
          process_area_code_snapshot: 'general',
          code: 'POP-GENERAL-014',
        },
        refetch,
      } as never);

    const { rerender } = render(<DocumentPublishedPage />);

    expect(screen.getByText('Carregando documento…')).toBeTruthy();

    rerender(<DocumentPublishedPage />);

    expect(screen.getByRole('button', { name: 'Publicar / Agendar' })).toBeTruthy();
  });
});




