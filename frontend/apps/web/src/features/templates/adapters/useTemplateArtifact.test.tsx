import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';

vi.mock('../../../store/auth.store', () => ({
  useAuthStore: (selector: (state: { user: { userId: string; displayName: string; roles: string[] } }) => unknown) =>
    selector({
      user: { userId: 'admin-user', displayName: 'Administrator', roles: ['system_admin'] },
    }),
}));

vi.mock('../queries/useTemplateDetailQuery', () => ({
  useTemplateDetailQuery: vi.fn(),
}));

import { useTemplateDetailQuery } from '../queries/useTemplateDetailQuery';
import { useTemplateArtifact } from './useTemplateArtifact';
import { formatRevisionCode } from '../../../lib/labels/revisionCode';

const BASE_TEMPLATE = {
  id: 'tpl-1',
  tenant_id: 'tenant-1',
  doc_type_code: 'POP',
  key: 'pop-onboarding-v1',
  name: 'Procedimento de Onboarding',
  description: 'Template de onboarding para novos colaboradores',
  latest_version: 3,
  latest_revision_number: 3,
  published_version_id: 'ver-2',
  published_version_number: 2,
  current_revision_number: 2,
  created_by: 'admin-user',
  created_at: '2026-01-15T10:00:00.000Z',
  archived_at: null,
};

const BASE_VERSION = {
  id: 'ver-3',
  template_id: 'tpl-1',
  version_number: 3,
  revision_number: 3,
  status: 'draft' as const,
  created_at: '2026-06-01T08:00:00.000Z',
};

describe('useTemplateArtifact', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: { template: BASE_TEMPLATE, latest_version: BASE_VERSION },
      refetch: vi.fn(),
    } as never);
  });

  it('maps the template detail into a kind=template view-model with null code', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.kind).toBe('template');
    expect(model.id).toBe('tpl-1');
    expect(model.code).toBeNull();
  });

  it('maps title, status, and revisionLabel from the template and version', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.title).toBe('Procedimento de Onboarding');
    expect(model.status).toBe('draft');
    expect(model.revisionLabel).toBe(formatRevisionCode(BASE_VERSION.revision_number));
  });

  it('sets approvalChain to null (templates have no approval-instance model)', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.approvalChain).toBeNull();
  });

  it('exposes a single documento tab and an empty actions array', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.tabs).toHaveLength(1);
    expect(model.tabs[0].key).toBe('documento');
    expect(model.actions).toEqual([]);
  });

  it('includes a key badge with label equal to template.key', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const keyBadge = model.hero.badges.find((b) => b.key === 'key');
    expect(keyBadge).toBeDefined();
    expect(keyBadge?.label).toBe(BASE_TEMPLATE.key);
    expect(keyBadge?.variant).toBe('code');
  });

  it('includes a status badge with revision label and pt-BR status text', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const statusBadge = model.hero.badges.find((b) => b.key === 'status');
    expect(statusBadge).toBeDefined();
    expect(statusBadge?.label).toContain(formatRevisionCode(BASE_VERSION.revision_number));
    expect(statusBadge?.label).toContain('rascunho');
    expect(statusBadge?.variant).toBe('status');
  });

  it('includes a type badge with doc_type_code when present', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const typeBadge = model.hero.badges.find((b) => b.key === 'type');
    expect(typeBadge).toBeDefined();
    expect(typeBadge?.label).toBe('POP');
    expect(typeBadge?.variant).toBe('type');
  });

  it('builds hero breadcrumb with Modelos root and template name', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.hero.breadcrumb).toEqual([
      { label: 'Modelos', href: '/templates' },
      { label: 'Procedimento de Onboarding' },
    ]);
  });

  it('sets hero.subtitle to null', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.hero.subtitle).toBeNull();
  });

  it('sets ownerName to displayName when created_by matches current user', () => {
    // BASE_TEMPLATE.created_by === 'admin-user' which matches mocked userId
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.meta.ownerName).toBe('Administrator');
  });

  it('sets ownerName to the raw created_by when creator is a different user', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        template: { ...BASE_TEMPLATE, created_by: 'other-user' },
        latest_version: BASE_VERSION,
      },
      refetch: vi.fn(),
    } as never);

    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.meta.ownerName).toBe('other-user');
  });

  it('sets meta.profileLabel from doc_type_code and leaves area/visibility/file meta null', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.meta.profileLabel).toBe('POP');
    expect(model.meta.areaLabel).toBeNull();
    expect(model.meta.visibilityLabel).toBeNull();
    expect(model.meta.fileSizeBytes).toBeNull();
    expect(model.meta.pageCount).toBeNull();
    expect(model.meta.effectiveFrom).toBeNull();
    expect(model.meta.nextReviewAt).toBeNull();
    expect(model.meta.ownerDescriptor).toBeNull();
  });

  it('emits an empty lineage array (no version-list endpoint exists yet)', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    expect(model.lineage).toEqual([]);
  });

  it('returns currentVersion KPI with revision label and pt-BR status hint', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const cell = model.kpis.find((k) => k.key === 'currentVersion');
    expect(cell?.value).toBe(formatRevisionCode(BASE_VERSION.revision_number));
    expect(cell?.hint).toBe('rascunho');
  });

  it('returns publishedVersion KPI with published revision label when published_version_id is set', () => {
    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const cell = model.kpis.find((k) => k.key === 'publishedVersion');
    expect(cell?.value).toBe(formatRevisionCode(BASE_TEMPLATE.current_revision_number));
    expect(cell?.hint).toBe('vigente');
  });

  it('returns publishedVersion KPI with EM_DASH and "não publicada" when no published_version_id', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: {
        template: { ...BASE_TEMPLATE, published_version_id: null, current_revision_number: null },
        latest_version: BASE_VERSION,
      },
      refetch: vi.fn(),
    } as never);

    const { model } = renderHook(() => useTemplateArtifact('tpl-1')).result.current;

    const cell = model.kpis.find((k) => k.key === 'publishedVersion');
    expect(cell?.value).toBe('—');
    expect(cell?.hint).toBe('não publicada');
  });

  it('surfaces isLoading from the query', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: true,
      isError: false,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    const result = renderHook(() => useTemplateArtifact('tpl-1')).result.current;
    expect(result.isLoading).toBe(true);
  });

  it('surfaces isError when the query errors', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: false,
      isError: true,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    const result = renderHook(() => useTemplateArtifact('tpl-1')).result.current;
    expect(result.isError).toBe(true);
  });

  it('surfaces isError when data is absent (no error flag needed)', () => {
    vi.mocked(useTemplateDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: undefined,
      refetch: vi.fn(),
    } as never);

    const result = renderHook(() => useTemplateArtifact('tpl-1')).result.current;
    expect(result.isError).toBe(true);
  });
});
