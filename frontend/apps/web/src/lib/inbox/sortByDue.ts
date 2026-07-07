// F5 (M2c): pure due-date-ascending sort for the approval worklist. Items with
// no due_at (null) sort last — they carry no urgency signal, so they should
// never crowd out items that do. Does not mutate the input array.
import type { InboxItem } from '../../features/approval/api/approvalTypes';

export function sortByDueAsc(items: InboxItem[]): InboxItem[] {
  return [...items].sort((a, b) => {
    if (a.due_at == null && b.due_at == null) return 0;
    if (a.due_at == null) return 1;
    if (b.due_at == null) return -1;
    return new Date(a.due_at).getTime() - new Date(b.due_at).getTime();
  });
}
