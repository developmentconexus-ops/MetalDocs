//go:build integration && !production

package test

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"metaldocs/internal/platform/httprouter"
	"metaldocs/internal/platform/problem"
)

const (
	e2eAreaCode    = "qa"
	e2eProfileCode = "seed_profile"
	e2ePassword    = "test1234"
)

// e2eAssertedCaps is the tier-3 capability superset the seed must assert so the
// migration-0231 tripwire (enforce_capability_asserted) admits the guarded
// writes: taxonomy (areas/profiles), user roles, area membership, documents,
// templates/versions. The trigger matches on "cap" only (area is ignored), and
// uses ANY-of for each table, so a flat cap list with one cap per guarded family
// suffices. Set tx-local via set_config(..., true) at the top of the seed tx.
const e2eAssertedCaps = `[{"cap":"taxonomy.manage"},{"cap":"user.manage"},{"cap":"membership.manage"},{"cap":"document.create"},{"cap":"document.edit"},{"cap":"template.create"},{"cap":"controlled_documents.create"}]`

type seedHandler struct {
	db               *sql.DB
	runSchedulerTick func(context.Context) error
}

type seedRequest struct {
	TenantID string   `json:"tenant_id"`
	DocID    string   `json:"docId"`
	Roles    []string `json:"roles"`
}

type resetRequest struct {
	TenantID string `json:"tenant_id"`
}

type governanceEventRow struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	EventType     string          `json:"event_type"`
	ActorUserID   string          `json:"actor_user_id"`
	ResourceType  string          `json:"resource_type"`
	ResourceID    string          `json:"resource_id"`
	Reason        string          `json:"reason,omitempty"`
	PayloadJSON   json.RawMessage `json:"payload_json"`
	CreatedAt     string          `json:"created_at"`
	DedupeKey     string          `json:"dedupe_key,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	InstanceID    string          `json:"instance_id,omitempty"`
	DocumentID    string          `json:"doc_id,omitempty"`
}

type seededUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type seedResponse struct {
	TenantID string `json:"tenant_id"`
	DocID    string `json:"docId"`
	Users    struct {
		Author   seededUser `json:"author"`
		Reviewer seededUser `json:"reviewer"`
		Approver seededUser `json:"approver"`
		Admin    seededUser `json:"admin"`
	} `json:"users"`
	Cookies map[string]string `json:"cookies"`
	// ContentHash is the document's content_hash_at_submit, seeded deterministically.
	// F6 (no-fallback hash chain): a real HTTP signoff proof driven through the
	// actual submit→freeze pipeline must echo the approval instance's
	// frozen_content_hash (fetched from the active-document endpoint after
	// submit/freeze), NOT this static seed value — signoff no longer reads
	// documents.content_hash_at_submit at all. This field remains useful only
	// for proofs that seed a document directly into a pre-submit/draft state
	// and inspect the active-document endpoint's head-revision-hash branch
	// before any approval instance exists.
	ContentHash string `json:"content_hash"`
}

// RegisterE2EHandlers mounts the e2e scaffolding unconditionally. Mount is
// total (§4): db == nil and runSchedulerTick == nil are NOT mount
// conditionals — each handler answers 501 when its dependency is absent, so
// the route always exists and the boot-time declared/mounted assertion never
// depends on which optional dependency happened to be wired. !E2EEnabled()
// is likewise not checked here; it is a composition-root condition
// (e2eHandlersEnabled + e2ePublisher, apps/api/cmd/metaldocs-api/main.go —
// Task 15b replaced mountE2EHandlersIfEnabled's callback indirection with the
// inline useE2E gate) and survives
// per-handler as the real guard below. mux == nil is a wiring bug, not a
// branch, and panics.
func RegisterE2EHandlers(mux httprouter.Muxer, db *sql.DB, runSchedulerTick func(context.Context) error) {
	if mux == nil {
		panic("e2e_seed: RegisterE2EHandlers called with a nil Muxer")
	}

	h := &seedHandler{db: db, runSchedulerTick: runSchedulerTick}
	mux.HandleFunc("POST /internal/test/seed", h.seed)
	mux.HandleFunc("POST /internal/test/reset", h.reset)
	mux.HandleFunc("GET /internal/test/governance-events", h.governanceEvents)
	mux.HandleFunc("POST /internal/test/trigger-scheduler-tick", h.triggerSchedulerTick)
}

func (h *seedHandler) seed(w http.ResponseWriter, r *http.Request) {
	if !E2EEnabled() {
		http.NotFound(w, r)
		return
	}
	if h.db == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "e2e scaffolding requires a database"))
		return
	}

	var req seedRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	tenantID := strings.TrimSpace(req.TenantID)
	docID := strings.TrimSpace(req.DocID)
	if tenantID == "" || docID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId and docId are required"})
		return
	}

	roles := normalizeRoles(req.Roles)

	tx, err := h.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Assert the tier-3 capability superset (tx-local) so the migration-0231
	// tripwire admits the guarded taxonomy/role/membership/document/template
	// writes below. Without this the first guarded INSERT raises P0001.
	if _, err := tx.ExecContext(r.Context(), `SELECT set_config('metaldocs.asserted_caps', $1, true)`, e2eAssertedCaps); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Areas and profiles are keyed by (tenant_id, code) — a composite PK, so the
	// same code can exist in multiple tenants without collision. The seed still
	// scopes its area/profile codes per tenant so concurrent e2e runs sharing a
	// tenant don't race each other's rows.
	slug := sanitizeSlug(tenantID)
	areaCode := e2eAreaCode + "-" + slug
	profileCode := e2eProfileCode + "-" + slug

	if err := ensureTenant(r.Context(), tx, tenantID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if err := ensureAreaAndProfile(r.Context(), tx, tenantID, areaCode, profileCode); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	usersByRole := map[string]seededUser{}
	cookiesByRole := map[string]string{}
	for _, role := range roles {
		user, cookieValue, createErr := upsertSeedUser(r.Context(), tx, tenantID, role, areaCode)
		if createErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": createErr.Error()})
			return
		}
		usersByRole[role] = user
		cookiesByRole[role] = cookieValue
	}

	author, ok := usersByRole["author"]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "roles must include author"})
		return
	}
	admin := usersByRole["admin"]
	if admin.ID == "" {
		admin = author
	}

	tplVersionID, err := ensureTemplateVersion(r.Context(), tx, tenantID, admin.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := ensureApprovalRoute(r.Context(), tx, tenantID, admin.ID, profileCode, areaCode); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	cdID, err := ensureControlledDocument(r.Context(), tx, tenantID, profileCode, areaCode, author.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Deterministic, docID-derived content hash. The finalize/submit path computes
	// its own hash for approval_instances; this value seeds documents.content_hash_at_submit
	// as a pre-submit/draft-state fixture value only. F6 (no-fallback hash chain):
	// once a real instance is submitted and frozen, signoff/publish read ONLY
	// approval_instances.frozen_content_hash — a proof script driving the real
	// submit→freeze→signoff flow must fetch that pin from the active-document
	// endpoint after submit, not echo this seed value. Format-only (signoff
	// validates 64-hex + equality, not recomputation from content).
	sum := sha256.Sum256([]byte("metaldocs-e2e-content:" + docID))
	contentHash := hex.EncodeToString(sum[:])

	if err := upsertDraftDocument(r.Context(), tx, tenantID, docID, tplVersionID, author.ID, areaCode, profileCode, cdID, contentHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	res := seedResponse{TenantID: tenantID, DocID: docID, Cookies: cookiesByRole, ContentHash: contentHash}
	res.Users.Author = usersByRole["author"]
	res.Users.Reviewer = usersByRole["reviewer"]
	res.Users.Approver = usersByRole["approver"]
	res.Users.Admin = usersByRole["admin"]
	writeJSON(w, http.StatusOK, res)
}

func (h *seedHandler) reset(w http.ResponseWriter, r *http.Request) {
	if !E2EEnabled() {
		http.NotFound(w, r)
		return
	}
	if h.db == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "e2e scaffolding requires a database"))
		return
	}

	var req resetRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId is required"})
		return
	}

	tx, err := h.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback() }()

	statements := []string{
		// DB-10: legacy public.signoffs was removed from the reset list (2026-07-03).
		// The table does not exist in the current canonical baseline
		// (db/baseline/0001_current_schema.sql; grep-confirmed zero CREATE TABLE for
		// a bare "signoffs" table anywhere in the live schema) — only
		// public.approval_signoffs, deleted above, is live. The DELETE FROM signoffs
		// statement previously here was a permanent no-op: every execution hit
		// isUndefinedTable below and was swallowed by the `continue`. Confirmed zero
		// runtime writers to a legacy "signoffs" table anywhere in the tree (grep
		// across internal/ and tests/ for bare `signoffs` finds only this reset line,
		// the approval_signoffs family, and the /signoffs HTTP route segment).
		`DELETE FROM approval_signoffs s USING approval_instances i WHERE s.approval_instance_id = i.id AND i.tenant_id = $1`,
		`DELETE FROM approval_stage_instances s USING approval_instances i WHERE s.approval_instance_id = i.id AND i.tenant_id = $1`,
		`DELETE FROM approval_instances WHERE tenant_id = $1`,
		`DELETE FROM approval_route_stages rs USING approval_routes r WHERE rs.route_id = r.id AND r.tenant_id = $1`,
		`DELETE FROM approval_routes WHERE tenant_id = $1`,
		`DELETE FROM governance_events WHERE tenant_id = $1`,
		`DELETE FROM documents WHERE tenant_id = $1`,
		`DELETE FROM controlled_documents WHERE tenant_id = $1`,
		`DELETE FROM metaldocs.auth_sessions s USING metaldocs.iam_users u WHERE s.user_id = u.user_id AND u.tenant_id = $1`,
		`DELETE FROM metaldocs.auth_identities i USING metaldocs.iam_users u WHERE i.user_id = u.user_id AND u.tenant_id = $1`,
		`DELETE FROM metaldocs.iam_user_roles ur USING metaldocs.iam_users u WHERE ur.user_id = u.user_id AND u.tenant_id = $1`,
		`DELETE FROM metaldocs.iam_users WHERE tenant_id = $1`,
		`DELETE FROM users WHERE tenant_id = $1`,
		`DELETE FROM tenants WHERE id = $1`,
	}

	for _, q := range statements {
		if _, execErr := tx.ExecContext(r.Context(), q, tenantID); execErr != nil {
			if isUndefinedTable(execErr) || isUndefinedColumn(execErr) {
				continue
			}
			slog.Error("e2e reset failed", "tenant", tenantID, "err", execErr)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "reset failed"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *seedHandler) governanceEvents(w http.ResponseWriter, r *http.Request) {
	if !E2EEnabled() {
		http.NotFound(w, r)
		return
	}
	if h.db == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "e2e scaffolding requires a database"))
		return
	}

	tenantID := strings.TrimSpace(r.URL.Query().Get("tenantId"))
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tenantId is required"})
		return
	}

	docID := strings.TrimSpace(r.URL.Query().Get("docId"))
	instanceID := strings.TrimSpace(r.URL.Query().Get("instanceId"))

	rows, err := h.db.QueryContext(r.Context(), `
SELECT
  ge.id::text,
  ge.tenant_id::text,
  ge.event_type,
  ge.actor_user_id,
  ge.resource_type,
  ge.resource_id,
  COALESCE(ge.reason, ''),
  ge.payload_json,
  ge.created_at,
  COALESCE(ge.dedupe_key, ''),
  COALESCE(ge.correlation_id, ''),
  COALESCE(NULLIF(ge.payload_json->>'instance_id', ''), CASE WHEN ge.resource_type = 'approval_instance' THEN ge.resource_id ELSE '' END) AS instance_id,
  COALESCE(
    NULLIF(ge.payload_json->>'doc_id', ''),
    NULLIF(ge.payload_json->>'document_id', ''),
    CASE WHEN ge.resource_type = 'document' THEN ge.resource_id ELSE '' END,
    ai.document_id::text
  ) AS doc_id
FROM governance_events ge
LEFT JOIN approval_instances ai
  ON ge.resource_type = 'approval_instance'
 AND ge.resource_id = ai.id::text
WHERE ge.tenant_id = $1
  AND ($2 = '' OR
       ge.resource_id = $2 OR
       ge.payload_json->>'doc_id' = $2 OR
       ge.payload_json->>'document_id' = $2 OR
       ai.document_id::text = $2)
  AND ($3 = '' OR
       ge.resource_id = $3 OR
       ge.payload_json->>'instance_id' = $3)
ORDER BY ge.created_at ASC, ge.id ASC
`, tenantID, docID, instanceID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	events := make([]governanceEventRow, 0)
	for rows.Next() {
		var row governanceEventRow
		var createdAt time.Time
		if scanErr := rows.Scan(
			&row.ID,
			&row.TenantID,
			&row.EventType,
			&row.ActorUserID,
			&row.ResourceType,
			&row.ResourceID,
			&row.Reason,
			&row.PayloadJSON,
			&createdAt,
			&row.DedupeKey,
			&row.CorrelationID,
			&row.InstanceID,
			&row.DocumentID,
		); scanErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": scanErr.Error()})
			return
		}
		row.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
		events = append(events, row)
	}
	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, events)
}

// triggerSchedulerTick matches api/openapi/internal-e2e.yaml's
// e2eTriggerSchedulerTick: 204 on a successful tick, 501 when no scheduler
// tick is wired (runSchedulerTick == nil — true of the shipped metaldocs-api
// binary, which always passes nil; only the periodic-jobs test harness wires
// a real one).
func (h *seedHandler) triggerSchedulerTick(w http.ResponseWriter, r *http.Request) {
	if !E2EEnabled() {
		http.NotFound(w, r)
		return
	}
	if h.db == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "e2e scaffolding requires a database"))
		return
	}
	if h.runSchedulerTick == nil {
		_ = problem.Write(w, problem.New(http.StatusNotImplemented, problem.CodeInternalUnknown, "no scheduler tick is wired"))
		return
	}

	if err := h.runSchedulerTick(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ensureTenant(ctx context.Context, tx *sql.Tx, tenantID string) error {
	// Post-v1 re-baseline: the tenant table is metaldocs.tenants (id/name/slug all
	// NOT NULL); legacy public.tenants was removed. $1 is cast to uuid for the id
	// column; $2 (the same value, bound as text) sources the slug — kept distinct
	// so the ::uuid cast on $1 doesn't force a uuid type onto the left() argument.
	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.tenants (id, name, slug)
VALUES ($1::uuid, 'E2E Seed Tenant', 'e2e-' || left($2, 8))
ON CONFLICT (id) DO NOTHING`, tenantID, tenantID); err != nil {
		return fmt.Errorf("upsert tenant: %w", err)
	}
	return nil
}

func normalizeRoles(roles []string) []string {
	if len(roles) == 0 {
		return []string{"author", "reviewer", "approver", "admin"}
	}

	allowed := map[string]bool{"author": true, "reviewer": true, "approver": true, "admin": true}
	seen := map[string]bool{}
	out := make([]string, 0, 4)

	for _, role := range roles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if !allowed[normalized] || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}

	if len(out) == 0 {
		return []string{"author", "reviewer", "approver", "admin"}
	}
	return out
}

func ensureAreaAndProfile(ctx context.Context, tx *sql.Tx, tenantID, areaCode, profileCode string) error {
	// document_process_areas' PK was promoted to (tenant_id, code) by migration
	// 0264, and document_profiles' PK was promoted to (tenant_id, code) by
	// migration 0308 (0264's deferred HALF B) -- both tables now key on the
	// composite, so the same code may legally belong to more than one tenant.
	// Conflict resolution targets each table's composite PK.
	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.document_process_areas (tenant_id, code, name, description, is_active)
VALUES ($1, $2, 'QA', 'E2E seed area', TRUE)
ON CONFLICT (tenant_id, code) DO NOTHING`, tenantID, areaCode); err != nil {
		return fmt.Errorf("seed area: %w", err)
	}

	var familyCode string
	if err := tx.QueryRowContext(ctx, `SELECT code FROM metaldocs.document_families ORDER BY code LIMIT 1`).Scan(&familyCode); err != nil {
		return fmt.Errorf("select family: %w", err)
	}

	// alias has a 1-24 char CHECK (chk_document_profiles_alias_length); the tenant-
	// scoped profileCode ("seed_profile-<slug>") can exceed 24, so truncate.
	alias := profileCode
	if len(alias) > 24 {
		alias = alias[:24]
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.document_profiles (tenant_id, code, family_code, name, description, review_interval_days, alias)
VALUES ($1, $2, $3, 'Seed Profile', 'E2E seed profile', 365, $4)
ON CONFLICT (tenant_id, code) DO NOTHING`, tenantID, profileCode, familyCode, alias); err != nil {
		return fmt.Errorf("seed profile: %w", err)
	}

	return nil
}

func upsertSeedUser(ctx context.Context, tx *sql.Tx, tenantID, role, areaCode string) (seededUser, string, error) {
	slug := sanitizeSlug(tenantID)
	userID := fmt.Sprintf("e2e-%s-%s", role, slug)
	email := fmt.Sprintf("%s@%s.e2e", role, tenantID)
	displayName := fmt.Sprintf("E2E %s", titleCase(role))

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_users (user_id, display_name, is_active, tenant_id, deactivated_at, created_at, updated_at)
VALUES ($1, $2, TRUE, $3, NULL, now(), now())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name,
              is_active = TRUE,
              tenant_id = EXCLUDED.tenant_id,
              deactivated_at = NULL,
              updated_at = now()`, userID, displayName, tenantID); err != nil {
		return seededUser{}, "", fmt.Errorf("upsert iam user (%s): %w", role, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(e2ePassword), bcrypt.DefaultCost)
	if err != nil {
		return seededUser{}, "", fmt.Errorf("hash password: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.auth_identities (user_id, username, email, display_name, is_active, password_hash, password_algo, must_change_password, last_login_at, failed_login_attempts, locked_until, created_at, updated_at)
VALUES ($1, $2, $3, $4, TRUE, $5, 'bcrypt', FALSE, NULL, 0, NULL, now(), now())
ON CONFLICT (user_id)
DO UPDATE SET username = EXCLUDED.username,
              email = EXCLUDED.email,
              display_name = EXCLUDED.display_name,
              is_active = TRUE,
              password_hash = EXCLUDED.password_hash,
              password_algo = 'bcrypt',
              must_change_password = FALSE,
              updated_at = now()`, userID, userID, email, displayName, string(hash)); err != nil {
		return seededUser{}, "", fmt.Errorf("upsert auth identity (%s): %w", role, err)
	}

	iamRole := mapRoleToIAM(role)
	// tenant_id is required: the auth role resolver joins iam_user_roles on
	// (user_id, tenant_id), so a role row without the matching tenant_id reads as
	// "no roles assigned" even though the row exists.
	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, now(), 'e2e-seed')
ON CONFLICT (user_id, role_code)
DO UPDATE SET tenant_id = EXCLUDED.tenant_id, assigned_at = now(), assigned_by = EXCLUDED.assigned_by`, userID, tenantID, iamRole); err != nil {
		return seededUser{}, "", fmt.Errorf("upsert iam role (%s): %w", role, err)
	}

	membershipRole := mapRoleToMembership(role)
	// granted_by carries an FK (tenant_id, granted_by) -> iam_users; the legacy
	// 'e2e-seed' sentinel is not a real user. Self-grant (granted_by = userID) —
	// the user row was upserted above, so it always satisfies the FK in-tenant.
	// Direct INSERT is the sole path: the old grant_area_membership fallback was
	// dead code — that SECURITY DEFINER fn's uppercase area_code validation is
	// unsatisfiable against document_process_areas' lowercase area_code_format
	// CHECK, so it could never succeed and only masked the real insert error.
	if _, err := tx.ExecContext(ctx, `
INSERT INTO user_process_areas (user_id, tenant_id, area_code, role, effective_from, effective_to, granted_by, revoked_by)
VALUES ($1, $2, $3, $4, now(), NULL, $5, NULL)
ON CONFLICT (tenant_id, user_id, area_code, role) WHERE effective_to IS NULL DO NOTHING`, userID, tenantID, areaCode, membershipRole, userID); err != nil {
		return seededUser{}, "", fmt.Errorf("grant area membership (%s): %w", role, err)
	}

	cookieValue, err := createSessionValue(ctx, tx, userID, tenantID)
	if err != nil {
		return seededUser{}, "", err
	}

	return seededUser{ID: userID, Email: email}, cookieValue, nil
}

func mapRoleToIAM(role string) string {
	// Valid role_codes (chk_iam_user_roles_role_code): system_admin, approver,
	// author, editor, viewer. Legacy 'admin'/'reviewer' codes were decommissioned
	// (ADR 0022); map onto the surviving canonical roles.
	switch role {
	case "admin":
		return "system_admin"
	case "reviewer", "approver":
		return "approver"
	case "author":
		// The seed author must submit its own draft via POST /documents/{id}/submit
		// (ADR 0073 removed the /finalize wrapper). The submit path also drives the
		// approver-only surfaces in later e2e steps, so the seed author carries the
		// approver IAM role at tier-1 (system_admin is avoided — it would
		// short-circuit both the owner check and the tier-2 Require, defeating the
		// real-path proof). Its tier-2 area membership stays 'author' (grants
		// document.submit/edit), so the submit_service Require calls still exercise
		// the genuine author-area grant.
		return "approver"
	default:
		return "editor"
	}
}

func mapRoleToMembership(role string) string {
	switch role {
	case "author":
		// 'author' area role grants document.submit (role_capabilities) — required
		// to submit. 'editor' does NOT grant submit, so it would 403.
		return "author"
	default:
		// 'reviewer' was a non-functional legacy area role (decommissioned,
		// ADR 0022) — map any non-author seed role to the canonical approver
		// area membership, which grants document.signoff.
		return "approver"
	}
}

// ensureTemplateVersion seeds a published template + version in the CANONICAL
// templates_template / templates_template_version family (TST-01). The finalize/
// snapshot flow this harness exercises over HTTP resolves docx_storage_key via
// docgenv2.TemplatesTemplateReader — the only template reader since DB-01
// closed (2026-07-03): the legacy fallback chain (FanoutTemplateReader +
// legacy TemplateReader) was deleted and the legacy public.templates /
// template_versions tables dropped (migration 0268) after the run-window
// proof showed zero legacy fallback reads.
//
// templates_template_version carries an extra trigger beyond the tier-3 capability
// tripwire (enforce_capability_asserted, satisfied by e2eAssertedCaps' template.create):
// trg_template_version_tenant_consistent requires the tx-local metaldocs.tenant_id
// GUC to be set and to match the parent template's tenant. Set it once per seed tx.
func ensureTemplateVersion(ctx context.Context, tx *sql.Tx, tenantID, actorID string) (string, error) {
	templateID := uuid.NewString()
	templateVersionID := uuid.NewString()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.tenant_id', $1, true)`, tenantID,
	); err != nil {
		return "", fmt.Errorf("set tenant_id GUC: %w", err)
	}

	if err := tx.QueryRowContext(ctx, `
INSERT INTO templates_template (id, tenant_id, doc_type_code, key, name, description, latest_version, published_version_id, created_by, created_at)
VALUES ($1, $2::uuid, '', $3, 'E2E Template', 'E2E seed template', 1, NULL, $4, now())
ON CONFLICT (tenant_id, key)
DO UPDATE SET created_by = EXCLUDED.created_by
RETURNING id::text`, templateID, tenantID, "e2e-seed-template", actorID).Scan(&templateID); err != nil {
		if queryErr := tx.QueryRowContext(ctx,
			`SELECT id::text FROM templates_template WHERE tenant_id = $1::uuid AND key = $2`,
			tenantID, "e2e-seed-template",
		).Scan(&templateID); queryErr != nil {
			return "", fmt.Errorf("upsert template: %w", err)
		}
	}

	// content_hash must be a 64-hex value for any non-draft status
	// (chk_template_version_content_hash / chk_template_version_content_hash_non_draft).
	// docx_storage_key is globally unique (uq_templates_template_version_docx_storage_key)
	// and only one published version is allowed per template
	// (uq_one_published_per_template), so key off template_id to stay idempotent.
	docxStorageKey := "templates/e2e-seed/" + templateID + "/body.docx"
	contentHash := strings.Repeat("0", 64)

	if err := tx.QueryRowContext(ctx, `
INSERT INTO templates_template_version (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash, metadata_schema, placeholder_schema, author_id, published_at)
VALUES ($1, $2::uuid, $3, 1, 'published', $4, $5, '{}'::jsonb, '{"placeholders":[]}'::jsonb, $6, now())
ON CONFLICT (template_id, version_number)
DO UPDATE SET status = 'published',
              published_at = now()
RETURNING id::text`, templateVersionID, tenantID, templateID, docxStorageKey, contentHash, actorID,
	).Scan(&templateVersionID); err != nil {
		if queryErr := tx.QueryRowContext(ctx,
			`SELECT id::text FROM templates_template_version WHERE template_id = $1 AND version_number = 1`,
			templateID,
		).Scan(&templateVersionID); queryErr != nil {
			return "", fmt.Errorf("upsert template version: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE templates_template SET published_version_id = $2 WHERE id = $1`,
		templateID, templateVersionID,
	); err != nil {
		return "", fmt.Errorf("update template current version: %w", err)
	}

	return templateVersionID, nil
}

func ensureApprovalRoute(ctx context.Context, tx *sql.Tx, tenantID, actorID, profileCode, areaCode string) error {
	var routeID string
	err := tx.QueryRowContext(ctx, `
INSERT INTO approval_routes (tenant_id, profile_code, name, version, created_by, active)
VALUES ($1, $2, 'E2E Route', 1, $3, TRUE)
ON CONFLICT (tenant_id, profile_code)
DO UPDATE SET name = EXCLUDED.name, active = TRUE
RETURNING id::text`, tenantID, profileCode, actorID).Scan(&routeID)
	if err != nil {
		if scanErr := tx.QueryRowContext(ctx,
			`SELECT id::text FROM approval_routes WHERE tenant_id = $1 AND profile_code = $2`,
			tenantID, profileCode,
		).Scan(&routeID); scanErr != nil {
			return fmt.Errorf("upsert approval route: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM approval_route_stages WHERE route_id = $1`, routeID); err != nil {
		return fmt.Errorf("clear route stages: %w", err)
	}

	// Eligibility resolves user_process_areas.role == the stage's
	// role_in_fixed_area selector's Role exactly (ResolveEligibleActorsForSelectors).
	// 'reviewer' is not a valid area role (decommissioned, ADR 0022) so a
	// 'reviewer' stage is unresolvable — both stages require 'approver', which
	// the seed grants and which carries document.signoff. Selectors is the
	// sole source of truth for a route stage's actor pool (unit 3.2 slice 6b):
	// approval_route_stages.required_role/area_code were dropped by migration
	// 0305, so each stage's eligible-actor pool is seeded as an explicit
	// approval_route_stage_selectors row below instead of flat columns.
	var reviewStageID, approvalStageID string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO approval_route_stages (route_id, stage_order, name, required_capability, quorum, quorum_m, on_eligibility_drift)
VALUES ($1, 1, 'Review', 'document.signoff', 'any_1_of', NULL, 'fail_stage')
RETURNING id::text`, routeID).Scan(&reviewStageID); err != nil {
		return fmt.Errorf("insert review stage: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `
INSERT INTO approval_route_stages (route_id, stage_order, name, required_capability, quorum, quorum_m, on_eligibility_drift)
VALUES ($1, 2, 'Approval', 'document.signoff', 'any_1_of', NULL, 'fail_stage')
RETURNING id::text`, routeID).Scan(&approvalStageID); err != nil {
		return fmt.Errorf("insert approval stage: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO approval_route_stage_selectors (tenant_id, route_stage_id, selector_order, kind, role, area_code)
VALUES
  ($1, $2, 1, 'role_in_fixed_area', 'approver', $4),
  ($1, $3, 1, 'role_in_fixed_area', 'approver', $4)`,
		tenantID, reviewStageID, approvalStageID, areaCode); err != nil {
		return fmt.Errorf("insert route stage selectors: %w", err)
	}

	return nil
}

// ensureControlledDocument creates (or reuses) the controlled_document the draft
// links to via documents.controlled_document_id. The submit path's in-tx profile
// resolution (ADR 0073) reads profile_code from THIS row (not the draft's
// profile_code_snapshot); without the link submit fails with ErrProfileNotConfigured.
// code is a constant — the row is uniquely keyed (tenant_id, profile_code, code),
// both of which are tenant-scoped.
func ensureControlledDocument(ctx context.Context, tx *sql.Tx, tenantID, profileCode, areaCode, ownerID string) (string, error) {
	var cdID string
	if err := tx.QueryRowContext(ctx, `
INSERT INTO controlled_documents (tenant_id, profile_code, process_area_code, code, title, owner_user_id, status, visibility_scope)
VALUES ($1, $2, $3, 'E2E-DOC', 'E2E Controlled Document', $4, 'active', 'company')
ON CONFLICT (tenant_id, profile_code, code)
DO UPDATE SET process_area_code = EXCLUDED.process_area_code,
              title = EXCLUDED.title,
              owner_user_id = EXCLUDED.owner_user_id,
              status = 'active',
              updated_at = now()
RETURNING id::text`, tenantID, profileCode, areaCode, ownerID).Scan(&cdID); err != nil {
		return "", fmt.Errorf("upsert controlled document: %w", err)
	}
	return cdID, nil
}

func upsertDraftDocument(ctx context.Context, tx *sql.Tx, tenantID, docID, templateVersionID, authorID, areaCode, profileCode, controlledDocID, contentHash string) error {
	// process_area_code_snapshot is the area the document.submit/signoff area-grade
	// cap checks resolve against (LoadDocumentAreaCode); a NULL snapshot resolves to
	// "" and fail-closes the cap. Seed it (and the profile snapshot) to the document's
	// own area/profile so the author/approver area memberships authorize the flow.
	// controlled_document_id links the draft to its profile (submit in-tx resolution,
	// ADR 0073); content_hash_at_submit seeds the signoff's active content hash.
	//
	// The enforce_snapshot_on_submit trigger requires the six placeholder/composition/
	// body-docx snapshot columns to be non-NULL on any transition into under_review
	// (which submit performs). The UPDATE in submit_service only touches
	// status/title/version, so these must already be populated on the draft row —
	// seed them with constant placeholders (hashes are bytea, snapshots jsonb).
	if _, err := tx.ExecContext(ctx, `
INSERT INTO documents (id, tenant_id, template_version_id, name, status, form_data_json, created_by, process_area_code_snapshot, profile_code_snapshot, controlled_document_id, content_hash_at_submit,
                       placeholder_schema_snapshot, placeholder_schema_hash, composition_config_snapshot, composition_config_hash, body_docx_snapshot_s3_key, body_docx_hash)
VALUES ($1, $2, $3, 'E2E Draft', 'draft', '{}'::jsonb, $4, $5, $6, $7::uuid, $8,
        '{}'::jsonb, decode(repeat('00',32),'hex'), '{}'::jsonb, decode(repeat('00',32),'hex'), 'seed/body.docx', decode(repeat('00',32),'hex'))
ON CONFLICT (id)
DO UPDATE SET tenant_id = EXCLUDED.tenant_id,
              template_version_id = EXCLUDED.template_version_id,
              name = EXCLUDED.name,
              status = 'draft',
              form_data_json = '{}'::jsonb,
              created_by = EXCLUDED.created_by,
              process_area_code_snapshot = EXCLUDED.process_area_code_snapshot,
              profile_code_snapshot = EXCLUDED.profile_code_snapshot,
              controlled_document_id = EXCLUDED.controlled_document_id,
              content_hash_at_submit = EXCLUDED.content_hash_at_submit,
              placeholder_schema_snapshot = EXCLUDED.placeholder_schema_snapshot,
              placeholder_schema_hash = EXCLUDED.placeholder_schema_hash,
              composition_config_snapshot = EXCLUDED.composition_config_snapshot,
              composition_config_hash = EXCLUDED.composition_config_hash,
              body_docx_snapshot_s3_key = EXCLUDED.body_docx_snapshot_s3_key,
              body_docx_hash = EXCLUDED.body_docx_hash,
              updated_at = now()`, docID, tenantID, templateVersionID, authorID, areaCode, profileCode, controlledDocID, contentHash); err != nil {
		return fmt.Errorf("upsert document: %w", err)
	}
	return nil
}

func createSessionValue(ctx context.Context, tx *sql.Tx, userID, tenantID string) (string, error) {
	secret := strings.TrimSpace(os.Getenv("METALDOCS_AUTH_SESSION_SECRET"))
	if secret == "" {
		return "", fmt.Errorf("METALDOCS_AUTH_SESSION_SECRET is required for e2e seed sessions")
	}

	ttlHours := 12
	if raw := strings.TrimSpace(os.Getenv("METALDOCS_AUTH_SESSION_TTL_HOURS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			ttlHours = parsed
		}
	}

	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	sessionID := hashToken(token)
	signature := signToken(token, secret)
	now := time.Now().UTC()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.auth_sessions (session_id, user_id, tenant_id, created_at, expires_at, revoked_at, ip_address, user_agent, last_seen_at)
VALUES ($1, $2, $3::uuid, $4, $5, NULL, '127.0.0.1', 'e2e-seed', $4)
ON CONFLICT (session_id)
DO UPDATE SET expires_at = EXCLUDED.expires_at,
              revoked_at = NULL,
              last_seen_at = EXCLUDED.last_seen_at`,
		sessionID,
		userID,
		tenantID,
		now,
		now.Add(time.Duration(ttlHours)*time.Hour),
	); err != nil {
		return "", fmt.Errorf("insert auth session: %w", err)
	}

	return token + "." + signature, nil
}

func sanitizeSlug(value string) string {
	out := strings.ToLower(value)
	out = strings.ReplaceAll(out, "_", "")
	out = strings.ReplaceAll(out, "-", "")
	if len(out) > 12 {
		out = out[:12]
	}
	if out == "" {
		return "seed"
	}
	return out
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + strings.ToLower(value[1:])
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func signToken(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func readJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Warn("writeJSON encode error", "err", err)
	}
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func isUndefinedColumn(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42703"
}
