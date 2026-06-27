import { describe, it, expect } from 'vitest';
import { partitionTokens, partitionDetected } from './tokens';
import type { DetectedToken } from '@metaldocs/shared-tokens';

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

describe('partitionDetected', () => {
  const dt = (name: string, valid: boolean): DetectedToken => ({ name, kind: 'var', valid, reserved: false });

  it('splits valid-in-catalog, valid-unknown, and invalid tokens', () => {
    const detected = [dt('doc_code', true), dt('nope', true), dt('a.b', false)];
    const r = partitionDetected(detected, new Set(['doc_code']));
    expect(r.usedKeys).toEqual(new Set(['doc_code']));
    expect(r.unknownTokens).toEqual(['nope']);
    expect(r.invalidTokens).toEqual(['a.b']);
  });
});
