import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (state: { user: { userId: string; displayName: string; roles: string[] } }) => unknown) =>
    selector({
      user: { userId: 'admin-user', displayName: 'Administrator', roles: ['system_admin'] },
    }),
}));

// FE-11: revision-initiate gating is capability-based (ADR 0022), not role-based.
// This mock only needs to satisfy the hook call — DocumentDetailRoute.test.tsx
// covers the gating behavior end to end.
vi.mock('../../../lib/iam/useHasCapability', () => ({
  useHasCapability: vi.fn(() => true),
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
// NOT mocked — the real settling predicate, so the gate the (mocked) detail query
// polls on is covered here too.
import { isDocumentLifecycleSettling } from '../lib/documentReleasePresentation';
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
            actors: [
              {
                user_id: 'approver-1',
                display_name: 'Ana Souza',
                status: 'approved',
                decision: 'approve',
              },
            ],
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
    expect(model.hero.badges).toContainEqual({ key: 'code', label: 'POP-GENERAL-014', variant: 'code' });
    expect(model.hero.badges).toContainEqual({ key: 'status', label: 'REV01 · publicado', variant: 'status' });
    expect(model.hero.badges).toContainEqual({ key: 'type', label: 'Procedimento Operacional', variant: 'type' });
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

  it('maps the approval instance stage actors into approvalChain with display names', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.approvalChain).not.toBeNull();
    expect(model.approvalChain).toHaveLength(1);
    expect(model.approvalChain?.[0]).toMatchObject({
      stageIndex: 0,
      label: 'Qualidade',
      status: 'approved',
      actorUserId: 'approver-1',
      // F-QA4-8: the backend-resolved display name, never the raw user id.
      actorDisplay: 'Ana Souza',
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

  it('emits an empty actions array (detail model has no action sidebar)', () => {
    const { model } = renderHook(() => useDocumentArtifact('doc-published-1')).result.current;

    expect(model.actions).toEqual([]);
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
    expect(coverageCell?.href).toBe('distribution');
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

  // ── ADR 0085 Stage C: release readiness-hold projection ────────────────────
  // The projection is the ONLY source of release status. It rides the existing
  // document-detail response — these cases add no query.

  function withRelease(release: Record<string, unknown> | null) {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, release },
      refetch: vi.fn(),
    } as never);
    return renderHook(() => useDocumentArtifact('doc-published-1')).result.current;
  }

  const HOLD_BASE = {
    generation_id: 'gen-1',
    state: 'hold',
    hold_reason: null,
    hold_detail: null,
    planned_effective_from: null,
    released_at: null,
    last_evaluated_at: '2026-07-28T10:00:00.000Z',
  };

  it('W4 — released: "Publicado" with the released_at date', () => {
    const { release } = withRelease({
      ...HOLD_BASE,
      state: 'released',
      hold_reason: null,
      // Midday UTC so the pt-BR calendar day is stable in any runner timezone.
      released_at: '2026-05-19T12:00:00.000Z',
    });

    expect(release).toEqual({
      tone: 'released',
      title: 'Publicado',
      detail: 'Liberado em 19 de maio de 2026',
    });
  });

  it('W4 — released without released_at: no fabricated date', () => {
    const { release } = withRelease({ ...HOLD_BASE, state: 'released' });

    expect(release).toEqual({ tone: 'released', title: 'Publicado', detail: null });
  });

  it.each(['materializing', 'awaiting_approval_fact'])(
    'W4 — hold/%s is a transient processing state (no user action)',
    (holdReason) => {
      const { release } = withRelease({ ...HOLD_BASE, hold_reason: holdReason });

      expect(release).toEqual({
        tone: 'progress',
        title: 'Preparando artefatos finais…',
        detail: null,
      });
    },
  );

  it('W4 — hold/awaiting_effective_date announces the planned effective date', () => {
    const { release } = withRelease({
      ...HOLD_BASE,
      hold_reason: 'awaiting_effective_date',
      planned_effective_from: '2026-08-03T12:00:00.000Z',
    });

    expect(release).toEqual({
      tone: 'scheduled',
      title: 'Aprovado — vigência programada para 03/08/2026',
      detail: null,
    });
  });

  it('W4 — hold/awaiting_effective_date without a planned date stays honest', () => {
    const { release } = withRelease({ ...HOLD_BASE, hold_reason: 'awaiting_effective_date' });

    expect(release).toEqual({
      tone: 'scheduled',
      title: 'Aprovado — aguardando a data de vigência',
      detail: null,
    });
  });

  it.each([
    ['supersede_conflict', 'Publicação retida — conflito de substituição'],
    ['plan_invalid', 'Publicação retida — plano de publicação inválido'],
    ['failed', 'Publicação retida — falha no processamento'],
  ])('W4 — hold/%s is an anomaly carrying hold_detail', (holdReason, title) => {
    const { release } = withRelease({
      ...HOLD_BASE,
      hold_reason: holdReason,
      hold_detail: 'generation 3 conflita com POP-GENERAL-015',
    });

    expect(release).toEqual({
      tone: 'anomaly',
      title,
      detail: 'generation 3 conflita com POP-GENERAL-015',
    });
  });

  it('W4 — hold not yet evaluated (hold_reason null) is transient, not an anomaly', () => {
    const { release } = withRelease({ ...HOLD_BASE, last_evaluated_at: null });

    expect(release).toEqual({ tone: 'progress', title: 'Preparando publicação…', detail: null });
  });

  it('W4 — release null (no generation / legacy row): no release presentation at all', () => {
    expect(withRelease(null).release).toBeNull();
  });

  // ── W5: lifecycle settling poll ────────────────────────────────────────────
  // Post-ADR-0085 a document sits at status='approved' while the coordinator works,
  // so the old status==='scheduled' gate alone froze the page on "Preparando
  // artefatos finais…" until a manual reload. ONE predicate
  // (`isDocumentLifecycleSettling`) drives both consumers: the detail query polls on
  // it internally (hence the `pollLifecycleUntilSettled` flag asserted below — that
  // hook is mocked here, so its resolved interval is covered by the predicate test)
  // and the adapter applies the same predicate to the sibling queries.

  /** Resolved refetchInterval the adapter handed to [activeDocument, revisionHistory]. */
  function siblingPollIntervals() {
    return [
      vi.mocked(useControlledDocumentActiveDocumentQuery).mock.calls.at(-1)?.[1]?.refetchInterval,
      vi.mocked(useDocumentRevisionHistoryQuery).mock.calls.at(-1)?.[1]?.refetchInterval,
    ];
  }

  it('W5 — the detail query is always asked to poll until the lifecycle settles', () => {
    withRelease({ ...HOLD_BASE, hold_reason: 'materializing' });

    expect(vi.mocked(useDocumentDetailQuery).mock.calls.at(-1)?.[1]).toEqual({ pollLifecycleUntilSettled: true });
  });

  it.each(['materializing', 'awaiting_approval_fact'])(
    'W5 — transient hold/%s keeps every lifecycle query polling at 5s',
    (holdReason) => {
      withRelease({ ...HOLD_BASE, hold_reason: holdReason });

      expect(siblingPollIntervals()).toEqual([5_000, 5_000]);
      expect(isDocumentLifecycleSettling({ status: 'approved', release: { ...HOLD_BASE, hold_reason: holdReason } as never })).toBe(true);
    },
  );

  it('W5 — a hold the coordinator has not evaluated yet (hold_reason null) keeps polling', () => {
    withRelease({ ...HOLD_BASE });

    expect(siblingPollIntervals()).toEqual([5_000, 5_000]);
  });

  it.each(['supersede_conflict', 'plan_invalid', 'failed'])(
    'W5 — terminal anomaly hold/%s does NOT poll (waiting never resolves it)',
    (holdReason) => {
      withRelease({ ...HOLD_BASE, hold_reason: holdReason, hold_detail: 'detalhe do operador' });

      expect(siblingPollIntervals()).toEqual([false, false]);
      expect(isDocumentLifecycleSettling({ status: 'approved', release: { ...HOLD_BASE, hold_reason: holdReason } as never })).toBe(false);
    },
  );

  it('W5 — awaiting_effective_date does NOT poll (date-bounded wait, not a transient hold)', () => {
    withRelease({
      ...HOLD_BASE,
      hold_reason: 'awaiting_effective_date',
      planned_effective_from: '2026-08-03T12:00:00.000Z',
    });

    expect(siblingPollIntervals()).toEqual([false, false]);
  });

  it('W5 — polling stops once the generation is released', () => {
    withRelease({ ...HOLD_BASE, state: 'released', released_at: '2026-05-19T12:00:00.000Z' });

    expect(siblingPollIntervals()).toEqual([false, false]);
    expect(
      isDocumentLifecycleSettling({
        status: 'published',
        release: { ...HOLD_BASE, state: 'released', released_at: '2026-05-19T12:00:00.000Z' } as never,
      }),
    ).toBe(false);
  });

  it('W5 — the legacy scheduled-status gate is unchanged (still polls at 5s)', () => {
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { ...BASE_DOC, status: 'scheduled', release: null },
      refetch: vi.fn(),
    } as never);
    renderHook(() => useDocumentArtifact('doc-published-1'));

    expect(siblingPollIntervals()).toEqual([5_000, 5_000]);
    expect(isDocumentLifecycleSettling({ status: 'scheduled', release: null })).toBe(true);
  });

  // ── W6: publication date source ────────────────────────────────────────────
  // The projection is the sole source of the release FACT; approval.completed_at is
  // approval-COMPLETION time and may only stand in for legacy rows with no generation.

  it('W6 — publication date comes from release.released_at, not approval completion', () => {
    const { model } = withRelease({
      ...HOLD_BASE,
      state: 'released',
      // Deliberately a different day from the approval completion (19 de maio).
      released_at: '2026-06-02T12:00:00.000Z',
    });

    expect(model.hero.subtitle).toBe('publicado em 02 de junho de 2026');
    expect(model.meta.ownerDescriptor).toBe('publicado em 02 de junho de 2026');
    expect(model.kpis.find((k) => k.key === 'currentVersion')?.hint).toBe('desde 02/06/2026');
  });

  it('W6 — legacy document with no release generation falls back to approval completion', () => {
    const { model } = withRelease(null);

    expect(model.hero.subtitle).toBe('publicado em 19 de maio de 2026');
    expect(model.kpis.find((k) => k.key === 'currentVersion')?.hint).toBe('desde 19/05/2026');
  });

  it('W6 — a held generation has NO publication date (approval timestamp is not relabelled)', () => {
    const { model } = withRelease({ ...HOLD_BASE, hold_reason: 'materializing' });

    expect(model.hero.subtitle).toBe('Publicado');
    expect(model.kpis.find((k) => k.key === 'currentVersion')?.hint).not.toMatch(/desde/);
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
