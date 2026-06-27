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
