import { useMemo } from 'react';
import { usePlaceholderCatalogQuery } from '../queries/usePlaceholderCatalogQuery';
import { useTokensQuery } from '../../tokens/queries/useTokensQuery';
import { toUnifiedTokens, type Token } from './tokenCatalog';

/** Unified authoring token catalog: computed (render/domain via templates) +
 * dictionary (tokens module). Composed on the client; see ADR 0050 §promotion. */
export function useTokenCatalog(): {
  tokens: Token[];
  computedFailed: boolean;
  dictionaryFailed: boolean;
} {
  const computedQ = usePlaceholderCatalogQuery();
  const dictionaryQ = useTokensQuery();
  const tokens = useMemo(
    () => toUnifiedTokens(computedQ.data ?? [], dictionaryQ.data ?? []),
    [computedQ.data, dictionaryQ.data],
  );
  return {
    tokens,
    computedFailed: computedQ.isError,
    dictionaryFailed: dictionaryQ.isError,
  };
}
