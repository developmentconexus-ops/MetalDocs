import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTemplateSchemas } from '../useTemplateSchemas';
import * as api from '../../api/templates';
import { StaleLockVersionError } from '../../api/templates';

vi.mock('../../api/templates', async () => {
  const actual = await vi.importActual<typeof api>('../../api/templates');
  return {
    ...actual,
    getTemplateSchemas: vi.fn(),
    putTemplateSchemas: vi.fn(),
  };
});

const schemas = {
  placeholders: [{ id: 'p1', label: 'Title', type: 'text' as const }],
  composition: null,
};

describe('useTemplateSchemas', () => {
  beforeEach(() => {
    vi.mocked(api.getTemplateSchemas).mockResolvedValue({ schemas, lockVersion: 0 });
    vi.mocked(api.putTemplateSchemas).mockResolvedValue({ lockVersion: 1 });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('loads schemas on mount', async () => {
    const { result } = renderHook(() => useTemplateSchemas('t1', 1));
    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.schemas).toEqual(schemas);
    expect(result.current.error).toBeNull();
    expect(result.current.staleConflict).toBe(false);
  });

  it('sets error on load failure', async () => {
    vi.mocked(api.getTemplateSchemas).mockRejectedValue(new Error('not found'));
    const { result } = renderHook(() => useTemplateSchemas('t1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe('not found');
    expect(result.current.schemas).toBeNull();
  });

  it('save sends current lockVersion and bumps it on success', async () => {
    vi.mocked(api.getTemplateSchemas).mockResolvedValue({ schemas, lockVersion: 4 });
    vi.mocked(api.putTemplateSchemas).mockResolvedValue({ lockVersion: 5 });

    const { result } = renderHook(() => useTemplateSchemas('t1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const updated = { ...schemas, placeholders: [] };
    await act(async () => {
      await result.current.save(updated);
    });

    expect(vi.mocked(api.putTemplateSchemas)).toHaveBeenCalledWith('t1', 1, updated, 4);
    expect(result.current.schemas).toEqual(updated);
    expect(result.current.staleConflict).toBe(false);

    // Second save uses bumped lock.
    await act(async () => {
      await result.current.save(updated);
    });
    expect(vi.mocked(api.putTemplateSchemas)).toHaveBeenLastCalledWith('t1', 1, updated, 5);
  });

  it('surfaces staleConflict on 412, does not overwrite local schemas, refetch clears it', async () => {
    vi.mocked(api.getTemplateSchemas).mockResolvedValueOnce({ schemas, lockVersion: 1 });
    vi.mocked(api.putTemplateSchemas).mockRejectedValueOnce(new StaleLockVersionError());

    const { result } = renderHook(() => useTemplateSchemas('t1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const updated = { ...schemas, placeholders: [] };
    await act(async () => {
      await expect(result.current.save(updated)).rejects.toBeInstanceOf(StaleLockVersionError);
    });

    expect(result.current.staleConflict).toBe(true);
    expect(result.current.schemas).toEqual(schemas); // not silently overwritten

    const serverSchemas = { ...schemas, placeholders: [{ id: 'p2', label: 'Other', type: 'text' as const }] };
    vi.mocked(api.getTemplateSchemas).mockResolvedValueOnce({ schemas: serverSchemas, lockVersion: 2 });

    await act(async () => {
      await result.current.refetch();
    });

    expect(result.current.staleConflict).toBe(false);
    expect(result.current.schemas).toEqual(serverSchemas);
  });

  it('saving flag is true during save', async () => {
    let resolve!: (v: { lockVersion: number }) => void;
    vi.mocked(api.putTemplateSchemas).mockReturnValue(
      new Promise<{ lockVersion: number }>((r) => {
        resolve = r;
      }),
    );

    const { result } = renderHook(() => useTemplateSchemas('t1', 1));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      void result.current.save(schemas);
    });
    expect(result.current.saving).toBe(true);

    await act(async () => {
      resolve({ lockVersion: 1 });
    });
    expect(result.current.saving).toBe(false);
  });
});
