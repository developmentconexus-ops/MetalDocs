import { renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../queries/useAreasQuery', () => ({ useAreasQuery: vi.fn() }));
vi.mock('../../taxonomy/queries/useProfilesQuery', () => ({ useProfilesQuery: vi.fn() }));
vi.mock('../../controlled-documents/queries/useControlledDocumentDetailQuery', () => ({
  useControlledDocumentDetailQuery: vi.fn(),
}));

import { useAreasQuery } from '../queries/useAreasQuery';
import { useProfilesQuery } from '../../taxonomy/queries/useProfilesQuery';
import { useControlledDocumentDetailQuery } from '../../controlled-documents/queries/useControlledDocumentDetailQuery';
import { useDocumentIdentityMeta } from './useDocumentIdentityMeta';
import type { DocumentDetail } from '../api/documents';

function asQuery<T>(value: unknown): T {
  return value as T;
}

const BASE_DOC = {
  id: 'doc-1',
  code: 'IT-USINAGEM-001',
  name: 'IT Usinagem',
  status: 'draft',
  controlled_document_id: 'cd-1',
  profile_code_snapshot: 'it',
  process_area_code_snapshot: 'usinagem',
  current_revision_page_count: 4,
  current_revision_file_size_bytes: 1304,
} as unknown as DocumentDetail;

function mockQueries(
  overrides: { areas?: unknown; profiles?: unknown; controlledDocument?: unknown } = {},
) {
  vi.mocked(useAreasQuery).mockReturnValue(
    asQuery(overrides.areas ?? { data: [{ code: 'usinagem', name: 'Usinagem' }], isError: false }),
  );
  vi.mocked(useProfilesQuery).mockReturnValue(
    asQuery(overrides.profiles ?? { data: [{ code: 'it', name: 'Instrução de Trabalho' }], isError: false }),
  );
  vi.mocked(useControlledDocumentDetailQuery).mockReturnValue(
    asQuery(
      overrides.controlledDocument ?? {
        data: { visibility: { scope: 'company', area_codes: [], user_ids: [] } },
        isError: false,
      },
    ),
  );
}

describe('useDocumentIdentityMeta', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockQueries();
  });

  it('resolves profile, area and visibility labels from the query data', () => {
    const { result } = renderHook(() => useDocumentIdentityMeta(BASE_DOC));

    expect(result.current.profileLabel).toBe('Instrução de Trabalho');
    expect(result.current.areaLabel).toBe('Usinagem');
    expect(result.current.visibilityLabel).toBe('Toda empresa');
    expect(result.current.absenceReasons).toEqual({});
  });

  it('degrades to the raw snapshot codes when the taxonomy lists have not resolved', () => {
    mockQueries({ areas: { data: undefined, isError: false }, profiles: { data: undefined, isError: false } });
    const { result } = renderHook(() => useDocumentIdentityMeta(BASE_DOC));

    expect(result.current.profileLabel).toBe('it');
    expect(result.current.areaLabel).toBe('usinagem');
    expect(result.current.absenceReasons.profileLabel).toBeUndefined();
  });

  it('explains an unlinked controlled document instead of blanking visibility', () => {
    mockQueries({ controlledDocument: { data: undefined, isError: false } });
    const doc = { ...BASE_DOC, controlled_document_id: null } as DocumentDetail;
    const { result } = renderHook(() => useDocumentIdentityMeta(doc));

    expect(result.current.visibilityLabel).toBeNull();
    expect(result.current.absenceReasons.visibilityLabel).toMatch(/ainda não está vinculada a um/i);
  });

  it('distinguishes a failed controlled-document read from a pending one', () => {
    mockQueries({ controlledDocument: { data: undefined, isError: true } });
    const { result } = renderHook(() => useDocumentIdentityMeta(BASE_DOC));
    expect(result.current.absenceReasons.visibilityLabel).toMatch(/Não foi possível carregar/i);

    mockQueries({ controlledDocument: { data: undefined, isError: false } });
    const pending = renderHook(() => useDocumentIdentityMeta(BASE_DOC));
    expect(pending.result.current.absenceReasons.visibilityLabel).toMatch(/Carregando/i);
  });

  it('explains absent snapshots and an absent page count', () => {
    const doc = {
      ...BASE_DOC,
      profile_code_snapshot: null,
      process_area_code_snapshot: null,
      current_revision_page_count: null,
    } as DocumentDetail;
    const { result } = renderHook(() => useDocumentIdentityMeta(doc));

    expect(result.current.profileLabel).toBeNull();
    expect(result.current.areaLabel).toBeNull();
    expect(result.current.absenceReasons.profileLabel).toMatch(/perfil registrado/i);
    expect(result.current.absenceReasons.areaLabel).toMatch(/área responsável/i);
    expect(result.current.absenceReasons.pageCount).toMatch(/contagem de páginas/i);
  });

  it('returns no page-count explanation when there is no document at all', () => {
    const { result } = renderHook(() => useDocumentIdentityMeta(null));
    expect(result.current.absenceReasons.pageCount).toBeUndefined();
  });
});
