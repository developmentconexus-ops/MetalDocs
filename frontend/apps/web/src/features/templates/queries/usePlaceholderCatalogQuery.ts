import { useQuery } from '@tanstack/react-query';
import { fetchPlaceholderCatalog } from '../api/catalog';
import { QK } from '../../../lib/queryKeys';

// Placeholder catalog is static server config (the set of system-filled tokens),
// so it never goes stale within a session: staleTime Infinity, no refetch churn.
export function usePlaceholderCatalogQuery() {
  return useQuery({
    queryKey: QK.templates.placeholderCatalog(),
    queryFn: fetchPlaceholderCatalog,
    staleTime: Infinity,
  });
}
