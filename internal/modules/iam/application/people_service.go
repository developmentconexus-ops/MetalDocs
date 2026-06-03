// Package application — PeopleService is the People-tab orchestrator authored
// in PR-4 of the IAM Admin Center rebuild. It composes:
//   - the auth Service (identity creation / password generation / metadata update)
//   - the role admin repository (single-tenant-role assignment)
//   - the area membership service (multi-area grants)
//
// Atomicity note (see PR-4 spec §"Architectural choices"): InviteUser writes
// the auth identity + tenant role in a single transaction via the existing
// CreateUserWithInput tx-aware path. Area memberships use the membership
// service, which writes one tx per area. We validate area codes before
// committing the user; if a post-commit grant fails we surface the failure
// with the partially-created userId so the operator can retry from the UI.
// This is a deliberate, narrow deviation from the spec's "single tx for
// user+role+memberships" because extending the auth Service's two-actor tx
// pattern to a third actor would require a cross-package tx-handle contract
// that is out of scope for PR-4 and PR-5's owner (Roles & Caps matrix).
package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	authapp "metaldocs/internal/modules/auth/application"
	authdomain "metaldocs/internal/modules/auth/domain"
	iamdomain "metaldocs/internal/modules/iam/domain"
)

// ErrPeopleValidation is returned for client-side input violations that the
// handler maps to a 400 Problem with code VALIDATION_ERROR.
var ErrPeopleValidation = errors.New("people_validation_error")

// ErrAreaUnknown signals a referenced areaCode does not exist in process_areas
// for the tenant. Maps to a 400 with detail listing the missing codes.
var ErrAreaUnknown = errors.New("area_unknown")

// ErrUserNotInTenant signals the target userID is not a member of the caller's
// tenant. Handlers MUST map this to 404 to avoid leaking the existence of
// users in other tenants.
var ErrUserNotInTenant = errors.New("user_not_in_tenant")

// ErrCursorExpired signals the cursor anchor's user is no longer present in
// the filtered set (deleted / changed role / etc). Handlers map to 410 Gone
// with code CURSOR_EXPIRED so clients restart pagination explicitly instead
// of silently resetting to page 1 and emitting duplicate rows.
var ErrCursorExpired = errors.New("cursor_expired")

// AreaCatalogReader lets PeopleService verify that an areaCode supplied in an
// invite exists before we burn an auth identity on it. PR-4 supplies a no-op
// implementation for in-memory tests; the postgres impl reads process_areas.
type AreaCatalogReader interface {
	AreaCodeExists(ctx context.Context, tenantID, areaCode string) (bool, error)
}

// PermissiveAreaCatalog accepts every areaCode without a database check. Used
// only by the in-memory test surface; the real wiring goes through a
// Postgres-backed reader.
type PermissiveAreaCatalog struct{}

func (PermissiveAreaCatalog) AreaCodeExists(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}

// ListedUser mirrors the subset of fields the People-tab table column model
// needs. We do NOT reuse authdomain.ManagedUser directly because that type's
// "Roles" field is the legacy multi-role view; the new contract is one tenant
// role plus N area memberships.
type ListedUser struct {
	UserID              string
	Username            string
	Email               string
	DisplayName         string
	IsActive            bool
	MustChangePassword  bool
	LastLoginAt         *time.Time
	LockedUntil         *time.Time
	FailedLoginAttempts int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	TenantRole          iamdomain.Role
	AreaMemberships     []iamdomain.UserProcessArea
}

// ListFilters mirrors the OpenAPI ListUsersParams. Limit is bounded server-side.
type ListFilters struct {
	IsActive *bool
	Role     *iamdomain.Role
	AreaCode *string
	Q        *string
	Cursor   *string
	Limit    int
}

// ListResult is the paginated payload. Total is optional — we compute it only
// when the underlying repository can do so cheaply.
type ListResult struct {
	Items      []ListedUser
	NextCursor string
	HasMore    bool
	Total      *int
}

// BulkOutcome is the per-user disposition emitted by BulkAction.
type BulkOutcome struct {
	Succeeded []string
	Failed    []BulkFailure
}

type BulkFailure struct {
	UserID  string
	Code    string
	Message string
}

// PeopleAuthService is the test-visible alias of the auth-service contract
// PeopleService depends on. Exported in PR-4 so tests/unit/iam_people can
// inject a lightweight fake without standing up bcrypt; production code
// continues to use NewPeopleService with the concrete *authapp.Service.
type PeopleAuthService = peopleAuthService

// PeopleRoleProvider mirrors PeopleAuthService — test-visible alias.
type PeopleRoleProvider = peopleRoleProvider

// PeopleMembershipService mirrors PeopleAuthService — test-visible alias.
type PeopleMembershipService = peopleMembershipService

// PeopleServiceFromInterfaces is the test-only constructor that lets external
// packages compose a PeopleService from interface-typed dependencies. The
// production constructor (NewPeopleService) still takes a concrete
// *authapp.Service so this back door does not bleed into application wiring.
func PeopleServiceFromInterfaces(
	auth PeopleAuthService,
	roles PeopleRoleProvider,
	roleAdmin iamdomain.RoleAdminRepository,
	memberships PeopleMembershipService,
	areaCatalog AreaCatalogReader,
	invalidator RoleCacheInvalidator,
) *PeopleService {
	return newPeopleServiceFromIfaces(auth, roles, roleAdmin, memberships, areaCatalog, invalidator)
}

// peopleAuthService is the subset of authapp.Service we depend on. Defined as
// an interface so the unit test can stand a fake without spinning up bcrypt.
type peopleAuthService interface {
	CreateUserWithInput(ctx context.Context, input authdomain.CreateUserInput) error
	UpdateUser(ctx context.Context, params authdomain.UpdateUserParams, newPassword string) error
	AdminResetPassword(ctx context.Context, userID, newPassword string) error
	UnlockUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, tenantID string) ([]authdomain.ManagedUser, error)
}

// peopleRoleProvider resolves a user's single canonical tenant role.
type peopleRoleProvider interface {
	RolesByUserID(ctx context.Context, userID, tenantID string) ([]iamdomain.Role, error)
}

// peopleMembershipService is the subset of AreaMembershipService we need.
type peopleMembershipService interface {
	Grant(ctx context.Context, userID, tenantID, areaCode string, role iamdomain.Role, grantedBy string) error
	ListActive(ctx context.Context, userID, tenantID string) ([]iamdomain.UserProcessArea, error)
}

// PeopleService is the People-tab orchestrator (PR-4).
type PeopleService struct {
	auth         peopleAuthService
	roles        peopleRoleProvider
	roleAdmin    iamdomain.RoleAdminRepository
	memberships  peopleMembershipService
	areaCatalog  AreaCatalogReader
	invalidator  RoleCacheInvalidator
	tempPassword func() (string, error)
	nowFn        func() time.Time
}

// NewPeopleService wires the People orchestrator. memberships may be nil in
// configurations without the area membership service (memory dev mode); the
// invite path then skips membership writes and returns an error if any are
// requested.
func NewPeopleService(
	auth *authapp.Service,
	roles peopleRoleProvider,
	roleAdmin iamdomain.RoleAdminRepository,
	memberships *AreaMembershipService,
	areaCatalog AreaCatalogReader,
	invalidator RoleCacheInvalidator,
) *PeopleService {
	var mIface peopleMembershipService
	if memberships != nil {
		mIface = memberships
	}
	return newPeopleServiceFromIfaces(auth, roles, roleAdmin, mIface, areaCatalog, invalidator)
}

func newPeopleServiceFromIfaces(
	auth peopleAuthService,
	roles peopleRoleProvider,
	roleAdmin iamdomain.RoleAdminRepository,
	memberships peopleMembershipService,
	areaCatalog AreaCatalogReader,
	invalidator RoleCacheInvalidator,
) *PeopleService {
	if areaCatalog == nil {
		areaCatalog = PermissiveAreaCatalog{}
	}
	return &PeopleService{
		auth:         auth,
		roles:        roles,
		roleAdmin:    roleAdmin,
		memberships:  memberships,
		areaCatalog:  areaCatalog,
		invalidator:  invalidator,
		tempPassword: generateTempPassword,
		nowFn:        func() time.Time { return time.Now().UTC() },
	}
}

// WithTempPasswordGenerator lets tests inject a deterministic generator.
func (s *PeopleService) WithTempPasswordGenerator(fn func() (string, error)) *PeopleService {
	if fn != nil {
		s.tempPassword = fn
	}
	return s
}

// InviteInput is the normalized server-side projection of the OpenAPI
// UserInviteRequest. The handler is responsible for decoding the wire body
// into this shape so the service has no JSON dependency.
type InviteInput struct {
	Username        string
	Email           string
	DisplayName     string
	TenantRole      iamdomain.Role
	AreaMemberships []InviteAreaInput
}

type InviteAreaInput struct {
	AreaCode string
	Role     iamdomain.Role
}

// InviteResult is the one-time payload returned to the operator. TempPassword
// MUST NOT be persisted in audit payloads or log lines.
type InviteResult struct {
	UserID       string
	TempPassword string
}

// Invite provisions a new tenant-managed user. Validation is performed before
// any write so that a malformed request never leaves a half-created identity.
func (s *PeopleService) Invite(ctx context.Context, tenantID, actorID string, input InviteInput) (InviteResult, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return InviteResult{}, fmt.Errorf("%w: tenant required", ErrPeopleValidation)
	}
	if err := s.validateInvite(ctx, tenantID, input); err != nil {
		return InviteResult{}, err
	}

	tempPwd, err := s.tempPassword()
	if err != nil {
		return InviteResult{}, fmt.Errorf("generate temp password: %w", err)
	}
	userID := strings.TrimSpace(input.Username)
	// CreateUserWithInput runs INSERT auth_identity + INSERT iam_user_roles in a
	// single tx when the repo implements the tx-aware interfaces (postgres). The
	// memory repo falls through to the sequential path; both are correct.
	createErr := s.auth.CreateUserWithInput(ctx, authdomain.CreateUserInput{
		UserID:      authdomain.UserID(userID),
		Username:    input.Username,
		Email:       input.Email,
		DisplayName: input.DisplayName,
		Password:    authdomain.PlainPassword(tempPwd),
		TenantID:    authdomain.TenantID(tenantID),
		Roles:       []iamdomain.Role{input.TenantRole},
		CreatedBy:   strings.TrimSpace(actorID),
	})
	if createErr != nil {
		return InviteResult{}, createErr
	}
	if s.invalidator != nil {
		s.invalidator.InvalidateUserTenant(userID, tenantID)
	}

	// Membership grants happen post-commit. If a grant fails we surface the
	// failure with the userId so the operator can retry the membership grant.
	// We deliberately do not delete the partially-created user — that would
	// require a tx-handle contract the auth Service does not expose. Leaving
	// the user in place lets the operator finish the grant from the area
	// memberships UI instead of restarting from invite. See package comment.
	if len(input.AreaMemberships) > 0 {
		if s.memberships == nil {
			return InviteResult{UserID: userID, TempPassword: tempPwd},
				fmt.Errorf("%w: area memberships requested but membership service is not wired", ErrPeopleValidation)
		}
		for _, m := range input.AreaMemberships {
			if err := s.memberships.Grant(ctx, userID, tenantID, m.AreaCode, m.Role, strings.TrimSpace(actorID)); err != nil {
				return InviteResult{UserID: userID, TempPassword: tempPwd},
					fmt.Errorf("grant membership %s: %w", m.AreaCode, err)
			}
		}
	}

	return InviteResult{UserID: userID, TempPassword: tempPwd}, nil
}

func (s *PeopleService) validateInvite(ctx context.Context, tenantID string, input InviteInput) error {
	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("%w: username required", ErrPeopleValidation)
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		return fmt.Errorf("%w: displayName required", ErrPeopleValidation)
	}
	if email := strings.TrimSpace(input.Email); email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return fmt.Errorf("%w: email %q is not a valid address", ErrPeopleValidation, email)
		}
	}
	if !iamdomain.IsValidRole(input.TenantRole) {
		return fmt.Errorf("%w: tenantRole %q is not in canonical 8", ErrPeopleValidation, input.TenantRole)
	}
	for _, m := range input.AreaMemberships {
		if strings.TrimSpace(m.AreaCode) == "" {
			return fmt.Errorf("%w: areaCode required", ErrPeopleValidation)
		}
		if !isAreaScopedRole(m.Role) {
			return fmt.Errorf("%w: area role %q must be one of signer|area_admin|qms_admin", ErrPeopleValidation, m.Role)
		}
		ok, err := s.areaCatalog.AreaCodeExists(ctx, tenantID, m.AreaCode)
		if err != nil {
			return fmt.Errorf("verify area %q: %w", m.AreaCode, err)
		}
		if !ok {
			return fmt.Errorf("%w: area %q does not exist", ErrAreaUnknown, m.AreaCode)
		}
	}
	return nil
}

// isAreaScopedRole enforces the spec rule that area-membership roles are
// drawn from the three area-only canonical roles. The tenant role enum is
// broader (8 values); area roles are a strict subset.
func isAreaScopedRole(r iamdomain.Role) bool {
	switch r {
	case iamdomain.RoleSigner, iamdomain.RoleAreaAdmin, iamdomain.RoleQmsAdmin:
		return true
	}
	return false
}

// PatchInput mirrors UpdateManagedUserRequest. TenantRole, if provided, drives
// a second non-tx step that replaces the user's tenant role. See package
// comment for the atomicity caveat.
type PatchInput struct {
	DisplayName        *string
	Email              *string
	IsActive           *bool
	MustChangePassword *bool
	TenantRole         *iamdomain.Role
}

// PatchAtomic applies a partial update to a managed user. "Atomic" here means
// that metadata + role flip are applied in sequence and roll forward if both
// succeed; the limitation vs. a true single-tx atomic update is documented in
// the package comment.
func (s *PeopleService) PatchAtomic(ctx context.Context, tenantID, actorID, userID string, input PatchInput) (map[string]any, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	if tenantID == "" || userID == "" {
		return nil, fmt.Errorf("%w: tenant and userId required", ErrPeopleValidation)
	}
	changes := map[string]any{}

	if input.Email != nil {
		if trimmed := strings.TrimSpace(*input.Email); trimmed != "" {
			if _, err := mail.ParseAddress(trimmed); err != nil {
				return nil, fmt.Errorf("%w: email %q is not a valid address", ErrPeopleValidation, trimmed)
			}
			*input.Email = trimmed
		}
	}
	if input.TenantRole != nil && !iamdomain.IsValidRole(*input.TenantRole) {
		return nil, fmt.Errorf("%w: tenantRole %q is not in canonical 8", ErrPeopleValidation, *input.TenantRole)
	}

	if hasMetadataUpdate(input) {
		params := authdomain.UpdateUserParams{
			UserID:             userID,
			DisplayName:        input.DisplayName,
			Email:              input.Email,
			IsActive:           input.IsActive,
			MustChangePassword: input.MustChangePassword,
		}
		if err := s.auth.UpdateUser(ctx, params, ""); err != nil {
			return nil, err
		}
		if input.DisplayName != nil {
			changes["displayName"] = *input.DisplayName
		}
		if input.Email != nil {
			changes["email"] = *input.Email
		}
		if input.IsActive != nil {
			changes["isActive"] = *input.IsActive
		}
		if input.MustChangePassword != nil {
			changes["mustChangePassword"] = *input.MustChangePassword
		}
	}

	if input.TenantRole != nil {
		displayName := strings.TrimSpace(userID)
		if input.DisplayName != nil {
			displayName = strings.TrimSpace(*input.DisplayName)
		}
		// roleAdmin.ReplaceUserRoles already calls authz.Require(CapUserManage)
		// inside its INSERT tx (T-004 already closed in postgres/role_admin_repository.go).
		if err := s.roleAdmin.ReplaceUserRoles(ctx, userID, displayName, tenantID, *input.TenantRole, strings.TrimSpace(actorID)); err != nil {
			return nil, err
		}
		if s.invalidator != nil {
			s.invalidator.InvalidateUserTenant(userID, tenantID)
		}
		changes["tenantRole"] = string(*input.TenantRole)
	}

	return changes, nil
}

func hasMetadataUpdate(input PatchInput) bool {
	return input.DisplayName != nil || input.Email != nil || input.IsActive != nil || input.MustChangePassword != nil
}

// BulkAction applies a per-user mutation, isolating failures so one bad userId
// does not block the rest. force-logout is intentionally not implemented in
// PR-4; PR-7 will provide it once the session-manage path lands.
func (s *PeopleService) BulkAction(ctx context.Context, tenantID, actorID, action string, userIDs []string) (BulkOutcome, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return BulkOutcome{}, fmt.Errorf("%w: tenant required", ErrPeopleValidation)
	}
	out := BulkOutcome{Succeeded: []string{}, Failed: []BulkFailure{}}
	if action == "force-logout" {
		// Spec: PR-7 dependency. Surface 501 NOT_IMPLEMENTED at the handler.
		return out, errForceLogoutDeferred
	}
	for _, raw := range userIDs {
		userID := strings.TrimSpace(raw)
		if userID == "" {
			out.Failed = append(out.Failed, BulkFailure{UserID: raw, Code: "VALIDATION_ERROR", Message: "userId required"})
			continue
		}
		if err := s.VerifyUserInTenant(ctx, tenantID, userID); err != nil {
			out.Failed = append(out.Failed, BulkFailure{UserID: userID, Code: "NOT_FOUND", Message: "User not found"})
			continue
		}
		if err := s.applyBulkAction(ctx, action, userID); err != nil {
			out.Failed = append(out.Failed, BulkFailure{UserID: userID, Code: bulkErrorCode(err), Message: err.Error()})
			continue
		}
		out.Succeeded = append(out.Succeeded, userID)
	}
	return out, nil
}

// errForceLogoutDeferred is returned by BulkAction when callers request the
// force-logout action that PR-7 will implement. The handler maps it to a 501
// Problem so the frontend can show a deferred-feature notice.
var errForceLogoutDeferred = errors.New("force_logout_not_implemented")

func IsForceLogoutDeferred(err error) bool { return errors.Is(err, errForceLogoutDeferred) }

func (s *PeopleService) applyBulkAction(ctx context.Context, action, userID string) error {
	switch action {
	case "activate":
		isActive := true
		return s.auth.UpdateUser(ctx, authdomain.UpdateUserParams{UserID: userID, IsActive: &isActive}, "")
	case "deactivate":
		isActive := false
		return s.auth.UpdateUser(ctx, authdomain.UpdateUserParams{UserID: userID, IsActive: &isActive}, "")
	case "unlock":
		return s.auth.UnlockUser(ctx, userID)
	default:
		return fmt.Errorf("%w: unknown action %q", ErrPeopleValidation, action)
	}
}

func bulkErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrPeopleValidation):
		return "VALIDATION_ERROR"
	case errors.Is(err, authdomain.ErrIdentityNotFound):
		return "NOT_FOUND"
	case errors.Is(err, authdomain.ErrPasswordPolicy):
		return "VALIDATION_ERROR"
	}
	return "INTERNAL_ERROR"
}

// ListFiltered returns paginated users for the People table. The current
// implementation reads via auth.ListUsers (which already joins the tenant
// role) and applies filters + cursor pagination in Go. This is acceptable for
// PR-4 because the dev/QA dataset is small; PR-5 or PR-11 may push the filter
// into a dedicated SQL projection if the dataset grows.
func (s *PeopleService) ListFiltered(ctx context.Context, tenantID string, filters ListFilters) (ListResult, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ListResult{}, fmt.Errorf("%w: tenant required", ErrPeopleValidation)
	}
	limit := filters.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	managed, err := s.auth.ListUsers(ctx, tenantID)
	if err != nil {
		return ListResult{}, err
	}
	listed := make([]ListedUser, 0, len(managed))
	for _, m := range managed {
		tenantRole := iamdomain.RoleViewer
		if len(m.Roles) > 0 {
			tenantRole = m.Roles[0]
		}
		var areas []iamdomain.UserProcessArea
		if s.memberships != nil {
			areas, _ = s.memberships.ListActive(ctx, m.UserID, tenantID)
		}
		listed = append(listed, ListedUser{
			UserID:              m.UserID,
			Username:            m.Username,
			Email:               m.Email,
			DisplayName:         m.DisplayName,
			IsActive:            m.IsActive,
			MustChangePassword:  m.MustChangePassword,
			LastLoginAt:         m.LastLoginAt,
			LockedUntil:         m.LockedUntil,
			FailedLoginAttempts: m.FailedLoginAttempts,
			CreatedAt:           m.CreatedAt,
			UpdatedAt:           m.UpdatedAt,
			TenantRole:          tenantRole,
			AreaMemberships:     areas,
		})
	}
	filtered := applyPeopleFilters(listed, filters)
	cursorIdx := 0
	if filters.Cursor != nil && strings.TrimSpace(*filters.Cursor) != "" {
		idx, found := decodeCursorIndex(*filters.Cursor, filtered)
		if !found {
			return ListResult{}, ErrCursorExpired
		}
		cursorIdx = idx
	}
	if cursorIdx < 0 {
		cursorIdx = 0
	}
	end := cursorIdx + limit
	hasMore := false
	if end < len(filtered) {
		hasMore = true
	} else {
		end = len(filtered)
	}
	page := filtered[cursorIdx:end]
	next := ""
	if hasMore && len(page) > 0 {
		next = encodeCursor(page[len(page)-1])
	}
	total := len(filtered)
	return ListResult{
		Items:      page,
		NextCursor: next,
		HasMore:    hasMore,
		Total:      &total,
	}, nil
}

// VerifyUserInTenant returns nil if userID is a member of tenantID per the
// auth ListUsers projection (which joins iam_users tenant_id). Returns
// ErrUserNotInTenant on miss. Used as a tenant-membership guard before
// delegating to tenant-agnostic auth mutations (reset password, unlock).
func (s *PeopleService) VerifyUserInTenant(ctx context.Context, tenantID, userID string) error {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	if tenantID == "" || userID == "" {
		return ErrUserNotInTenant
	}
	users, err := s.auth.ListUsers(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.UserID == userID {
			return nil
		}
	}
	return ErrUserNotInTenant
}

// ListMemberships returns the active area memberships for a single user. Mirrors
// the OpenAPI ListMembershipsResponse the People-tab drawer consumes.
func (s *PeopleService) ListMemberships(ctx context.Context, tenantID, userID string) ([]iamdomain.UserProcessArea, error) {
	tenantID = strings.TrimSpace(tenantID)
	userID = strings.TrimSpace(userID)
	if tenantID == "" || userID == "" {
		return nil, fmt.Errorf("%w: tenant and userId required", ErrPeopleValidation)
	}
	if s.memberships == nil {
		return []iamdomain.UserProcessArea{}, nil
	}
	return s.memberships.ListActive(ctx, userID, tenantID)
}

func applyPeopleFilters(in []ListedUser, f ListFilters) []ListedUser {
	out := make([]ListedUser, 0, len(in))
	q := ""
	if f.Q != nil {
		q = strings.ToLower(strings.TrimSpace(*f.Q))
	}
	for _, item := range in {
		if f.IsActive != nil && item.IsActive != *f.IsActive {
			continue
		}
		if f.Role != nil && item.TenantRole != *f.Role {
			continue
		}
		if f.AreaCode != nil && strings.TrimSpace(*f.AreaCode) != "" {
			if !hasArea(item.AreaMemberships, *f.AreaCode) {
				continue
			}
		}
		if q != "" {
			hay := strings.ToLower(item.DisplayName + " " + item.Username + " " + item.Email)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		out = append(out, item)
	}
	// We intentionally avoid sort.Slice here — auth.ListUsers already returns
	// rows ORDER BY created_at DESC, which gives the People-tab table its
	// "newest first" ordering by default. Future SQL-side filter projection
	// (PR-5/PR-11) will own deterministic sort.
	return out
}

func hasArea(memberships []iamdomain.UserProcessArea, areaCode string) bool {
	for _, m := range memberships {
		if m.AreaCode == areaCode {
			return true
		}
	}
	return false
}

func encodeCursor(u ListedUser) string {
	return u.UserID
}

// decodeCursorIndex returns (index after the cursor's anchor row, true) when
// the anchor is still in the filtered slice. Returns (0, false) when the
// anchor is missing — caller surfaces ErrCursorExpired so the client restarts
// pagination explicitly instead of silently returning page 1 (which causes
// duplicate rows on append).
func decodeCursorIndex(cursor string, in []ListedUser) (int, bool) {
	for i, item := range in {
		if item.UserID == cursor {
			return i + 1, true
		}
	}
	return 0, false
}

// generateTempPassword returns a 16-character password with at least one
// uppercase, lowercase, digit, and symbol. Generated via crypto/rand.
// Ambiguous characters (0/O, 1/l/I) are excluded so operators reading the
// one-time value over the phone do not mis-type.
func generateTempPassword() (string, error) {
	const (
		uppers   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lowers   = "abcdefghijkmnopqrstuvwxyz"
		digits   = "23456789"
		symbols  = "!@#$%^&*-_=+?"
		alphabet = uppers + lowers + digits + symbols
		length   = 16
	)
	pick := func(pool string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
		if err != nil {
			return 0, err
		}
		return pool[n.Int64()], nil
	}
	buf := make([]byte, length)
	pools := [...]string{uppers, lowers, digits, symbols}
	for i, pool := range pools {
		ch, err := pick(pool)
		if err != nil {
			return "", err
		}
		buf[i] = ch
	}
	for i := len(pools); i < length; i++ {
		ch, err := pick(alphabet)
		if err != nil {
			return "", err
		}
		buf[i] = ch
	}
	// Shuffle so the policy bytes don't always live in the first four slots.
	for i := length - 1; i > 0; i-- {
		jIdx, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := jIdx.Int64()
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf), nil
}
