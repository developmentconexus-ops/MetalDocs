import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTemplateSchemas } from '../useTemplateSchemas';
import * as api from '../../api/templates';
import { StaleLockVersionError, type VersionDTO } from '../../api/templates';

vi.mock('../../api/templates', async () => {
  const actual = await vi.importActual<typeof api>('../../api/templates');
  return {
    ...actual, // deriveTemplateSchemas is pure — exercise the real projection.
    putTemplateSchemas: vi.fn(),
  };
});

// Minimal version carrying just the fields the schema projection reads.
function versionFixture(over: Partial<VersionDTO> = {}): VersionDTO {
  return {
    placeholder_schema: [{ id: 'p1', label: 'Title', type: 'text', required: false }],
    lock_version: 0,
    ...over,
  } as unknown as VersionDTO;
}

const expectedSchemas = {
  placeholders: [{ id: 'p1', label: 'Title', type: 'text' as const }],
  composition: null,
};

describe('useTemplateSchemas', () => {
  beforeEach(() => {
    vi.mocked(api.putTemplateSchemas).mockResolvedValue({ lockVersion: 1 });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('is loading until a version is provided, then derives schemas from it', async () => {
    const { result, rerender } = renderHook(
      ({ v }: { v: VersionDTO | null }) => useTemplateSchemas('t1', 1, v),
      { initialProps: { v: null as VersionDTO | null } },
    );
    expect(result.current.loading).toBe(true);
    expect(result.current.schemas).toBeNull();

    rerender({ v: versionFixture() });
    expect(result.current.loading).toBe(false);
    expect(result.current.schemas).toEqual(expectedSchemas);
    expect(result.current.error).toBeNull();
    expect(result.current.staleConflict).toBe(false);
  });

  it('projects an empty placeholder list without throwing (regression: flat body, no { data } envelope)', () => {
    // Stable identity across renders, mirroring draft.version (React state).
    const v = versionFixture({ placeholder_schema: [] });
    const { result } = renderHook(() => useTemplateSchemas('t1', 1, v));
    expect(result.current.schemas).toEqual({ placeholders: [], composition: null });
    expect(result.current.error).toBeNull();
  });

  it('save sends the version lockVersion and bumps it on success', async () => {
    vi.mocked(api.putTemplateSchemas).mockResolvedValue({ lockVersion: 5 });

    const v = versionFixture({ lock_version: 4 });
    const { result } = renderHook(() => useTemplateSchemas('t1', 1, v));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const updated = { ...expectedSchemas, placeholders: [] };
    await act(async () => {
      await result.current.save(updated);
    });
    expect(vi.mocked(api.putTemplateSchemas)).toHaveBeenCalledWith('t1', 1, updated, 4);
    expect(result.current.staleConflict).toBe(false);

    // Second save uses the bumped lock.
    await act(async () => {
      await result.current.save(updated);
    });
    expect(vi.mocked(api.putTemplateSchemas)).toHaveBeenLastCalledWith('t1', 1, updated, 5);
  });

  it('surfaces staleConflict on 412 without overwriting schemas; a reloaded version clears it', async () => {
    vi.mocked(api.putTemplateSchemas).mockRejectedValueOnce(new StaleLockVersionError());

    const { result, rerender } = renderHook(
      ({ v }: { v: VersionDTO }) => useTemplateSchemas('t1', 1, v),
      { initialProps: { v: versionFixture({ lock_version: 1 }) } },
    );
    await waitFor(() => expect(result.current.loading).toBe(false));

    const updated = { ...expectedSchemas, placeholders: [] };
    await act(async () => {
      await expect(result.current.save(updated)).rejects.toBeInstanceOf(StaleLockVersionError);
    });

    expect(result.current.staleConflict).toBe(true);
    // Schemas reflect server truth (the loaded version), never the rejected save.
    expect(result.current.schemas).toEqual(expectedSchemas);

    // Parent reloads the draft → fresh version re-seeds lock + clears the conflict.
    const serverVersion = versionFixture({
      lock_version: 2,
      placeholder_schema: [{ id: 'p2', label: 'Other', type: 'text', required: false }],
    });
    act(() => rerender({ v: serverVersion }));

    expect(result.current.staleConflict).toBe(false);
    expect(result.current.schemas).toEqual({
      placeholders: [{ id: 'p2', label: 'Other', type: 'text' }],
      composition: null,
    });
  });

  it('saving flag is true during save', async () => {
    let resolve!: (v: { lockVersion: number }) => void;
    vi.mocked(api.putTemplateSchemas).mockReturnValue(
      new Promise<{ lockVersion: number }>((r) => {
        resolve = r;
      }),
    );

    const v = versionFixture();
    const { result } = renderHook(() => useTemplateSchemas('t1', 1, v));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      void result.current.save(expectedSchemas);
    });
    expect(result.current.saving).toBe(true);

    await act(async () => {
      resolve({ lockVersion: 1 });
    });
    expect(result.current.saving).toBe(false);
  });
});
