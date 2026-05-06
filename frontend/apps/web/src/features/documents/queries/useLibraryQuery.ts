import { useQuery } from '@tanstack/react-query';
import { fetchLibrary } from '../api/library';
import { QK } from '../../../lib/queryKeys';

type LibraryQueryParams = {
  page?: number;
  pageSize?: number;
  status?: string;
  areaCode?: string;
  profileCode?: string;
  q?: string;
  includeArchived?: boolean;
};

export function useLibraryQuery(params: LibraryQueryParams) {
  return useQuery({
    queryKey: QK.documents.list(params),
    queryFn: () => fetchLibrary(params),
  });
}
