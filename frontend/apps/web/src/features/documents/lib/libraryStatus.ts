// Single source of truth for the document statuses surfaced in the Library
// view. Maps the backend status (which is a wider superset) to the URL filter
// slug, the pt-BR label, and the sidebar/tabs display order.
//
// Anywhere in the Library feature that needs to translate between status,
// filter, and label should import from this module — do NOT redeclare these
// strings in component files.

import type { DocumentStatus } from '../../../components/ui/StatusPill';

export type LibraryFilter =
  | 'todos'
  | 'rascunhos'
  | 'em_revisao'
  | 'aprovados'
  | 'publicados'
  | 'rejeitados'
  | 'obsoletos';

type StatusEntry = {
  /** Backend status value (codegen / `DocumentStatus`). */
  status: DocumentStatus;
  /** URL/UI filter slug used by the sidebar + tabs. */
  filter: Exclude<LibraryFilter, 'todos'>;
  /** pt-BR label for tabs / sidebar / labels. */
  label: string;
};

export const LIBRARY_STATUSES: readonly StatusEntry[] = [
  { status: 'draft',         filter: 'rascunhos',  label: 'Rascunhos'   },
  { status: 'under_review',  filter: 'em_revisao', label: 'Em Revisão'  },
  { status: 'approved',      filter: 'aprovados',  label: 'Aprovados'   },
  { status: 'published',     filter: 'publicados', label: 'Publicados'  },
  { status: 'rejected',      filter: 'rejeitados', label: 'Rejeitados'  },
  { status: 'obsolete',      filter: 'obsoletos',  label: 'Obsoletos'   },
] as const;

export function filterToStatus(filter: LibraryFilter): DocumentStatus | undefined {
  if (filter === 'todos') return undefined;
  return LIBRARY_STATUSES.find((entry) => entry.filter === filter)?.status;
}
