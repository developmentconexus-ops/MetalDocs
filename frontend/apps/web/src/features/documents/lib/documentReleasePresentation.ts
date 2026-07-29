// ADR 0085 Stage C — presentation for the release readiness-hold projection.
//
// The projection (`DocumentDetailResponse.release`) is the UI's ONLY source of
// release status: there is no manual publish gate any more, and no client-side
// readiness inference (the old content_hash comparison died in Stage B). This
// module maps the projection onto the copy the detail screen renders — kept
// standalone (mirror of documentStatusPresentation.ts) so the governed PT-BR
// strings live in one auditable place instead of hiding inside JSX.
//
// `release === null` means "no release generation at all" (pre-approval drafts,
// legacy rows not backfilled) → null presentation → the route renders NOTHING.
// Honest absence, never a fabricated state.

import { formatPublishedAt, formatShortDate } from '../../../lib/format/dates';
import type { components } from '../../../lib/api-types';

export type DocumentReleaseProjection = components['schemas']['DocumentReleaseProjection'];

/**
 * Semantic tone of the release notice. `progress`/`scheduled` are informational
 * (no user action exists — the coordinator owns the transition); `anomaly` is a
 * neutral-error state whose resolution is backend/operator work, so it carries
 * no CTA either.
 */
export type ReleaseNoticeTone = 'released' | 'progress' | 'scheduled' | 'anomaly';

export interface ReleasePresentation {
  tone: ReleaseNoticeTone;
  /** Primary line — the release state in the operator's language. */
  title: string;
  /** Secondary line (release date / coordinator hold detail); null when absent. */
  detail: string | null;
}

/**
 * Cadence of the lifecycle settling poll. This is the interval the pre-ADR-0085
 * `status === 'scheduled'` gate has always used — the settling gate below reuses
 * the SAME mechanism and the SAME number rather than introducing a second one.
 */
export const LIFECYCLE_POLL_INTERVAL_MS = 5_000;

/**
 * Hold reasons the coordinator clears on its own: the screen WILL change with no
 * user action, so it must keep polling or it lies (frozen "Preparando artefatos
 * finais…" until a manual reload).
 *
 * Deliberately NOT here:
 *  - `supersede_conflict` / `plan_invalid` / `failed` — terminal anomalies. They
 *    never resolve by waiting (only backend/operator work clears them), so polling
 *    them is pure noise.
 *  - `awaiting_effective_date` — resolves on a calendar date that can be weeks out.
 *    The only cadence this mechanism has is the scheduled gate's 5s, and hammering
 *    the API every 5s for weeks to catch one transition is not a trade worth making;
 *    the flip is picked up on the next navigation/refetch. Documents that still carry
 *    the legacy `status === 'scheduled'` keep their existing gate (see
 *    `isDocumentLifecycleSettling`), so the near-term scheduled case is unchanged.
 */
const SETTLING_HOLD_REASONS: ReadonlySet<string> = new Set(['awaiting_approval_fact', 'materializing']);

/**
 * True while the release generation is in a transient hold the coordinator is
 * actively working through. Mirrors the transient/terminal split the presentation
 * below already encodes (`progress` tone), kept beside it so the reason list has a
 * single home and cannot drift.
 */
export function isReleaseSettling(release: DocumentReleaseProjection | null | undefined): boolean {
  if (!release || release.state !== 'hold') return false;
  // hold_reason == null: a generation the coordinator has not evaluated yet — the
  // most transient state there is.
  if (release.hold_reason == null) return true;
  return SETTLING_HOLD_REASONS.has(release.hold_reason);
}

/**
 * Document-level poll gate: is this document's lifecycle still moving on its own?
 * Post-ADR-0085 a document sits at `status = 'approved'` while the release
 * coordinator works, so the legacy status check alone leaves the page stale.
 */
export function isDocumentLifecycleSettling(
  doc: { status?: string | null; release?: DocumentReleaseProjection | null } | null | undefined,
): boolean {
  if (!doc) return false;
  // Legacy pre-ADR-0085 gate, unchanged: a scheduled document flips without user action.
  if (doc.status === 'scheduled') return true;
  return isReleaseSettling(doc.release);
}

export function getDocumentReleasePresentation(
  release: DocumentReleaseProjection | null | undefined,
): ReleasePresentation | null {
  if (!release) return null;

  if (release.state === 'released') {
    return {
      tone: 'released',
      title: 'Publicado',
      detail: release.released_at ? `Liberado em ${formatPublishedAt(release.released_at)}` : null,
    };
  }

  switch (release.hold_reason) {
    // Both are transient coordinator processing with no user-visible difference:
    // the fact is being confirmed / the artifacts are being produced.
    case 'awaiting_approval_fact':
    case 'materializing':
      return { tone: 'progress', title: 'Preparando artefatos finais…', detail: null };

    case 'awaiting_effective_date':
      return {
        tone: 'scheduled',
        title: release.planned_effective_from
          ? `Aprovado — vigência programada para ${formatShortDate(release.planned_effective_from)}`
          : 'Aprovado — aguardando a data de vigência',
        detail: null,
      };

    // Anomalies — neutral-error tone, hold_detail verbatim, no CTA.
    case 'supersede_conflict':
      return {
        tone: 'anomaly',
        title: 'Publicação retida — conflito de substituição',
        detail: release.hold_detail,
      };
    case 'plan_invalid':
      return {
        tone: 'anomaly',
        title: 'Publicação retida — plano de publicação inválido',
        detail: release.hold_detail,
      };
    case 'failed':
      return {
        tone: 'anomaly',
        title: 'Publicação retida — falha no processamento',
        detail: release.hold_detail,
      };

    // hold_reason === null: a generation the coordinator has not evaluated yet.
    default:
      return { tone: 'progress', title: 'Preparando publicação…', detail: null };
  }
}
