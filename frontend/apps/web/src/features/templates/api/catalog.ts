import { apiFetch } from '../../../lib/api/client';
import type { components, paths } from '../../../lib/api-types';

export type PlaceholderCatalogEntry = components['schemas']['PlaceholderCatalogEntry'];
type PlaceholderCatalogResponse =
  paths['/api/v1/templates/placeholder-catalog']['get']['responses'][200]['content']['application/json'];

export async function fetchPlaceholderCatalog(): Promise<PlaceholderCatalogEntry[]> {
  const body = await apiFetch<PlaceholderCatalogResponse>('/api/v1/templates/placeholder-catalog', {});
  return body.items;
}
