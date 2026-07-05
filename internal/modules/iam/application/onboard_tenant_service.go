// Package application — OnboardTenantService (M7 F7.2 Task C) provisions a
// brand-new tenant end to end in one business transaction: the tenant root
// row, its first system_admin, and that admin's login identity, plus the
// tenant.onboarded audit event. Pattern anchor: AdminService (admin_service.go)
// for the TxRunner.Do + buildAuditEvent + audit.RecordTx shape.
//
// Auth-identity creation: rather than reaching into auth's tables/SQL
// directly (a module-boundary violation) or calling auth's application
// Service.CreateUserWithInput (which owns its own transaction and cannot be
// composed into this service's single tx), this service depends on a locally
// declared, narrow createUserTxRepository interface satisfied structurally by
// *authpg.Repository's already-published CreateUserTx method. This mirrors
// the existing userUpdaterTx precedent in people_service.go (PR-4): the
// interface is declared in iam's application layer, not imported from auth's
// infrastructure package, so there is no import-cycle and no cross-module
// reach into auth's persistence internals — only auth's own exported tx
// method is invoked. Password hashing goes through auth's published seam
// authapp.HashPassword (injected as the PasswordHashFunc port) — iam never
// invokes bcrypt itself, so the cost factor lives in exactly one module.
package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	auditdomain "metaldocs/internal/modules/audit/domain"
	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	"metaldocs/internal/modules/iam/authz"
	iamdomain "metaldocs/internal/modules/iam/domain"
	"metaldocs/internal/platform/db"
	"metaldocs/internal/platform/iamtypes"
	"metaldocs/internal/platform/requesttrace"
)

// PasswordHashFunc is the password-hashing port OnboardTenantService depends
// on. The production implementation is authapp.HashPassword — auth's
// published seam wrapping its own bcrypt mechanism + cost factor — so iam
// never declares crypto parameters of its own.
type PasswordHashFunc func(plain string) (authdomain.PasswordHash, error)

// OnboardTenantInput is the application-layer request to provision a new
// tenant + its first admin identity. AdminPassword is plaintext in transit
// and must never be logged, audited, or echoed back.
type OnboardTenantInput struct {
	Name             string
	Slug             string
	AdminUserID      string
	AdminDisplayName string
	AdminPassword    string
}

// OnboardTenantResult is the successful outcome: the new tenant id (uuid
// text) and the admin's user id (echoed back, never re-derived).
type OnboardTenantResult struct {
	TenantID    string
	AdminUserID string
}

// tenantRepository is the tx-aware persistence port OnboardTenantService uses
// for the tenants-table INSERT. Defined in the application layer (not the
// domain port) for the same reason roleAdminTxRepository is in
// admin_service.go: it names *sql.Tx directly.
type tenantRepository interface {
	InsertTenantTx(ctx context.Context, tx *sql.Tx, tenantID, name, slug string) error
}

// createUserTxRepository is the narrow, locally declared subset of
// authpg.Repository's tx-aware surface OnboardTenantService needs to create
// the admin's login identity inside the shared tx. Mirrors userUpdaterTx in
// people_service.go — declared here (not imported from
// auth/infrastructure/postgres) to avoid a module-boundary / import-cycle
// violation; satisfied structurally by *authpg.Repository.
type createUserTxRepository interface {
	CreateUserTx(ctx context.Context, tx *sql.Tx, params authdomain.CreateUserParams) error
}

// TenantKeyProvisioner is the seam for per-tenant crypto-key provisioning
// (F7.3). F7.2 ships a no-op implementation (NoopTenantKeyProvisioner) so the
// contract's provisioning hook exists at the correct call site without F7.2
// blocking on F7.3's crypto-envelope work.
type TenantKeyProvisioner interface {
	// ProvisionTenantKey is called after the tenant row is inserted, inside
	// the same tx when tx is non-nil. F7.3's real implementation will
	// generate and wrap a per-tenant DEK; key material must never be logged.
	ProvisionTenantKey(ctx context.Context, tx *sql.Tx, tenantID string) error
}

// NoopTenantKeyProvisioner is the F7.2 placeholder — F7.3 replaces it with
// the real per-tenant envelope-key provisioner.
type NoopTenantKeyProvisioner struct{}

// ProvisionTenantKey does nothing (F7.2 seam only).
func (NoopTenantKeyProvisioner) ProvisionTenantKey(context.Context, *sql.Tx, string) error {
	return nil
}

// OnboardTenantService provisions new tenants (M7 F7.2). One TxRunner.Do tx:
// authz.Require(tenant.onboard) -> INSERT tenants -> INSERT iam_users ->
// INSERT iam_user_roles(system_admin) -> auth identity (CreateUserTx) ->
// TenantKeyProvisioner (no-op today) -> audit.RecordTx(tenant.onboarded).
type OnboardTenantService struct {
	tenants      tenantRepository
	createUserTx createUserTxRepository
	runner       db.TxRunner
	audit        auditdomain.Writer
	keys         TenantKeyProvisioner
	hashPassword PasswordHashFunc
}

// NewOnboardTenantService constructs the service. keys may be nil, in which
// case NoopTenantKeyProvisioner is used; hashPassword may be nil, in which
// case auth's published authapp.HashPassword seam is used.
func NewOnboardTenantService(tenants tenantRepository, createUserTx createUserTxRepository, runner db.TxRunner, audit auditdomain.Writer, keys TenantKeyProvisioner, hashPassword PasswordHashFunc) *OnboardTenantService {
	if keys == nil {
		keys = NoopTenantKeyProvisioner{}
	}
	if hashPassword == nil {
		hashPassword = authapp.HashPassword
	}
	return &OnboardTenantService{tenants: tenants, createUserTx: createUserTx, runner: runner, audit: audit, keys: keys, hashPassword: hashPassword}
}

// OnboardTenant provisions a new tenant end to end. actorID is the
// authenticated principal performing the onboarding (the audit ActorID and
// the tier-2 authz.Require actor) and actorTenantID is that principal's OWN
// tenant — grants are tenant-scoped, so authz.Require must resolve them under
// the actor's tenant, never the freshly minted one. Both come from the
// authenticated request context, never the client payload. Returns
// domain.ErrTenantOnboardValidation for malformed input,
// domain.ErrTenantSlugConflict / domain.ErrTenantAdminUserConflict on unique
// violations, and authz.ErrCapDenied when actorID lacks tenant.onboard.
func (s *OnboardTenantService) OnboardTenant(ctx context.Context, in OnboardTenantInput, actorID, actorTenantID string) (OnboardTenantResult, error) {
	name := strings.TrimSpace(in.Name)
	slug := strings.TrimSpace(in.Slug)
	adminUserID := strings.TrimSpace(in.AdminUserID)
	adminDisplayName := strings.TrimSpace(in.AdminDisplayName)
	adminPassword := in.AdminPassword
	actorID = strings.TrimSpace(actorID)
	actorTenantID = strings.TrimSpace(actorTenantID)

	if err := validateOnboardTenantInput(name, slug, adminUserID, adminDisplayName, adminPassword); err != nil {
		return OnboardTenantResult{}, err
	}
	if actorID == "" {
		return OnboardTenantResult{}, fmt.Errorf("%w: actor id required", iamdomain.ErrTenantOnboardValidation)
	}
	if actorTenantID == "" {
		return OnboardTenantResult{}, fmt.Errorf("%w: actor tenant id required", iamdomain.ErrTenantOnboardValidation)
	}
	if s.runner == nil || s.audit == nil || s.tenants == nil || s.createUserTx == nil || s.hashPassword == nil {
		return OnboardTenantResult{}, fmt.Errorf("iam: OnboardTenantService is not fully wired")
	}

	passwordHash, err := s.hashPassword(adminPassword)
	if err != nil {
		return OnboardTenantResult{}, fmt.Errorf("iam: hash admin password: %w", err)
	}

	tenantID := uuid.NewString()

	ev, err := s.buildOnboardedAuditEvent(ctx, tenantID, adminUserID, actorID, name, slug)
	if err != nil {
		return OnboardTenantResult{}, err
	}

	err = s.runner.Do(ctx, func(tx *sql.Tx) error {
		// Tripwire requires the caps asserted BEFORE any guarded INSERT, and
		// grants are tenant-scoped, so both Requires resolve under the
		// ACTOR's tenant (the new tenant has zero roles by definition).
		// tenant.onboard guards the tenants INSERT; user.manage guards the
		// iam_users/iam_user_roles INSERTs this provisioning also performs.
		txCtx := authz.WithCapCache(ctx)
		if err := authz.SeedTxIdentity(txCtx, tx, actorTenantID, actorID); err != nil {
			return fmt.Errorf("seed tenant.onboard authorization: %w", err)
		}
		if err := authz.Require(txCtx, tx, string(iamdomain.CapTenantOnboard), "tenant"); err != nil {
			return fmt.Errorf("require tenant.onboard authorization: %w", err)
		}
		if err := authz.Require(txCtx, tx, string(iamdomain.CapUserManage), "tenant"); err != nil {
			return fmt.Errorf("require user.manage authorization: %w", err)
		}

		// Provisioning rows (tenants/iam_users/iam_user_roles/audit) all
		// carry tenant_id = the NEW tenant, so re-seed the tx-local tenant
		// GUC to it before writing — under FORCE RLS (M7 F7.4) the WITH
		// CHECK clause compares row tenant_id against this GUC. The
		// asserted_caps GUC set by the Requires above is tx-local and
		// persists across this switch.
		if err := authz.SeedTxTenant(txCtx, tx, tenantID); err != nil {
			return fmt.Errorf("seed provisioning tenant: %w", err)
		}

		if err := s.tenants.InsertTenantTx(txCtx, tx, tenantID, name, slug); err != nil {
			return err
		}

		if err := s.keys.ProvisionTenantKey(txCtx, tx, tenantID); err != nil {
			return fmt.Errorf("provision tenant key: %w", err)
		}

		if err := s.createUserTx.CreateUserTx(txCtx, tx, authdomain.CreateUserParams{
			UserID:             adminUserID,
			Username:           adminUserID,
			Email:              adminEmailOrEmpty(adminUserID),
			DisplayName:        adminDisplayName,
			PasswordHash:       authdomain.PasswordHash(string(passwordHash)),
			PasswordAlgo:       "bcrypt",
			MustChangePassword: true,
			IsActive:           true,
			CreatedBy:          actorID,
		}); err != nil {
			if errors.Is(err, authdomain.ErrUserAlreadyExists) {
				return iamdomain.ErrTenantAdminUserConflict
			}
			return fmt.Errorf("create tenant admin identity: %w", err)
		}

		if err := upsertUserAndSystemAdminRoleTx(txCtx, tx, adminUserID, adminDisplayName, tenantID, actorID); err != nil {
			return err
		}

		return s.audit.RecordTx(txCtx, tx, ev)
	})
	if err != nil {
		return OnboardTenantResult{}, err
	}

	return OnboardTenantResult{TenantID: tenantID, AdminUserID: adminUserID}, nil
}

// upsertUserAndSystemAdminRoleTx writes the new admin's iam_users row and its
// system_admin role assignment inside the caller's tx. Inlined here (rather
// than reusing roleAdminTxRepository) because the tripwire-guarded
// iam_users/iam_user_roles writes for a brand-new tenant need no
// DELETE-then-INSERT role replacement semantics (there are zero prior roles)
// and this service does not otherwise depend on domain.RoleAdminRepository.
func upsertUserAndSystemAdminRoleTx(ctx context.Context, tx *sql.Tx, userID, displayName, tenantID, assignedBy string) error {
	const upsertUser = `
INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id, is_active, updated_at)
VALUES ($1, $2, $3::uuid, TRUE, NOW())
ON CONFLICT (user_id)
DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id, is_active = TRUE, updated_at = NOW()
`
	if _, err := tx.ExecContext(ctx, upsertUser, userID, displayName, tenantID); err != nil {
		return fmt.Errorf("upsert tenant admin iam user: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO metaldocs.iam_user_roles (user_id, tenant_id, role_code, assigned_at, assigned_by)
VALUES ($1, $2::uuid, $3, NOW(), $4)
`, userID, tenantID, string(iamtypes.RoleSystemAdmin), assignedBy); err != nil {
		return fmt.Errorf("insert tenant admin role: %w", err)
	}
	return nil
}

// buildOnboardedAuditEvent stamps the tenant.onboarded Event. Payload carries
// only name/slug/admin_user_id — admin_password (plaintext or hashed) is
// NEVER included (contract §1.2).
func (s *OnboardTenantService) buildOnboardedAuditEvent(ctx context.Context, tenantID, adminUserID, actorID, name, slug string) (auditdomain.Event, error) {
	payloadJSON, err := json.Marshal(map[string]any{
		"name":          name,
		"slug":          slug,
		"admin_user_id": adminUserID,
	})
	if err != nil {
		return auditdomain.Event{}, err
	}
	return auditdomain.Event{
		ID:           "evt_" + uuid.NewString(),
		OccurredAt:   time.Now().UTC(),
		ActorID:      actorID,
		Action:       "tenant.onboarded",
		ResourceType: "tenant",
		ResourceID:   tenantID,
		PayloadJSON:  string(payloadJSON),
		TraceID:      requesttrace.Resolve(ctx),
		TenantID:     tenantID,
	}, nil
}

// validateOnboardTenantInput checks blank/malformed fields before any write
// is attempted. Wraps iamdomain.ErrTenantOnboardValidation so handlers can map
// to 400 via errors.Is.
func validateOnboardTenantInput(name, slug, adminUserID, adminDisplayName, adminPassword string) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", iamdomain.ErrTenantOnboardValidation)
	}
	if slug == "" {
		return fmt.Errorf("%w: slug is required", iamdomain.ErrTenantOnboardValidation)
	}
	if !isURLSafeSlug(slug) {
		return fmt.Errorf("%w: slug must be URL-safe (lowercase letters, digits, hyphens)", iamdomain.ErrTenantOnboardValidation)
	}
	if adminUserID == "" {
		return fmt.Errorf("%w: admin_user_id is required", iamdomain.ErrTenantOnboardValidation)
	}
	if adminDisplayName == "" {
		return fmt.Errorf("%w: admin_display_name is required", iamdomain.ErrTenantOnboardValidation)
	}
	if len(adminPassword) < 12 {
		return fmt.Errorf("%w: admin_password must be at least 12 characters", iamdomain.ErrTenantOnboardValidation)
	}
	return nil
}

// isURLSafeSlug reports whether slug contains only lowercase letters, digits,
// and hyphens (matches the tenant slug contract in the spec).
func isURLSafeSlug(slug string) bool {
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

// adminEmailOrEmpty returns userID as an email if it parses as one, else "".
// auth_identities.email is nullable and unique-when-present; admin_user_id is
// the canonical TEXT identifier (memory: tokens-actor-id-text-contract), not
// necessarily an email address.
func adminEmailOrEmpty(userID string) string {
	if _, err := mail.ParseAddress(userID); err == nil {
		return userID
	}
	return ""
}
