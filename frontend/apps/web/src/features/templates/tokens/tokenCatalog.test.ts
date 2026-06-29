import { describe, expect, it } from 'vitest';
import { toUnifiedTokens } from './tokenCatalog';
import type { PlaceholderCatalogEntry } from '../api/catalog';
import type { TokenDictionaryEntry } from '../../tokens/api/tokensTypes';

describe('toUnifiedTokens', () => {
  it('tags computed and dictionary tokens by kind', () => {
    const computed: PlaceholderCatalogEntry[] = [
      { key: 'doc_code', label: 'Código', description: 'd' },
    ];
    const dictionary: TokenDictionaryEntry[] = [
      {
        id: '1',
        name: 'COMPANY_NAME',
        value: 'ACME',
        label: 'Empresa',
        description: null,
        created_at: '',
        updated_at: '',
      },
    ];

    const tokens = toUnifiedTokens(computed, dictionary);

    expect(tokens).toEqual([
      { key: 'doc_code', label: 'Código', description: 'd', kind: 'computed' },
      { key: 'COMPANY_NAME', label: 'Empresa', description: undefined, kind: 'dictionary' },
    ]);
  });

  it('uses the dictionary entry name as the placeholder key', () => {
    const dictionary: TokenDictionaryEntry[] = [
      {
        id: '1',
        name: 'CITY',
        value: 'SP',
        label: 'Cidade',
        description: 'x',
        created_at: '',
        updated_at: '',
      },
    ];
    const [tok] = toUnifiedTokens([], dictionary);
    expect(tok.key).toBe('CITY');
    expect(tok.kind).toBe('dictionary');
  });
});
