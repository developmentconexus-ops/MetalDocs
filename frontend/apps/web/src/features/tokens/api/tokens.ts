import { apiFetch } from '../../../lib/api';
import type {
  CreateTokenDictionaryEntryRequest,
  ListTokenDictionaryEntriesResponse,
  TokenDictionaryEntry,
  UpdateTokenDictionaryEntryRequest,
} from './tokensTypes';

const BASE = '/api/v1/tokens';

export async function listTokens(): Promise<TokenDictionaryEntry[]> {
  const body = await apiFetch<ListTokenDictionaryEntriesResponse>(BASE, undefined);
  return body.items;
}

export async function createToken(
  req: CreateTokenDictionaryEntryRequest,
): Promise<TokenDictionaryEntry> {
  return apiFetch<TokenDictionaryEntry>(BASE, {
    method: 'POST',
    body: JSON.stringify(req),
  });
}

export async function updateToken(
  id: string,
  req: UpdateTokenDictionaryEntryRequest,
): Promise<TokenDictionaryEntry> {
  return apiFetch<TokenDictionaryEntry>(`${BASE}/${id}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  });
}

export async function deleteToken(id: string): Promise<void> {
  await apiFetch<void>(`${BASE}/${id}`, { method: 'DELETE' });
}
