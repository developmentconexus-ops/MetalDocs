import { describe, it, expect } from 'vitest';
import { partitionTokens } from './tokens';

describe('partitionTokens', () => {
  it('splits used tokens into catalog-known (used) and unknown', () => {
    const catalogKeys = new Set(['doc_code', 'author', 'effective_date']);
    const result = partitionTokens(['doc_code', 'nope', 'author', 'nope'], catalogKeys);
    expect(result.usedKeys).toEqual(new Set(['doc_code', 'author']));
    expect(result.unknownTokens).toEqual(['nope']); // de-duplicated, order preserved
  });

  it('returns empty sets for an empty document', () => {
    const result = partitionTokens([], new Set(['doc_code']));
    expect(result.usedKeys.size).toBe(0);
    expect(result.unknownTokens).toEqual([]);
  });
});
