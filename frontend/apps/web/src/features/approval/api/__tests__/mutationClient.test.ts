import { afterEach, describe, expect, it, vi } from 'vitest';

import { ApprovalError, mutate } from '../mutationClient';

describe('mutationClient problem+json handling', () => {
  afterEach(() => vi.restoreAllMocks());

  it('throws ApprovalError from an RFC 9457 problem response before legacy parsing', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValue(
      new Response(
        JSON.stringify({
          title: 'Validation failed',
          status: 422,
          code: 'VALIDATION_ERROR',
          detail: 'Payload is invalid.',
        }),
        {
          status: 422,
          headers: {
            'Content-Type': 'application/problem+json',
          },
        },
      ),
    );

    let caught: ApprovalError | undefined;

    try {
      await mutate('POST', '/api/test', {});
    } catch (error) {
      caught = error as ApprovalError;
    }

    expect(caught).toBeInstanceOf(ApprovalError);
    expect(caught?.status).toBe(422);
    expect(caught?.code).toBe('VALIDATION_ERROR');
    expect(caught?.message).toBe('Validation failed');
  });
});
