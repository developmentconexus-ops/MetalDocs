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
// inputs only (per ADR 0017): profile_code, name, and the canonical stage
// sequence. Server-resolved identifiers MUST NOT appear here.
func computeCreateRoutePayloadHash(profileCode, name string, stages []domain.Stage) string {
	return sha256Lines(
		"create",
		strings.TrimSpace(profileCode),
		strings.TrimSpace(name),
		canonicalStages(stages),
	)
}

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
