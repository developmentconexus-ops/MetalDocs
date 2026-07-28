// Date formatters for DocumentPublishedPage. Pure, locale-aware (pt-BR).
// Inputs are RFC3339 strings (or null/undefined for missing data); outputs are
// display strings with em-dash fallback so the page never renders "Invalid Date".

import type { components } from '../../../lib/api-types';

const EM_DASH = '—';

export const DEFAULT_INITIAL_REVISION_TITLE = 'Criação do documento';

// ADR 0035 discipline: the visibility shape comes from the generated contract
// types, never a hand-synced local mirror (the previous inline literal silently
// drifted from `ControlledDocumentVisibility` in the spec).
type ControlledDocumentVisibility = components['schemas']['ControlledDocumentVisibility'];

export function hasSettledSidebarIdentity(input: {
  code: string | null | undefined;
  profileLabel: string | null | undefined;
  areaLabel: string | null | undefined;
  visibilityLabel: string | null | undefined;
}): boolean {
  return Boolean(input.code && input.profileLabel && input.areaLabel && input.visibilityLabel);
}

// Canonical implementations lifted to `lib/format/dates.ts`.
// Re-exported here to keep existing documents-side imports working.
export { formatSignedAt, formatPublishedAt } from '../../../lib/format/dates';

// Canonical implementation lifted to `lib/format/fileSize.ts`.
// Re-exported here to keep existing documents-side imports working.
export { formatFileSize } from '../../../lib/format/fileSize';

// "12 páginas" → just the integer; em-dash when the API returns null/undefined.
export function formatPageCount(count: number | null | undefined): string {
  if (count == null || Number.isNaN(count) || count < 0) return EM_DASH;
  return String(count);
}

// Resolves profile code snapshot → human label using profiles list
export function resolveProfileLabel(code: string, profiles: Array<{ code: string; name: string }>): string {
  return profiles.find((p) => p.code === code)?.name ?? code;
}

// Resolves area code snapshot → human label using areas list
export function resolveAreaLabel(code: string, areas: Array<{ code: string; name: string }>): string {
  return areas.find((a) => a.code === code)?.name ?? code;
}

// ADR 0013: canonical implementation lifted to `lib/labels/revisionCode.ts`.
// Re-exported here to keep existing documents-side imports working.
export { formatRevisionCode } from '../../../lib/labels/revisionCode';

// Canonical implementation lifted to `lib/format/dates.ts`.
// Re-exported here to keep existing documents-side imports working.
export { formatShortDate } from '../../../lib/format/dates';

export function displayRevisionTitle(title: string | null | undefined, revisionCode: string): string {
  const trimmed = title?.trim();
  if (trimmed) return trimmed;
  return revisionCode === 'REV00' ? DEFAULT_INITIAL_REVISION_TITLE : EM_DASH;
}

export function buildVisibilityLabel(
  visibility: ControlledDocumentVisibility | null | undefined,
  areas: Array<{ code: string; name: string }>,
): string {
  if (!visibility) return EM_DASH;
  if (visibility.scope === 'company') return 'Toda empresa';
  if (visibility.user_ids.length > 0) {
    return `Restrito (${visibility.user_ids.length} usuarios, ${visibility.area_codes.length} areas)`;
  }
  if (visibility.area_codes.length === 0) return 'Restrito';
  const areaNames = visibility.area_codes.map((code) => resolveAreaLabel(code, areas));
  return `Restrito a area ${areaNames.join(', ')}`;
}
