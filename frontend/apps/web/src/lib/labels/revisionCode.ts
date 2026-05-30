// ADR 0013 — Template & Document revision label formatter.
//
// Shared helper consumed by both `features/documents` and `features/templates`.
// Renders the regulated 0-based `revision_number` as `REV{nn}` (zero-padded).
//
// Lifted from `features/documents/lib/documentDetailMeta.ts`; the original path
// re-exports this function for backward compatibility.

export function formatRevisionCode(revisionNumber: number | null | undefined): string {
  const safeRevisionNumber =
    typeof revisionNumber === 'number' && Number.isFinite(revisionNumber) ? revisionNumber : 0;
  return `REV${String(Math.max(safeRevisionNumber, 0)).padStart(2, '0')}`;
}
