// F2d.6 — extract display text from an opaque comment body (ProseMirror node[]).
// The API stores comment content as free-form DocumentCommentContentNode[]
// (lib/api-types: `{ [key: string]: unknown }`); text lives in `text` leaves.
// Fail-closed: unknown/empty shapes yield '' (never throw) — no-fallback friendly.
export function commentText(body: unknown): string {
  const walk = (node: unknown): string => {
    if (!node || typeof node !== 'object') return '';
    const n = node as { text?: unknown; content?: unknown };
    if (typeof n.text === 'string') return n.text;
    if (Array.isArray(n.content)) return n.content.map(walk).join('');
    return '';
  };
  if (!Array.isArray(body)) return '';
  return body.map(walk).join(' ').trim();
}
