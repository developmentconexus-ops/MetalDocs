import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (state: { user: { userId: string; displayName: string; roles: string[] } }) => unknown) =>
    selector({
      user: { userId: 'admin-user', displayName: 'Administrator', roles: ['system_admin'] },
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
vi.mock('../queries/useDistributionSummaryQuery', () => ({
  useDistributionSummaryQuery: vi.fn(),
}));

import { useDocumentDetailQuery } from '../queries/useDocumentDetailQuery';
import { useApprovalInstanceQuery } from '../queries/useApprovalInstanceQuery';
import { useControlledDocumentActiveDocumentQuery } from '../queries/useControlledDocumentActiveDocumentQuery';
import { useDocumentRevisionHistoryQuery } from '../queries/useDocumentRevisionHistoryQuery';
import { useDistributionSummaryQuery } from '../queries/useDistributionSummaryQuery';
import { useDocumentArtifact } from './useDocumentArtifact';

const BASE_DOC = {
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
  current_revision_page_count: 3,
};

describe('useDocumentArtifact', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: BASE_DOC,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useApprovalInstanceQuery).mockReturnValue({
      data: {
        completed_at: '2026-05-19T23:39:00.000Z',
        stages: [
          {
            id: 'stage-0',
            stage_index: 0,
            label: 'Qualidade',
            status: 'passed',
            actors: [],
            signoffs: [
              {
                id: 'so-0',
                actor_user_id: 'approver-1',
                decision: 'approve',
                signature_method: 'click',
                signed_at: '2026-05-19T23:39:00.000Z',
              },
            ],
          },
        ],
      },
      isLoading: false,
      refetch: vi.fn(),
    } as never);
    vi.mocked(useControlledDocumentActiveDocumentQuery).mockReturnValue({
      data: { document_id: 'doc-published-1', content_hash: 'hash-1', approval_state: 'published' },
      isLoading: false,
      isError: false,
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
      refetch: vi.fn(),
    } as never);
    vi.mocked(useDistributionSummaryQuery).mockReturnValue({
      data: { total_targets: 12 },
      isError: false,
      isLoading: false,
    } as never);
  });

  it('maps the document detail into a kind=document view-model with code/title/status', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.kind).toBe('document');
    expect(model.id).toBe('doc-published-1');
    expect(model.code).toBe('POP-GENERAL-014');
    expect(model.title).toBe('E2E Approval Flow 2026-05-19');
    expect(model.status).toBe('published');
    expect(model.versionNumber).toBe(1);
    expect(model.revisionLabel).toBe('REV01');
  });

  it('builds the hero breadcrumb, status badge, and subtitle from the governed status', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.hero.breadcrumb).toEqual([
      { label: 'Biblioteca', href: '/documents' },
      { label: 'Geral' },
      { label: 'POP-GENERAL-014' },
    ]);
    expect(model.hero.badges).toContainEqual({ label: 'POP-GENERAL-014', variant: 'code' });
    expect(model.hero.badges).toContainEqual({ label: 'REV01 · publicado', variant: 'status' });
    expect(model.hero.badges).toContainEqual({ label: 'Procedimento Operacional', variant: 'type' });
    expect(model.hero.subtitle).toBe('publicado em 19 de maio de 2026');
  });

  it('maps profile/area/file metadata into model.meta with resolver labels', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.meta.profileLabel).toBe('Procedimento Operacional');
    expect(model.meta.areaLabel).toBe('Geral');
    expect(model.meta.fileSizeBytes).toBe(1024);
    expect(model.meta.pageCount).toBe(3);
    expect(model.meta.effectiveFrom).toBe('2026-05-19T23:39:00.000Z');
    expect(model.meta.nextReviewAt).toBeNull();
  });

  it('maps the approval instance stages/signoffs into approvalChain', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.approvalChain).not.toBeNull();
    expect(model.approvalChain).toHaveLength(1);
    expect(model.approvalChain?.[0]).toMatchObject({
      stageIndex: 0,
      label: 'Qualidade',
      status: 'passed',
      actorUserId: 'approver-1',
      actorDisplay: 'approver-1',
      decision: 'approve',
      signedAt: '2026-05-19T23:39:00.000Z',
    });
  });

  it('returns null approvalChain when there is no approval instance', () => {
    vi.mocked(useApprovalInstanceQuery).mockReturnValue({ data: undefined, isLoading: false, refetch: vi.fn() } as never);

    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
    expect(model.approvalChain).toBeNull();
  });

  it('maps the governed revision history into lineage', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.lineage).toHaveLength(1);
    expect(model.lineage[0]).toMatchObject({
      revisionNumber: 1,
      revisionLabel: 'REV01',
      status: 'published',
      title: 'E2E Approval Flow 2026-05-19',
      isCurrent: true,
    });
  });

  it('exposes the documento + distribuição tabs', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.tabs).toEqual([
      { key: 'documento', label: 'Documento', href: '.' },
      { key: 'distribuicao', label: 'Distribuição', href: 'distribution' },
    ]);
  });

  it('exposes the full action set with every key present', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(Object.keys(model.actions).sort()).toEqual(
      ['approve', 'createVersion', 'publish', 'reject', 'review', 'submit'].sort(),
    );
    // Published document with no active sibling: createVersion is available.
    expect(model.actions.createVersion.available).toBe(true);
  });

  // ── Governed behavior guards (D1–D3) ────────────────────────────────────────

  it.each([
    {
      status: 'approved',
      subtitle: 'aprovado em 19 de maio de 2026',
      ownerDescriptor: 'aprovado em 19 de maio de 2026',
    },
    {
      status: 'scheduled',
      subtitle: 'aprovação concluída em 19 de maio de 2026',
      ownerDescriptor: 'publicação agendada · aprovado em 19 de maio de 2026',
    },
    {
      status: 'published',
      subtitle: 'publicado em 19 de maio de 2026',
      ownerDescriptor: 'publicado em 19 de maio de 2026',
    },
    {
      status: 'superseded',
      subtitle: 'publicado em 19 de maio de 2026',
      ownerDescriptor: 'substituído · publicado em 19 de maio de 2026',
    },
  ])(
    'D1 — per-status: hero.subtitle and meta.ownerDescriptor are correct and distinct for $status',
    ({ status, subtitle, ownerDescriptor }) => {
      vi.mocked(useDocumentDetailQuery).mockReturnValue({
        isLoading: false,
        isError: false,
        data: { ...BASE_DOC, status },
        refetch: vi.fn(),
      } as never);

      const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

      expect(model.hero.subtitle).toBe(subtitle);
      expect(model.meta.ownerDescriptor).toBe(ownerDescriptor);
    },
  );

  it('D1 — scheduled and superseded have DISTINCT subtitle vs ownerDescriptor', () => {
    // scheduled
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, status: 'scheduled' },
      refetch: vi.fn(),
    } as never);
    const scheduledResult = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
    expect(scheduledResult.model.hero.subtitle).not.toBe(scheduledResult.model.meta.ownerDescriptor);

    // superseded
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, status: 'superseded' },
      refetch: vi.fn(),
    } as never);
    const supersededResult = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
    expect(supersededResult.model.hero.subtitle).not.toBe(supersededResult.model.meta.ownerDescriptor);
  });

  it('D2 — scheduled fixture: currentVersion KPI shows published-head label (REV08 not REV09), hint has no "desde"', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, id: 'doc-scheduled-1', status: 'scheduled', revision_number: 9 },
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
      isError: false,
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
      refetch: vi.fn(),
    } as never);

    const { model } = renderHook(() => useDocumentArtifact('doc-scheduled-1')).result.current;
    const currentVersionCell = model.kpis.find((k) => k.key === 'currentVersion');

    expect(currentVersionCell?.value).toBe('REV08');
    expect(currentVersionCell?.value).not.toBe('REV09');
    expect(currentVersionCell?.hint).not.toMatch(/desde/);
  });

  it('D3 — distribution fixture: coverage cell value is "12" when total_targets=12', () => {
    vi.mocked(useDistributionSummaryQuery).mockReturnValue({
      data: { total_targets: 12 },
      isError: false,
      isLoading: false,
    } as never);

    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
    const coverageCell = model.kpis.find((k) => k.key === 'coverage');

    expect(coverageCell?.value).toBe('12');
    expect(coverageCell?.href).toBe('/distribution');
  });

  it('D3 — distribution error: coverage cell value is "—" (never fabricated 0)', () => {
    vi.mocked(useDistributionSummaryQuery).mockReturnValue({
      data: undefined,
      isError: true,
      isLoading: false,
    } as never);

    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
    const coverageCell = model.kpis.find((k) => k.key === 'coverage');

    expect(coverageCell?.value).toBe('—');
  });

  it('surfaces loading and error from the detail query', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: true,
      isError: false,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    const loading = useDocumentArtifact('doc-published-1');
    expect(loading.isLoading).toBe(true);

    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: true,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    const errored = useDocumentArtifact('doc-published-1');
    expect(errored.isError).toBe(true);
  });
});
