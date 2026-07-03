import { describe, expect, it, vi } from 'vitest';

vi.mock('@tanstack/react-query', () => ({
  useQuery: vi.fn((options: unknown) => options),
}));

vi.mock('../../templates', () => ({
  listTemplates: vi.fn(),
}));

import { useQuery } from '@tanstack/react-query';
import { listTemplates } from '../../templates';
import { QK } from '../../../lib/queryKeys';
import { useTemplatesByProfileQuery } from './useTemplatesByProfileQuery';

describe('useTemplatesByProfileQuery', () => {
  it('requests only published templates and uses a queryKey distinct from the admin (all-status) list/byProfile keys', () => {
    const options = useTemplatesByProfileQuery('PRC') as unknown as {
      queryKey: readonly unknown[];
      queryFn: () => unknown;
    };

    expect(options.queryKey).toEqual(QK.templates.byProfilePublished('PRC'));
    expect(options.queryKey).not.toEqual(QK.templates.byProfile('PRC'));
    expect(options.queryKey).not.toEqual(QK.templates.list());

    options.queryFn();
    expect(listTemplates).toHaveBeenCalledWith({ doc_type: 'PRC', published: true });
  });

  it('falls back to the admin list key and stays disabled when profileCode is null', () => {
    const options = useQuery as unknown as ReturnType<typeof vi.fn>;
    useTemplatesByProfileQuery(null);
    const lastCall = options.mock.calls[options.mock.calls.length - 1][0];

    expect(lastCall.queryKey).toEqual(QK.templates.list());
    expect(lastCall.enabled).toBe(false);
  });
});
