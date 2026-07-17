package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"metaldocs/internal/modules/approval/domain"
)

// RouteAdminReplay is the cached outcome envelope of a previously completed
// route admin operation. Update populates NewVersion; Create and Deactivate
// leave it nil.
type RouteAdminReplay struct {
	RouteID    string
	NewVersion *int
}

// RouteAdminReplayCommitter is the slot handle returned by a winning Begin*Replay
// call. The caller must resolve it with exactly one of Complete or Fail.
type RouteAdminReplayCommitter interface {
	Complete(routeID string, newVersion *int) error
	Fail(cause error) error
}

// RouteAdminIdempStore is the contract the service uses to persist replay
// envelopes for route-admin mutating operations. Production impl lives in
// infrastructure/postgres_route_admin_idemp_store.go.
type RouteAdminIdempStore interface {
	BeginCreateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error)
	BeginUpdateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error)
	BeginDeactivateReplay(ctx context.Context, tenantID, actorID, idempKey, payloadHash string) (RouteAdminReplayCommitter, *RouteAdminReplay, error)
}

// computeCreateRoutePayloadHash builds a stable fingerprint from client-supplied
// inputs only (per ADR 0017): profile_code, name, the canonical stage
// sequence, and the EFFECTIVE subject (kind+key). Server-resolved identifiers
// MUST NOT appear here.
//
// subject must be the already-resolved result of resolveCreateRouteSubject
// (the same defaulting the service applies: an absent/document kind defaults
// its key to profileCode) — not the raw, possibly-empty client input. Without
// the subject lines, every template-subject create hashed identically
// regardless of subject_key (profileCode is always "" for a template route),
// so a second create sharing an Idempotency-Key + name + stages but a
// DIFFERENT subject_key would silently replay the first route's result
// instead of surfacing an idempotency payload-mismatch conflict (QR-A finding
// A). For the legacy document case, subject.Key is always profileCode (rail
// R1), so these two new lines are redundant with the existing profile_code
// line but harmless — the hash for a byte-identical legacy request is still
// deterministic and stable across repeats AGAINST THIS CODE VERSION. It is
// NOT byte-stable across the deploy that introduced the subject lines: an
// envelope persisted pre-deploy and retried post-deploy surfaces
// ErrConflict (409) instead of a replay until the 24h idempotency TTL
// expires — fail-safe and time-bounded, never a wrong replay.
//
// subject.Key is hashed EXACTLY as resolved/persisted, with NO trimming:
// resolveCreateRouteSubject does not trim, domain.Subject.Validate only
// rejects an empty key, and the key is persisted byte-for-byte — so the hash
// identity must match validation, persistence, and event identity
// byte-for-byte too. Trimming here would make two distinct persisted
// subjects (e.g. "tmpl-a" and " tmpl-a ") hash identically, letting a reused
// Idempotency-Key silently replay across them instead of surfacing
// ErrConflict (codex gate round-3 CRITICAL). This is another hash-composition
// change covered by the same documented ≤24h 409 fail-safe described above.
func computeCreateRoutePayloadHash(profileCode, name string, stages []domain.Stage, subject domain.Subject) string {
	return sha256Lines(
		"create",
		strings.TrimSpace(profileCode),
		strings.TrimSpace(name),
		canonicalStages(stages),
		string(subject.Kind),
		subject.Key,
	)
}

// computeUpdateRoutePayloadHash has no equivalent subject-omission gap (QR-A
// finding A only applies to Create). UpdateRouteInput carries no
// SubjectKind/SubjectKey — a route's subject is immutable post-creation — and
// routeID is already part of this hash, uniquely identifying which route
// (and therefore which subject) is being mutated. Two updates differing only
// in "which route" already hash differently via routeID.
func computeUpdateRoutePayloadHash(routeID string, expectedVersion int, name string, stages []domain.Stage) string {
	return sha256Lines(
		"update",
		strings.TrimSpace(routeID),
		strconv.Itoa(expectedVersion),
		strings.TrimSpace(name),
		canonicalStages(stages),
	)
}

func computeDeactivateRoutePayloadHash(routeID string, expectedVersion int, reason string) string {
	return sha256Lines(
		"deactivate",
		strings.TrimSpace(routeID),
		strconv.Itoa(expectedVersion),
		strings.TrimSpace(reason),
	)
}

func canonicalStages(stages []domain.Stage) string {
	var sb strings.Builder
	for _, s := range stages {
		quorumM := ""
		if s.QuorumM != nil {
			quorumM = strconv.Itoa(*s.QuorumM)
		}
		fmt.Fprintf(&sb, "%d|%s|%s|%s|%s|%s|%s\n",
			s.Order,
			strings.TrimSpace(s.Name),
			strings.TrimSpace(s.RequiredCapability),
			string(s.Quorum),
			quorumM,
			string(s.OnEligibilityDrift),
			canonicalSelectors(s.Selectors),
		)
	}
	return sb.String()
}

// canonicalSelectors fingerprints a stage's Selectors (M4 ActorSelector, unit
// 3.2 slice 6b: Selectors is the sole source of truth for a stage's actor
// pool, so it replaces the flat RequiredRole/AreaCode fields that used to
// feed this hash). Selector order is preserved as-given — the HTTP boundary
// synthesis (route_admin_handler.go) produces a deterministic order for any
// given request, so this stays a stable fingerprint of client-supplied input.
func canonicalSelectors(selectors []domain.ActorSelector) string {
	var sb strings.Builder
	for _, sel := range selectors {
		fmt.Fprintf(&sb, "%s,%s,%s,%s;",
			string(sel.Kind),
			strings.TrimSpace(sel.UserID),
			strings.TrimSpace(sel.Role),
			strings.TrimSpace(sel.AreaCode),
		)
	}
	return sb.String()
}

func sha256Lines(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))
}
