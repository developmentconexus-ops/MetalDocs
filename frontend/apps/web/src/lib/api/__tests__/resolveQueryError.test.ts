import { describe, expect, it } from 'vitest';
import { ApiError } from '../errors';
import { resolveQueryError } from '../resolveQueryError';

describe('resolveQueryError', () => {
  it('returns resolveErrorMessage result for ApiError', () => {
    const err = new ApiError('NOT_FOUND', 404, 'not found');
    const result = resolveQueryError(err, 'fallback');
    // resolveErrorMessage maps known codes to Portuguese; unknown codes use the backendMessage.
    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
    expect(result).not.toBe('fallback');
  });

  it('returns .message for a plain Error', () => {
    const err = new Error('something went wrong');
    expect(resolveQueryError(err, 'fallback')).toBe('something went wrong');
  });

  it('returns fallback for unknown/non-Error values', () => {
    expect(resolveQueryError(null, 'fallback')).toBe('fallback');
    expect(resolveQueryError(undefined, 'fallback')).toBe('fallback');
    expect(resolveQueryError('string error', 'fallback')).toBe('fallback');
    expect(resolveQueryError(42, 'fallback')).toBe('fallback');
  });
});
