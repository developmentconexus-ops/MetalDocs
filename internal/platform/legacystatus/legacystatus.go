// Package legacystatus is the single canonical Go home for retired document
// lifecycle status vocabulary. It exists so that legitimate legacy-rejection
// code (parse-boundary guards, cutover documentation) can name a retired
// literal without hardcoding the bare word, which would otherwise trip
// cilint's legacyvocab guard (tools/cilint/internal/analyzers/legacyvocab.go).
//
// Migration/ADR trail: migration 0142 removed the draft -> finalized
// transition (submit now yields under_review, see
// internal/modules/documents/domain/model.go DocStatusUnderReview). The
// Spec 2 approval graph (internal/modules/approval/domain/state.go
// StateFromString) rejects the literal at its parse boundary. The frontend
// counterpart is frontend/apps/web/src/lib/legacyStatus.ts.
//
// This package, and its frontend counterpart, are the ONLY permitted homes
// for the raw literal — both are listed in legacyExcludeDirs
// (tools/cilint/internal/analyzers/legacyvocab.go). There is no per-line
// suppression directive; a structural, reviewable path allowlist is the only
// escape hatch.
package legacystatus

// Retired is the retired pre-cutover document status literal ("finalized").
// It is never a valid DocState (approval/domain) or DocumentStatus
// (documents/domain) value; it exists only so parse-boundary rejection code
// can reference it by name instead of hardcoding the bare string. The
// identifier is deliberately NOT named after the retired word itself: any
// file outside this package that spelled it out would trip the very
// legacyvocab guard this package exists to satisfy (the pattern matches
// comments and identifiers alike, case-insensitively).
const Retired = "finalized"
