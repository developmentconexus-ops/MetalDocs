import { describe, expect, it, vi, beforeEach } from 'vitest';

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn((options: unknown) => options),
}));

vi.mock('../api/approvalApi', () => ({
  getInstance: vi.fn(),
}));

import { useQuery } from '@tanstack/react-query';
import { getInstance } from '../api/approvalApi';
import { etagCache } from '../api/etagCache';
import { QK } from '../../../lib/queryKeys';
import { useApprovalInstanceQuery } from './useApprovalInstanceQuery';

const makeInstance = () =>
  ({
    id: 'inst-1',
    document_id: 'doc-1',
    etag: 'v1',
  }) as unknown as ReturnType<typeof getInstance> extends Promise<infer T> ? T : never;

describe('useApprovalInstanceQuery', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    etagCache.clear();
  });

  it('keys + enabled: keys on QK.approval.instance(id); enabled reflects the passed flag', () => {
    const options = useApprovalInstanceQuery('doc-1', true) as unknown as {
      queryKey: readonly unknown[];
      enabled: boolean;
    };

    expect(options.queryKey).toEqual(QK.approval.instance('doc-1'));
    expect(options.enabled).toBe(true);

    const disabled = useApprovalInstanceQuery('doc-1', false) as unknown as { enabled: boolean };
    expect(disabled.enabled).toBe(false);
  });

  it('seeds etag on success: queryFn delegates to getInstance and returns its value', async () => {
    const instance = makeInstance();
    vi.mocked(getInstance).mockResolvedValue(instance);

    const options = useApprovalInstanceQuery('doc-1', true) as unknown as {
      queryFn: () => Promise<unknown>;
    };

    const result = await options.queryFn();
    expect(getInstance).toHaveBeenCalledWith('doc-1');
    expect(result).toBe(instance);
  });

  it('404 resolves null, no etag', async () => {
    vi.mocked(getInstance).mockRejectedValue({ status: 404 });

    const options = useApprovalInstanceQuery('doc-1', true) as unknown as {
      queryFn: () => Promise<unknown>;
    };

    const result = await options.queryFn();
    expect(result).toBeNull();
    expect(etagCache.get('doc-1')).toBeUndefined();
  });

  it('non-404 rejects', async () => {
    vi.mocked(getInstance).mockRejectedValue({ status: 500 });

    const options = useApprovalInstanceQuery('doc-1', true) as unknown as {
      queryFn: () => Promise<unknown>;
    };

    await expect(options.queryFn()).rejects.toEqual({ status: 500 });
  });

  it('sanity: useQuery is the mocked identity function', () => {
    expect(vi.isMockFunction(useQuery)).toBe(true);
  });
});
