import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../../lib/api/errors';
import { api } from '../../../lib/api/client';
import {
  archiveArea,
  archiveProfile,
  createArea,
  createFamily,
  createProfile,
  deactivateFamily,
  fetchAreas,
  fetchFamilies,
  fetchProfiles,
  setDefaultTemplate,
  updateArea,
  updateFamily,
  updateProfile,
} from './taxonomy';

// Mocks the generated `api` client directly (openapi-fetch), matching the
// precedent in features/iam/queries/__tests__/useMembershipsQuery.test.tsx —
// openapi-fetch builds a `Request` from the configured baseUrl before the
// injected fetch ever runs, which needs an absolute origin `global.fetch`
// stubbing doesn't provide under jsdom/Node. Mocking at the `api` boundary
// instead lets these tests assert the real wire-shape (snake_case) fixtures
// exercise the generated-contract mapping in taxonomy.ts, without depending
// on environment URL resolution.
vi.mock('../../../lib/api/client', () => ({
  api: { GET: vi.fn(), POST: vi.fn(), PATCH: vi.fn(), PUT: vi.fn(), DELETE: vi.fn() },
}));

const mockGet = api.GET as unknown as ReturnType<typeof vi.fn>;
const mockPost = api.POST as unknown as ReturnType<typeof vi.fn>;
const mockPatch = api.PATCH as unknown as ReturnType<typeof vi.fn>;
const mockPut = api.PUT as unknown as ReturnType<typeof vi.fn>;
const mockDelete = api.DELETE as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockGet.mockReset();
  mockPost.mockReset();
  mockPatch.mockReset();
  mockPut.mockReset();
  mockDelete.mockReset();
});

describe('fetchProfiles', () => {
  it('maps DocumentProfileItem (snake_case wire) to DocumentProfile (camelCase app shape)', async () => {
    mockGet.mockResolvedValue({
      data: {
        items: [
          {
            code: 'pop',
            family_code: 'sop',
            name: 'Procedimento Operacional',
            description: 'desc',
            review_interval_days: 365,
            active_schema_version: 1,
            workflow_profile: 'editor',
            approval_required: true,
            retention_days: 30,
            validity_days: 365,
          },
        ],
      },
      error: undefined,
    });

    const profiles = await fetchProfiles();

    expect(profiles).toHaveLength(1);
    expect(profiles[0]).toMatchObject({
      code: 'pop',
      familyCode: 'sop',
      name: 'Procedimento Operacional',
      description: 'desc',
      reviewIntervalDays: 365,
      editableByRole: 'editor',
    });
  });

  it('maps governance_class (wire) to governanceClass (app shape)', async () => {
    mockGet.mockResolvedValue({
      data: {
        items: [
          {
            code: 'pop',
            family_code: 'sop',
            name: 'Procedimento Operacional',
            description: 'desc',
            review_interval_days: 365,
            active_schema_version: 1,
            workflow_profile: 'editor',
            approval_required: true,
            retention_days: 30,
            validity_days: 365,
            governance_class: 'simples',
          },
        ],
      },
      error: undefined,
    });

    const profiles = await fetchProfiles();

    expect(profiles[0].governanceClass).toBe('simples');
  });

  it('returns an empty array when items is absent', async () => {
    mockGet.mockResolvedValue({ data: {}, error: undefined });
    const profiles = await fetchProfiles();
    expect(profiles).toEqual([]);
  });

  it('sends include_archived=true when requested', async () => {
    mockGet.mockResolvedValue({ data: { items: [] }, error: undefined });
    await fetchProfiles(true);
    expect(mockGet).toHaveBeenCalledWith('/taxonomy/profiles', {
      params: { query: { include_archived: true } },
    });
  });

  it('omits the query when includeArchived is not requested', async () => {
    mockGet.mockResolvedValue({ data: { items: [] }, error: undefined });
    await fetchProfiles();
    expect(mockGet).toHaveBeenCalledWith('/taxonomy/profiles', { params: { query: undefined } });
  });

  it('throws ApiError when the client reports an error', async () => {
    mockGet.mockResolvedValue({ data: undefined, error: { code: 'authz.capability_denied' } });
    await expect(fetchProfiles()).rejects.toBeInstanceOf(ApiError);
  });
});

describe('createProfile / updateProfile', () => {
  it('creates a profile and maps the response', async () => {
    mockPost.mockResolvedValue({
      data: {
        code: 'pop',
        family_code: 'sop',
        name: 'Procedimento Operacional',
        description: '',
        review_interval_days: 180,
        active_schema_version: 1,
        workflow_profile: 'admin',
        approval_required: false,
        retention_days: 0,
        validity_days: 0,
      },
      error: undefined,
    });

    const profile = await createProfile({
      code: 'pop',
      familyCode: 'sop',
      name: 'Procedimento Operacional',
      reviewIntervalDays: 180,
    });

    expect(profile.code).toBe('pop');
    expect(profile.reviewIntervalDays).toBe(180);
    expect(mockPost).toHaveBeenCalledWith('/taxonomy/profiles', {
      params: { header: { 'Idempotency-Key': expect.any(String) } },
      body: {
        code: 'pop',
        family_code: 'sop',
        name: 'Procedimento Operacional',
        description: '',
        review_interval_days: 180,
        editable_by_role: 'admin',
      },
    });
  });

  it('updates a profile against the {code} path', async () => {
    mockPatch.mockResolvedValue({
      data: {
        code: 'pop',
        family_code: 'sop',
        name: 'Novo nome',
        description: '',
        review_interval_days: 90,
        active_schema_version: 1,
        workflow_profile: 'admin',
        approval_required: false,
        retention_days: 0,
        validity_days: 0,
      },
      error: undefined,
    });

    const profile = await updateProfile('pop', { familyCode: 'sop', name: 'Novo nome', reviewIntervalDays: 90 });

    expect(profile.name).toBe('Novo nome');
    expect(mockPatch).toHaveBeenCalledWith(
      '/taxonomy/profiles/{code}',
      expect.objectContaining({ params: { path: { code: 'pop' } } }),
    );
  });

  it('throws ApiError with an empty-response fallback when the server sends no body', async () => {
    mockPost.mockResolvedValue({ data: undefined, error: undefined });
    await expect(
      createProfile({ code: 'pop', familyCode: 'sop', name: 'x', reviewIntervalDays: 1 }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});

describe('setDefaultTemplate', () => {
  it('sends the wire field name (template_version_id) to the default-template endpoint', async () => {
    mockPut.mockResolvedValue({ data: undefined, error: undefined });
    await setDefaultTemplate('pop', { templateVersionId: 'tv-1' });

    expect(mockPut).toHaveBeenCalledWith('/taxonomy/profiles/{code}/default-template', {
      params: { path: { code: 'pop' } },
      body: { template_version_id: 'tv-1' },
    });
  });
});

describe('archiveProfile', () => {
  it('resolves when the client reports no error', async () => {
    mockDelete.mockResolvedValue({ data: undefined, error: undefined });
    await expect(archiveProfile('pop')).resolves.toBeUndefined();
    expect(mockDelete).toHaveBeenCalledWith('/taxonomy/profiles/{code}', { params: { path: { code: 'pop' } } });
  });

  it('throws ApiError when the client reports an error', async () => {
    mockDelete.mockResolvedValue({ data: undefined, error: { code: 'conflict.profile_referenced' } });
    await expect(archiveProfile('pop')).rejects.toBeInstanceOf(ApiError);
  });
});

describe('fetchAreas', () => {
  it('maps ProcessAreaItem (snake_case wire) to ProcessArea (camelCase app shape)', async () => {
    mockGet.mockResolvedValue({
      data: {
        items: [
          {
            code: 'rh',
            tenant_id: 'tenant-1',
            name: 'Recursos Humanos',
            description: 'desc',
            parent_code: null,
            owner_user_id: null,
            default_approver_role: 'approver',
            archived_at: null,
            created_at: '2026-05-19T10:00:00Z',
          },
        ],
      },
      error: undefined,
    });

    const areas = await fetchAreas();

    expect(areas).toHaveLength(1);
    expect(areas[0]).toMatchObject({
      code: 'rh',
      tenantId: 'tenant-1',
      name: 'Recursos Humanos',
      parentCode: null,
      defaultApproverRole: 'approver',
      archivedAt: null,
    });
  });
});

describe('createArea / updateArea / archiveArea', () => {
  it('creates an area', async () => {
    mockPost.mockResolvedValue({
      data: {
        code: 'fin',
        tenant_id: 'tenant-1',
        name: 'Financeiro',
        description: '',
        parent_code: null,
        owner_user_id: null,
        default_approver_role: null,
        archived_at: null,
        created_at: '2026-05-19T10:00:00Z',
      },
      error: undefined,
    });
    const area = await createArea({ code: 'fin', name: 'Financeiro' });
    expect(area.code).toBe('fin');
  });

  it('updates an area', async () => {
    mockPut.mockResolvedValue({
      data: {
        code: 'fin',
        tenant_id: 'tenant-1',
        name: 'Financeiro Global',
        description: '',
        parent_code: null,
        owner_user_id: null,
        default_approver_role: null,
        archived_at: null,
        created_at: '2026-05-19T10:00:00Z',
      },
      error: undefined,
    });
    const area = await updateArea('fin', { name: 'Financeiro Global' });
    expect(area.name).toBe('Financeiro Global');
  });

  it('archives an area', async () => {
    mockDelete.mockResolvedValue({ data: undefined, error: undefined });
    await expect(archiveArea('fin')).resolves.toBeUndefined();
  });
});

describe('fetchFamilies', () => {
  it('maps DocumentFamilyItem (snake_case wire) to DocumentFamily (camelCase app shape)', async () => {
    mockGet.mockResolvedValue({
      data: {
        items: [
          {
            code: 'sop',
            tenant_id: 'tenant-1',
            name: 'Standard Operating Procedure',
            description: 'desc',
            is_active: true,
            created_at: '2026-05-19T10:00:00Z',
          },
        ],
      },
      error: undefined,
    });

    const families = await fetchFamilies();

    expect(families).toHaveLength(1);
    expect(families[0]).toMatchObject({
      code: 'sop',
      name: 'Standard Operating Procedure',
      isActive: true,
    });
  });

  it('sends include_inactive=true when requested', async () => {
    mockGet.mockResolvedValue({ data: { items: [] }, error: undefined });
    await fetchFamilies(true);
    expect(mockGet).toHaveBeenCalledWith('/taxonomy/families', {
      params: { query: { include_inactive: true } },
    });
  });
});

describe('createFamily / updateFamily / deactivateFamily', () => {
  it('creates a family', async () => {
    mockPost.mockResolvedValue({
      data: {
        code: 'sop',
        tenant_id: 'tenant-1',
        name: 'SOP',
        description: '',
        is_active: true,
        created_at: '2026-05-19T10:00:00Z',
      },
      error: undefined,
    });
    const family = await createFamily({ code: 'sop', name: 'SOP' });
    expect(family.code).toBe('sop');
  });

  it('updates a family', async () => {
    mockPatch.mockResolvedValue({
      data: {
        code: 'sop',
        tenant_id: 'tenant-1',
        name: 'SOP Renamed',
        description: '',
        is_active: true,
        created_at: '2026-05-19T10:00:00Z',
      },
      error: undefined,
    });
    const family = await updateFamily('sop', { name: 'SOP Renamed' });
    expect(family.name).toBe('SOP Renamed');
  });

  it('deactivates a family', async () => {
    mockDelete.mockResolvedValue({ data: undefined, error: undefined });
    await expect(deactivateFamily('sop')).resolves.toBeUndefined();
  });
});
