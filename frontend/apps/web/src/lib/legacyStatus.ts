/**
 * Single canonical TypeScript home for retired document lifecycle status
 * vocabulary. Exists so legitimate legacy-rejection code (parse-boundary
 * guards, cutover-history assertions) can name a retired literal without
 * hardcoding the bare word, which would otherwise trip cilint's legacyvocab
 * guard (tools/cilint/internal/analyzers/legacyvocab.go).
 *
 * Migration/ADR trail: migration 0142 removed the draft -> finalized
 * transition ('review' was renamed to 'under_review'). FE-05
 * (frontend/apps/web/src/features/documents/lib/parseDocumentStatus.ts)
 * rejects the literal at the wire-status parse boundary rather than
 * silently aliasing it. The Go counterpart is
 * internal/platform/legacystatus/legacystatus.go.
 *
 * This file, and its Go counterpart, are the ONLY permitted homes for the
 * raw literal — both are listed in legacyExcludeDirs
 * (tools/cilint/internal/analyzers/legacyvocab.go). There is no per-line
 * suppression directive; a structural, reviewable path allowlist is the
 * only escape hatch.
 */

/**
 * The retired pre-cutover document status literal ('finalized'). It is
 * never a member of the canonical `DocumentStatus` union; it exists only so
 * legacy-rejection code can reference it by name instead of hardcoding the
 * bare string. Named LEGACY_RETIRED_STATUS rather than after the retired
 * word itself: a usage site that spelled the word out would trip the very
 * legacyvocab guard this file exists to satisfy (the pattern matches
 * comments and identifiers alike, case-insensitively).
 */
export const LEGACY_RETIRED_STATUS = 'finalized';
