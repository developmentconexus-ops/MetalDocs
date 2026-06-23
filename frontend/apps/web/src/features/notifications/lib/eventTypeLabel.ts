// eventTypeLabel maps a notification `event_type` (the producer's dotted domain
// event, e.g. "document.published") to a short pt-BR chip label. The row's title
// and message are already localized by the backend; this is the category chip.
// Unknown types fall back to a humanized last segment so a new producer event
// never renders as a raw machine string.

const LABELS: Record<string, string> = {
  'document.published': 'Publicado',
  'document.superseded': 'Substituído',
  'document.obsoleted': 'Obsoleto',
  'document.approved': 'Aprovado',
  'document.rejected': 'Reprovado',
};

export function eventTypeLabel(eventType: string): string {
  const known = LABELS[eventType];
  if (known) return known;
  // Fallback: take the segment after the last dot, capitalize it.
  const tail = eventType.includes('.') ? eventType.slice(eventType.lastIndexOf('.') + 1) : eventType;
  if (!tail) return 'Notificação';
  return tail.charAt(0).toUpperCase() + tail.slice(1);
}
