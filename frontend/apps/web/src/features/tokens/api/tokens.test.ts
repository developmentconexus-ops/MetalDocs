import { afterEach, describe, expect, it, vi } from 'vitest';
import * as apiModule from '../../../lib/api';
import { listTokens, createToken, updateToken, deleteToken } from './tokens';

afterEach(() => vi.restoreAllMocks());

describe('tokens api', () => {
  it('listTokens GETs /api/v1/tokens and returns items', async () => {
    const spy = vi
      .spyOn(apiModule, 'apiFetch')
      .mockResolvedValue({ items: [{ id: '1', name: 'slogan', value: 'v', label: 'L', created_at: 'x', updated_at: 'y' }] } as never);
    const items = await listTokens();
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens', undefined);
    expect(items).toHaveLength(1);
  });

  it('createToken POSTs the body', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue({ id: '1' } as never);
    await createToken({ name: 'slogan', value: 'v', label: 'L' });
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens', {
      method: 'POST',
      body: JSON.stringify({ name: 'slogan', value: 'v', label: 'L' }),
    });
  });

  it('updateToken PUTs to /api/v1/tokens/{id}', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue({ id: '1' } as never);
    await updateToken('1', { name: 'slogan', value: 'v2', label: 'L' });
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens/1', {
      method: 'PUT',
      body: JSON.stringify({ name: 'slogan', value: 'v2', label: 'L' }),
    });
  });

  it('deleteToken DELETEs /api/v1/tokens/{id}', async () => {
    const spy = vi.spyOn(apiModule, 'apiFetch').mockResolvedValue(undefined as never);
    await deleteToken('1');
    expect(spy).toHaveBeenCalledWith('/api/v1/tokens/1', { method: 'DELETE' });
  });
});
