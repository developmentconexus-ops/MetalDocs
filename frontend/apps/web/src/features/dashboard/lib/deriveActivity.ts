import type { AuditEventItem } from '../api/audit';

export interface ActivityItem {
  id: string;
  who: string;
  what: string;
  code: string;
  occurredAt: string;
}

// Known audit actions → a short PT verb phrase that reads between actor and code:
// "<who> <what> <code>". Unknown actions fall back to a readable de-dotted form so
// the line is never blank (audit actions are an open set).
const ACTION_LABELS: Record<string, string> = {
  'workflow.transitioned': 'avançou o fluxo de',
  'document.created': 'criou',
  'document.updated': 'editou',
  'document.published': 'publicou',
  'document.approved': 'aprovou',
  'document.rejected': 'devolveu',
  'document.archived': 'arquivou',
  'document.renamed': 'renomeou',
};

function humanizeAction(action: string): string {
  return ACTION_LABELS[action] ?? action.replace(/[._]/g, ' ').trim();
}

export function deriveActivityItems(events: AuditEventItem[]): ActivityItem[] {
  return events.map((e) => ({
    id: e.id,
    who: e.actor_id,
    what: humanizeAction(e.action),
    code: e.resource_id,
    occurredAt: e.occurred_at,
  }));
}
