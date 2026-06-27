export function partitionTokens(
  usedTokenKeys: string[],
  catalogKeys: Set<string>,
): { usedKeys: Set<string>; unknownTokens: string[] } {
  const usedKeys = new Set<string>();
  const unknownTokens: string[] = [];
  const seenUnknown = new Set<string>();
  for (const key of usedTokenKeys) {
    if (catalogKeys.has(key)) {
      usedKeys.add(key);
    } else if (!seenUnknown.has(key)) {
      seenUnknown.add(key);
      unknownTokens.push(key);
    }
  }
  return { usedKeys, unknownTokens };
}

import type { DetectedToken } from '@metaldocs/shared-tokens';

/**
 * Partition broad-detected tokens for the authoring panel:
 * - usedKeys: valid tokens present in the catalog (the "in use" set)
 * - unknownTokens: valid tokens NOT in the catalog (typo / not-yet-defined)
 * - invalidTokens: charset-invalid tokens freeze cannot fill (e.g. {a.b})
 */
export function partitionDetected(
  detected: DetectedToken[],
  catalogKeys: Set<string>,
): { usedKeys: Set<string>; unknownTokens: string[]; invalidTokens: string[] } {
  const usedKeys = new Set<string>();
  const unknownTokens: string[] = [];
  const invalidTokens: string[] = [];
  const seenUnknown = new Set<string>();
  const seenInvalid = new Set<string>();
  for (const d of detected) {
    if (!d.valid) {
      if (!seenInvalid.has(d.name)) { seenInvalid.add(d.name); invalidTokens.push(d.name); }
      continue;
    }
    if (catalogKeys.has(d.name)) {
      usedKeys.add(d.name);
    } else if (!seenUnknown.has(d.name)) {
      seenUnknown.add(d.name);
      unknownTokens.push(d.name);
    }
  }
  return { usedKeys, unknownTokens, invalidTokens };
}
