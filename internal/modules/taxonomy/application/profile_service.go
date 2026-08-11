package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/taxonomy/domain"
	"metaldocs/internal/platform/authn"
)

// TemplateVersionChecker is the cross-module read port ProfileService uses
// to validate a template version before binding it as a profile's default
// (SetDefaultTemplate). It returns whether the version is published and
// which profile code it belongs to.
type TemplateVersionChecker interface {
	IsPublished(ctx context.Context, versionID string) (bool, string, error)
}

// ProfileService is the use-case orchestrator for DocumentProfile: List,
// Get, Create, Update, Archive, SetDefaultTemplate. Every mutating method
// emits a governance event via govLogger inside the same transaction as the
// mutation (T-005 closed).
type ProfileService struct {
	profiles  domain.ProfileRepository
	tplCheck  TemplateVersionChecker
	govLogger domain.GovernanceLogger
	now       func() time.Time

	// routes is the approval-owned route-readiness port backing the admin
	// has_active_route badge. Wired post-construction via
	// WithRouteReadinessReader so NewProfileService's signature (and its test
	// call sites) stays stable; RouteReadySubjects fails closed when nil.
	routes approvaldomain.RouteReadinessReader
}

// NewProfileService builds a ProfileService with now defaulted to
// time.Now. It panics if govLogger or tplCheck is nil — taxonomy has no
// silent-audit-loss or unchecked-template fallback (fail-loud by design).
func NewProfileService(
	profiles domain.ProfileRepository,
	tplCheck TemplateVersionChecker,
	govLogger domain.GovernanceLogger,
) *ProfileService {
	if govLogger == nil {
		panic("taxonomy: ProfileService govLogger must not be nil")
	}
	if tplCheck == nil {
		panic("taxonomy: ProfileService template checker must not be nil")
	}
	return &ProfileService{profiles: profiles, tplCheck: tplCheck, govLogger: govLogger, now: time.Now}
}

// ErrRouteReadinessUnconfigured is returned by RouteReadySubjects when no
// route-readiness reader is wired — a composition-root bug. Reporting every
// profile as "no route" instead would be a silent fallback that mislabels a
// correctly configured tenant, so this fails closed with an explicit error.
var ErrRouteReadinessUnconfigured = errors.New("taxonomy: route readiness reader not configured")

// WithRouteReadinessReader wires the approval-owned route-readiness port
// post-construction. The taxonomy composition root injects it.
func (s *ProfileService) WithRouteReadinessReader(r approvaldomain.RouteReadinessReader) *ProfileService {
	if r == nil {
		panic("taxonomy: ProfileService route readiness reader must not be nil")
	}
	s.routes = r
	return s
}

// RouteReadiness carries per-subject-kind route readiness for a tenant's
// profiles. Both routes are profile-keyed (ADR 0086), so each set holds profile
// codes and delivery fills has_active_route / has_active_template_route with a
// plain membership test.
type RouteReadiness struct {
	Documents map[string]struct{}
	Templates map[string]struct{}
}

// RouteReadySubjects returns, for tenantID, the profile codes that have an
// active DOCUMENT approval route and those that have an active TEMPLATE one
// (approval_routes, subject_kind=document / template — both keyed on the
// profile code since ADR 0086). It is the single place taxonomy names either
// subject kind.
//
// Two bulk reads, never a per-profile query: the profile listing annotates its
// whole page from these two sets.
//
// Readiness only — taxonomy never sees a route id or any route internals.
func (s *ProfileService) RouteReadySubjects(ctx context.Context, tenantID string) (RouteReadiness, error) {
	if s.routes == nil {
		return RouteReadiness{}, ErrRouteReadinessUnconfigured
	}
	documents, err := s.routes.ActiveRouteSubjectKeys(ctx, tenantID, string(approvaldomain.SubjectKindDocument))
	if err != nil {
		return RouteReadiness{}, fmt.Errorf("taxonomy: read approval route readiness: %w", err)
	}
	templates, err := s.routes.ActiveRouteSubjectKeys(ctx, tenantID, string(approvaldomain.SubjectKindTemplate))
	if err != nil {
		return RouteReadiness{}, fmt.Errorf("taxonomy: read template approval route readiness: %w", err)
	}
	return RouteReadiness{Documents: documents, Templates: templates}, nil
}

// List returns profiles for tenantID, including archived ones when
// includeArchived is true.
func (s *ProfileService) List(ctx context.Context, tenantID string, includeArchived bool) ([]domain.DocumentProfile, error) {
	profiles, err := s.profiles.List(ctx, tenantID, includeArchived)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: list profiles: %w", err)
	}
	return profiles, nil
}

// Get returns the profile identified by (tenantID, code), or a wrapped
// domain.ErrProfileNotFound if no such row exists.
func (s *ProfileService) Get(ctx context.Context, tenantID string, code domain.ProfileCode) (*domain.DocumentProfile, error) {
	profile, err := s.profiles.GetByCode(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: get profile %q: %w", code, err)
	}
	return profile, nil
}

// Create validates p, inserts it, and logs a profile.created governance
// event, all inside a single transaction (BeginTx...Commit); any failure
// rolls the transaction back so no partial profile/event pair is persisted.
func (s *ProfileService) Create(ctx context.Context, p *domain.DocumentProfile) error {
	newProfile, err := domain.NewDocumentProfile(*p)
	if err != nil {
		return fmt.Errorf("taxonomy: validate profile create: %w", err)
	}
	// A3.3 (T1): ActorUserID is the governance-event attribution for this
	// mutation, so an absent actor is a precondition failure, not a late one.
	// It resolves here — after the purely local validation above, before
	// BeginTx — so a request with no principal opens no transaction and calls
	// no repository method at all. "Rolled back" is a weaker property than
	// "never started".
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
	}
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin profile create tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.profiles.CreateTx(ctx, tx, newProfile); err != nil {
		return fmt.Errorf("taxonomy: create profile %q: %w", newProfile.Code, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"code": string(newProfile.Code),
		"name": newProfile.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal profile create governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     newProfile.TenantID,
		EventType:    domain.GovernanceEventTypeProfileCreated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   string(newProfile.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile create governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit profile create tx: %w", err)
	}
	committed = true
	return nil
}

// Update persists p and logs a profile.updated governance event, both
// inside a single transaction; any failure rolls the transaction back.
// Unlike Create, Update does not re-validate p through
// domain.NewDocumentProfile — callers must pass an already-normalized
// profile. The one exception is editable_by_role: it is a bound role
// vocabulary (F-QA4-2), so BOTH write paths validate it and an update cannot
// smuggle a free-text role past the constructor that Create runs.
func (s *ProfileService) Update(ctx context.Context, p *domain.DocumentProfile) error {
	if err := domain.ValidateEditableByRole(p.EditableByRole); err != nil {
		return fmt.Errorf("taxonomy: validate profile update: %w", err)
	}
	// A3.3 (T1): resolved before BeginTx / UpdateTx, so an actorless request
	// reaches no repository mutation. See Create.
	actorUserID, err := authn.RequireUserID(ctx)
	if err != nil {
		return err
	}
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin profile update tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := s.profiles.UpdateTx(ctx, tx, p); err != nil {
		return fmt.Errorf("taxonomy: update profile %q: %w", p.Code, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"code": string(p.Code),
		"name": p.Name,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal profile update governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     p.TenantID,
		EventType:    domain.GovernanceEventTypeProfileUpdated,
		ActorUserID:  actorUserID,
		ResourceType: "document_profile",
		ResourceID:   string(p.Code),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile update governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit profile update tx: %w", err)
	}
	committed = true
	return nil
}

// Reclassify changes a profile's governance_class (G1). It locks the profile
// row (FOR UPDATE), rejects archived profiles, and is a no-op (no event) when
// the class is unchanged. Otherwise it writes only the governance_class column
// and logs a profile.governance_class_change governance event with the
// from/to classes — all inside one transaction. A reclassification that would
// leave an active approval route violating the new class is rejected by the DB
// reclassify-guard trigger and surfaces at Commit as
// domain.ErrClassChangeRouteConflict (mapped in the repository), rolling the
// transaction (and its audit event) back.
func (s *ProfileService) Reclassify(ctx context.Context, tenantID string, profileCode domain.ProfileCode, newClass domain.GovernanceClass, actorID string) error {
	if err := newClass.Validate(); err != nil {
		return fmt.Errorf("taxonomy: validate reclassify %q: %w", profileCode, err)
	}
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin reclassify tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock profile %q: %w", profileCode, err)
	}
	if !profile.IsActive() {
		return domain.ErrProfileArchived
	}

	oldClass := profile.GovernanceClass
	if oldClass == "" {
		oldClass = domain.GovernanceControlado // normalize legacy rows to the column default
	}
	if oldClass == newClass {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("taxonomy: commit reclassify no-op %q: %w", profileCode, err)
		}
		committed = true
		return nil
	}

	if err := s.profiles.SetGovernanceClassTx(ctx, tx, tenantID, profileCode, newClass); err != nil {
		return fmt.Errorf("taxonomy: set governance_class %q: %w", profileCode, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"from": string(oldClass),
		"to":   string(newClass),
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal reclassify governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileGovernanceClassChange,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   string(profileCode),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log reclassify governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit reclassify %q: %w", profileCode, err)
	}
	committed = true
	return nil
}

// SetDefaultTemplate locks the profile row (FOR UPDATE), rejects archived
// profiles with domain.ErrProfileArchived, validates the template version
// via tplCheck (domain.ErrTemplateNotPublished if unpublished,
// domain.ErrTemplateProfileMismatch if it belongs to a different profile),
// then persists DefaultTemplateVersionID and logs a
// profile.default_template_change governance event — all inside one
// transaction.
func (s *ProfileService) SetDefaultTemplate(ctx context.Context, tenantID string, profileCode domain.ProfileCode, templateVersionID, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin set default template tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock profile %q: %w", profileCode, err)
	}
	if !profile.IsActive() {
		return domain.ErrProfileArchived
	}

	published, templateProfileCode, err := s.tplCheck.IsPublished(ctx, templateVersionID)
	if err != nil {
		return fmt.Errorf("taxonomy: check template version %q: %w", templateVersionID, err)
	}
	if !published {
		return domain.ErrTemplateNotPublished
	}
	if templateProfileCode != string(profileCode) {
		return domain.ErrTemplateProfileMismatch
	}

	profile.DefaultTemplateVersionID = &templateVersionID
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return fmt.Errorf("taxonomy: update profile default template %q: %w", profileCode, err)
	}
	payload, err := marshalGovernancePayload(map[string]string{
		"template_version_id": templateVersionID,
	})
	if err != nil {
		return fmt.Errorf("taxonomy: marshal default template governance payload: %w", err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileDefaultTemplateChange,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   string(profileCode),
		PayloadJSON:  payload,
	}); err != nil {
		return fmt.Errorf("taxonomy: log default template governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit set default template tx: %w", err)
	}
	committed = true
	return nil
}

// Archive locks the profile row (FOR UPDATE), soft-archives it via
// DocumentProfile.Archive, persists the change, and logs a
// profile.archived governance event — all inside one transaction. Returns
// domain.ErrProfileArchived if the profile is already archived.
func (s *ProfileService) Archive(ctx context.Context, tenantID string, profileCode domain.ProfileCode, actorID string) error {
	tx, err := s.profiles.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("taxonomy: begin archive profile tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	profile, err := s.profiles.GetByCodeForUpdate(ctx, tx, tenantID, profileCode)
	if err != nil {
		return fmt.Errorf("taxonomy: lock profile %q: %w", profileCode, err)
	}
	if err := profile.Archive(s.now()); err != nil {
		return fmt.Errorf("taxonomy: archive profile %q: %w", profileCode, err)
	}
	if err := s.profiles.UpdateTx(ctx, tx, profile); err != nil {
		return fmt.Errorf("taxonomy: update archived profile %q: %w", profileCode, err)
	}
	if err := s.govLogger.LogTx(ctx, tx, domain.GovernanceEvent{
		TenantID:     tenantID,
		EventType:    domain.GovernanceEventTypeProfileArchived,
		ActorUserID:  actorID,
		ResourceType: "document_profile",
		ResourceID:   string(profileCode),
		PayloadJSON:  []byte(`{}`),
	}); err != nil {
		return fmt.Errorf("taxonomy: log profile archive governance event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("taxonomy: commit archive profile tx: %w", err)
	}
	committed = true
	return nil
}
