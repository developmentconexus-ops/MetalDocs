// Package httpdelivery is the IAM module's HTTP delivery layer: tier-1
// capability-gated route handlers for admin/roles, People-tab user
// management, area memberships, roles & capabilities catalogues, sessions,
// and observability (usage/KPI). Every handler resolves tenant scope from
// request context (never a client header) and maps cross-tenant lookups to
// 404, never 403, per the multi-tenant pooled invariant.
//
// router.go — CON-07 codegen rollout for IAM (ADR 0012 / target-arch N2).
//
// Mounts the oapi-codegen-generated iamapi.ServerInterface via
// iamapi.HandlerWithOptions instead of six independent hand-written
// mux.HandleFunc call sites (AdminHandler, PeopleHandler, MembershipHandler,
// RolesCapsHandler, SessionsHandler, plus presence's own snapshot route).
// Mirrors the templates pattern (internal/modules/templates/delivery/http/
// handler.go, commit e045f2de) and the original controlleddocuments
// reference (internal/modules/controlleddocuments/delivery/http/handler.go).
//
// Router is a thin ServerInterface adapter, NOT a reimplementation: every
// method below delegates straight through to the existing, already-tested
// handler methods on AdminHandler/PeopleHandler/MembershipHandler/
// RolesCapsHandler/SessionsHandler. Those handlers keep their own
// RegisterRoutes and internal methods unchanged (still directly unit-tested
// by admin_handler_test.go, routes_roles_caps_test.go,
// routes_memberships_contract_test.go, sessions_handler_test.go,
// method_not_allowed_test.go) — Router only owns route→typed-param
// extraction and mounting, per the task boundary ("update mounting details
// only").
//
// GetPresenceSnapshot delegates to presence.Handler.ServeSnapshot (a thin
// export of the existing, already-tested handleSnapshot). Only the WebSocket
// /iam/presence/stream upgrade stays on presence.Handler.RegisterRoutes —
// streamPresence is excluded from server codegen via cfg.yaml's
// exclude-operation-ids (WS upgrades have no meaningful oapi-codegen typed
// signature). Splitting snapshot (HTTP, codegen-eligible) from stream (WS,
// codegen-excluded) mounting is deliberate, not a fragmentation: it lets
// GetPresenceSnapshot join the other 18 generated ops under one
// HandlerWithOptions call while stream keeps its own explicit mux entry. See
// main.go startPresence for where presence.Handler is constructed and passed
// into NewRouter.
//
// CreateManagedUser (POST /iam/users, operationId createManagedUser) is
// declared in the spec and required by ServerInterface, but has never been
// mounted at runtime — PR-4 replaced it with POST /iam/users/invite and the
// spec still marks it deprecated: true. No business logic exists for it, so
// this adapter returns 501 NOT_IMPLEMENTED rather than guessing a semantics.
// This is the one honest gap surfaced by the codegen rollout; it is a
// pre-existing spec/runtime drift (dead deprecated op), not something this
// change introduces.
package httpdelivery

import (
	"metaldocs/internal/platform/httprouter"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"

	iamapi "metaldocs/internal/modules/iam/api"
	iampresence "metaldocs/internal/modules/iam/presence"
	"metaldocs/internal/platform/problem"
)

// Router implements iamapi.ServerInterface by delegating to the existing
// per-tab handler structs. Construct with NewRouter once all six handlers
// are wired, then mount via RegisterGenerated (or call
// iamapi.HandlerWithOptions directly with Router as the ServerInterface).
type Router struct {
	admin         *AdminHandler
	people        *PeopleHandler
	memberships   *MembershipHandler
	rolesCaps     *RolesCapsHandler
	sessions      *SessionsHandler
	observability *ObservabilityHandler
	presence      *iampresence.Handler
	tenants       *TenantHandler
}

// NewRouter builds the ServerInterface adapter. sessions, observability, and
// presence may be nil (no SQLDB boot path / presence not started); the
// corresponding operations then answer 501, matching the pre-codegen
// behavior in main.go where those handlers were only conditionally mounted.
func NewRouter(admin *AdminHandler, people *PeopleHandler, memberships *MembershipHandler, rolesCaps *RolesCapsHandler, sessions *SessionsHandler, observability *ObservabilityHandler, presence *iampresence.Handler) *Router {
	return &Router{
		admin:         admin,
		people:        people,
		memberships:   memberships,
		rolesCaps:     rolesCaps,
		sessions:      sessions,
		observability: observability,
		presence:      presence,
	}
}

// WithTenantHandler wires the M7 F7.2 tenant-onboarding handler. When unset,
// OnboardTenant answers 501 (matching the pre-Task-C stub behavior for boot
// paths without a configured SQLDB) — mirrors WithAuditEventLister's
// post-construction wiring convention.
func (rt *Router) WithTenantHandler(tenants *TenantHandler) *Router {
	rt.tenants = tenants
	return rt
}

// RegisterGenerated mounts the full generated IAM ServerInterface on mux
// under the /api/v1 base URL (AD-1: spec path keys are relative). Replaces
// the six RegisterRoutes call sites this router supersedes:
// AdminHandler.RegisterRoutes, PeopleHandler.RegisterRoutes,
// MembershipHandler.RegisterRoutes, RolesCapsHandler.RegisterRoutes,
// SessionsHandler.RegisterRoutes. Tier-1 authz, rate limiting, and
// panic/observability middleware are unaffected — they wrap the whole mux in
// main.go's buildChain and key off r.Method/r.URL.Path (permissions.go
// routeRule), not off mux dispatch mechanics, so swapping hand-written
// mux.HandleFunc registration for codegen-generated registration changes no
// tier-1 behavior. IAM has no per-route Idempotency-Key requirement (unlike
// templates/controlleddocuments), so no Middlewares closure is needed here.
func (rt *Router) RegisterGenerated(mux httprouter.Muxer) {
	iamapi.HandlerWithOptions(rt, iamapi.StdHTTPServerOptions{
		BaseRouter: mux,
		BaseURL:    "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			_ = problem.Write(w, problem.New(http.StatusBadRequest, problem.CodeRequestInvalid, err.Error()))
		},
	})
}

// ─── auth/sessions ──────────────────────────────────────────────────────────

// ListSessions delegates to SessionsHandler.handleSessions; answers 501 when
// sessions is not wired (SQLDB-less boot path).
func (rt *Router) ListSessions(w http.ResponseWriter, r *http.Request, _ iamapi.ListSessionsParams) {
	if rt.sessions == nil {
		writeIAMNotImplemented(w, "Sessions service is not configured")
		return
	}
	rt.sessions.handleSessions(w, r)
}

// RevokeSession delegates to SessionsHandler.handleSessionByID; answers 501
// when sessions is not wired.
func (rt *Router) RevokeSession(w http.ResponseWriter, r *http.Request, _ string) {
	if rt.sessions == nil {
		writeIAMNotImplemented(w, "Sessions service is not configured")
		return
	}
	// handleSessionByID re-derives session_id from r.URL.Path itself
	// (strings.TrimPrefix); the generated sessionId path param is discarded
	// here to avoid keeping two extraction paths in sync for no behavior
	// difference (Go 1.22 mux guarantees r.URL.Path and r.PathValue agree).
	rt.sessions.handleSessionByID(w, r)
}

// ─── iam/admin/overview ─────────────────────────────────────────────────────

// GetIamAdminOverview delegates to AdminHandler.handleAdminOverview, which
// composes KPI, presence, and recent-activity snapshots concurrently.
func (rt *Router) GetIamAdminOverview(w http.ResponseWriter, r *http.Request) {
	rt.admin.handleAdminOverview(w, r)
}

// ─── iam/area-memberships ───────────────────────────────────────────────────

// ListAreaMemberships delegates to MembershipHandler.listMemberships, which
// resolves the caller's directory scope (tenant-wide/managed-areas/self-only)
// before filtering (ADR 0022 Phase 4).
func (rt *Router) ListAreaMemberships(w http.ResponseWriter, r *http.Request, _ iamapi.ListAreaMembershipsParams) {
	rt.memberships.listMemberships(w, r)
}

// GrantAreaMembership delegates to MembershipHandler.grantMembership, which
// enforces the self-grant lockout and tier-2 area authz before writing.
func (rt *Router) GrantAreaMembership(w http.ResponseWriter, r *http.Request) {
	rt.memberships.grantMembership(w, r)
}

// RevokeAreaMembership delegates to MembershipHandler.revokeMembership.
// Self-revoke is permitted (self-de-escalation, not escalation).
func (rt *Router) RevokeAreaMembership(w http.ResponseWriter, r *http.Request, _ string, _ string) {
	// revokeMembership reads user_id/area_code via r.PathValue, which the
	// Go 1.22 mux populates identically whether the pattern was registered
	// by hand or by the generated HandlerWithOptions (same {user_id}/
	// {area_code} wildcard names — AD-1).
	rt.memberships.revokeMembership(w, r)
}

// ─── iam/capabilities, iam/roles, iam/role-capabilities ────────────────────

// ListCapabilities delegates to RolesCapsHandler.listCapabilities, serving
// the process-lifetime-cached capability catalogue.
func (rt *Router) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	rt.rolesCaps.listCapabilities(w, r)
}

// ListRoles delegates to RolesCapsHandler.listRoles, serving the
// process-lifetime-cached role catalogue.
func (rt *Router) ListRoles(w http.ResponseWriter, r *http.Request) {
	rt.rolesCaps.listRoles(w, r)
}

// ListRoleCapabilities delegates to RolesCapsHandler.listRoleCapabilities,
// reading the role→capability matrix from RoleCapabilitiesReader.
func (rt *Router) ListRoleCapabilities(w http.ResponseWriter, r *http.Request) {
	rt.rolesCaps.listRoleCapabilities(w, r)
}

// ─── iam/kpi, iam/usage ─────────────────────────────────────────────────────

// GetKpi delegates to ObservabilityHandler.handleKpi; answers 501 when
// observability is not wired.
func (rt *Router) GetKpi(w http.ResponseWriter, r *http.Request) {
	if rt.observability == nil {
		writeIAMNotImplemented(w, "Observability service is not configured")
		return
	}
	rt.observability.handleKpi(w, r)
}

// GetUsage delegates to ObservabilityHandler.handleUsage; answers 501 when
// observability is not wired.
func (rt *Router) GetUsage(w http.ResponseWriter, r *http.Request) {
	if rt.observability == nil {
		writeIAMNotImplemented(w, "Observability service is not configured")
		return
	}
	rt.observability.handleUsage(w, r)
}

// ─── iam/presence/snapshot ───────────────────────────────────────────────────

// GetPresenceSnapshot delegates to presence.Handler.ServeSnapshot; answers
// 501 when presence is not wired. The WebSocket stream upgrade is excluded
// from codegen and stays on presence.Handler.RegisterRoutes (see package doc).
func (rt *Router) GetPresenceSnapshot(w http.ResponseWriter, r *http.Request) {
	if rt.presence == nil {
		writeIAMNotImplemented(w, "Presence service is not configured")
		return
	}
	rt.presence.ServeSnapshot(w, r)
}

// ─── iam/users ───────────────────────────────────────────────────────────────

// ListUsers delegates to PeopleHandler.handleListUsers (filtered, cursor-paginated).
func (rt *Router) ListUsers(w http.ResponseWriter, r *http.Request, _ iamapi.ListUsersParams) {
	rt.people.handleListUsers(w, r)
}

// CreateManagedUser: see package doc comment. Dead, deprecated spec op with
// no runtime implementation prior to this change either.
func (rt *Router) CreateManagedUser(w http.ResponseWriter, r *http.Request) {
	writeIAMNotImplemented(w, "createManagedUser is deprecated and not implemented; use POST /iam/users/invite")
}

// BulkUsers delegates to PeopleHandler.handleBulk, applying a per-user
// mutation with per-item failure isolation.
func (rt *Router) BulkUsers(w http.ResponseWriter, r *http.Request) {
	rt.people.handleBulk(w, r)
}

// InviteUser delegates to PeopleHandler.handleInvite, provisioning a new
// tenant-managed user with a one-time temp password.
func (rt *Router) InviteUser(w http.ResponseWriter, r *http.Request) {
	rt.people.handleInvite(w, r)
}

// PatchUser delegates to PeopleHandler.handlePatch, applying a partial
// update to a managed user's metadata and/or tenant role.
func (rt *Router) PatchUser(w http.ResponseWriter, r *http.Request, _ string) {
	rt.people.handlePatch(w, r)
}

// ListMemberships delegates to PeopleHandler.handleListMemberships, returning
// a single user's active area memberships.
func (rt *Router) ListMemberships(w http.ResponseWriter, r *http.Request, _ string) {
	rt.people.handleListMemberships(w, r)
}

// ResetPassword delegates to PeopleHandler.handleResetPassword.
func (rt *Router) ResetPassword(w http.ResponseWriter, r *http.Request, _ string) {
	rt.people.handleResetPassword(w, r)
}

// UpsertUserRole delegates to AdminHandler.handleUserRoleUpsert.
func (rt *Router) UpsertUserRole(w http.ResponseWriter, r *http.Request, userId string) {
	rt.admin.handleUserRoleUpsert(w, r, userId)
}

// ReplaceUserRoles delegates to AdminHandler.handleReplaceUserRoles.
func (rt *Router) ReplaceUserRoles(w http.ResponseWriter, r *http.Request, userId string) {
	rt.admin.handleReplaceUserRoles(w, r, userId)
}

// UnlockUser delegates to PeopleHandler.handleUnlock.
func (rt *Router) UnlockUser(w http.ResponseWriter, r *http.Request, _ string) {
	rt.people.handleUnlock(w, r)
}

// OnboardTenant delegates to TenantHandler.handleOnboardTenant (M7 F7.2 Task
// C). Answers 501 when tenants is not wired (SQLDB-less boot path), matching
// the Task B stub's fallback behavior.
func (rt *Router) OnboardTenant(w http.ResponseWriter, r *http.Request) {
	if rt.tenants == nil {
		writeIAMNotImplemented(w, "Tenant onboarding service is not configured")
		return
	}
	rt.tenants.handleOnboardTenant(w, r)
}

// ExportTenant delegates to TenantHandler.handleExportTenant (M7 F7.3).
// Answers 501 when tenants is not wired (SQLDB-less boot path).
func (rt *Router) ExportTenant(w http.ResponseWriter, r *http.Request, tenantId openapi_types.UUID) {
	if rt.tenants == nil {
		writeIAMNotImplemented(w, "Tenant lifecycle service is not configured")
		return
	}
	rt.tenants.handleExportTenant(w, r, tenantId.String())
}

// EraseTenant delegates to TenantHandler.handleEraseTenant (M7 F7.3).
// Answers 501 when tenants is not wired (SQLDB-less boot path).
func (rt *Router) EraseTenant(w http.ResponseWriter, r *http.Request, tenantId openapi_types.UUID) {
	if rt.tenants == nil {
		writeIAMNotImplemented(w, "Tenant lifecycle service is not configured")
		return
	}
	rt.tenants.handleEraseTenant(w, r, tenantId.String())
}

func writeIAMNotImplemented(w http.ResponseWriter, detail string) {
	if err := problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalNotImplemented, detail)); err != nil {
		panic(err) // writer failure is unrecoverable here; surfaces as a panic caught by platform/middleware.Recovery
	}
}
