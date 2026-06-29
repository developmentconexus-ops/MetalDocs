import type { PlaceholderCatalogEntry } from '../api/catalog';
import type { TokenDictionaryEntry } from '../../tokens/api/tokensTypes';

export type TokenKind = 'computed' | 'dictionary';

export interface Token {
  key: string; // the {key} placeholder
  label: string;
  description?: string;
  kind: TokenKind;
}

/** Logic-free composition: pass each source through, tag with its constant kind.
 * No truth is invented here — keys/labels stay backend-authoritative. */
export function toUnifiedTokens(
  computed: PlaceholderCatalogEntry[],
  dictionary: TokenDictionaryEntry[],
): Token[] {
  return [
    ...computed.map((c) => ({
      key: c.key,
      label: c.label,
      description: c.description,
      kind: 'computed' as const,
    })),
    ...dictionary.map((d) => ({
      key: d.name,
      label: d.label,
      description: d.description ?? undefined,
      kind: 'dictionary' as const,
    })),
  ];
}
