//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	cddomain "metaldocs/internal/modules/controlleddocuments/domain"
	cdinfra "metaldocs/internal/modules/controlleddocuments/infrastructure"
	domain "metaldocs/internal/modules/approval/domain"
)

// factory.go — unified integration-test fixture builders, built ON the testdb
// harness (template-DB-cloned-per-test) and generalizing fixtures.go.
//
// Every builder: starts from sane defaults, applies functional WithX options,
// mints fresh UUIDs and per-call-unique taxonomy codes (so two builds in one
// database never fight over a globally-unique taxonomy code), seeds the FK
// parents it needs (auto-wiring missing associations), asserts the REAL
// tripwire capability via seedWithCaps (tx-local, pool-safe — never weakens or
// disables the tripwire), and returns a struct carrying the IDs/columns tests
// assert on.
//
// Schema names are written explicitly (public. / metaldocs.) — matching the
// runtime search path and the F4c.4 grep-guard (no bare unqualified table
// names in test code).

// ---------------------------------------------------------------------------
// Returned fixtures
// ---------------------------------------------------------------------------

type Tenant struct {
	ID string
}

type User struct {
	ID          string
	DisplayName string
	TenantID    string
}

type Taxonomy struct {
	TenantID        string
	FamilyCode      string
	ProcessAreaCode string
	ProfileCode     string
}

type ControlledDoc struct {
	ID              string
	TenantID        string
	Code            string
	ProfileCode     string
	ProcessAreaCode string
	OwnerUserID     string
	Status          string
}

type Document struct {
	ID                   string
	TenantID             string
	ControlledDocumentID string
	TemplateVersionID    string
	Owner                string
	Status               string
	RevisionNumber       int
	RevisionVersion      int
	ScheduleGeneration   int64
}

type ApprovalRoute struct {
	ID          string
	TenantID    string
	ProfileCode string
	SubjectKind string
}

type ApprovalInstance struct {
	ID                string
	TenantID          string
	DocumentID        string
	RouteID           string
	Status            string
	FrozenContentHash *string
}

type Notification struct {
	ID              string
	TenantID        string
	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string
	Status          string
}

// ---------------------------------------------------------------------------
// Functional options (shared Spec so the generic With* names work across
// builders; each builder reads the fields it cares about)
// ---------------------------------------------------------------------------

type Spec struct {
	TenantID      string
	UserID        string
	DisplayName   string
	Role          string
	Taxonomy      *Taxonomy
	ControlledDoc *ControlledDoc
	Document      *Document
	Route         *ApprovalRoute

	OwnerUserID       string
	TemplateVersionID string
	Name              string
	Status            string
	Code              string
	ProfileCode       string
	GovernanceClass   string
	SubjectKind       string
	StageSeeder       func(tx *sql.Tx, routeID string) error

	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string

	RevisionNumber       int
	RevisionVersion      int
	ScheduleGeneration   int64
	EffectiveFrom        time.Time
	PlannedEffectiveFrom time.Time

	FrozenContentHash *string

	hasRevisionVersion      bool
	hasScheduleGeneration   bool
	hasEffectiveFrom        bool
	hasPlannedEffectiveFrom bool
	stubSubmitSnapshots     bool
}

type Opt func(*Spec)

func WithTenant(id string) Opt            { return func(s *Spec) { s.TenantID = id } }
func WithUserID(id string) Opt            { return func(s *Spec) { s.UserID = id } }
func WithDisplayName(name string) Opt     { return func(s *Spec) { s.DisplayName = name } }
func WithRole(role string) Opt            { return func(s *Spec) { s.Role = role } }
func WithTaxonomy(tax Taxonomy) Opt       { return func(s *Spec) { s.Taxonomy = &tax } }
func WithControlledDoc(cd ControlledDoc) Opt { return func(s *Spec) { s.ControlledDoc = &cd } }
func WithDocument(d Document) Opt         { return func(s *Spec) { s.Document = &d } }
func WithRoute(r ApprovalRoute) Opt       { return func(s *Spec) { s.Route = &r } }
func WithOwner(userID string) Opt         { return func(s *Spec) { s.OwnerUserID = userID } }
func WithTemplateVersionID(id string) Opt { return func(s *Spec) { s.TemplateVersionID = id } }
func WithName(name string) Opt            { return func(s *Spec) { s.Name = name } }
func WithStatus(status string) Opt        { return func(s *Spec) { s.Status = status } }

// WithFrozenContentHash sets approval_instances.frozen_content_hash explicitly
// (F6 no-fallback hash chain: signoff/publish read ONLY this pin, never a
// COALESCE fallback). Pass "" to force NULL (the not-yet-frozen fixture case);
// any other value seeds that literal hash. When this option is omitted,
// NewApprovalInstance seeds a default frozen hash for any status at or beyond
// InstanceApproved, since a real approved instance always has a pin by then.
func WithFrozenContentHash(hash string) Opt {
	return func(s *Spec) { s.FrozenContentHash = &hash }
}
func WithCode(code string) Opt            { return func(s *Spec) { s.Code = code } }
func WithProfile(code string) Opt         { return func(s *Spec) { s.ProfileCode = code } }
func WithGovernanceClass(class string) Opt { return func(s *Spec) { s.GovernanceClass = class } }

// SetGovernanceClass reclassifies an existing document profile. Fixtures that
// seed route STAGES need a governed class (ADR 0087 / migration 0316 make a
// stageless active route DB-exclusive to livre, and conversely a livre route
// may carry no stages), but they often receive a profile that some other
// factory call already created with the default class. Call this BEFORE the
// route exists: the reclassification trigger validates every active route of
// the profile, and a profile with no routes reclassifies freely.
func SetGovernanceClass(t *testing.T, db *sql.DB, tenantID, profileCode, class string) {
	t.Helper()
	seedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`UPDATE metaldocs.document_profiles
			    SET governance_class = $3
			  WHERE tenant_id = $1::uuid AND code = $2`,
			tenantID, profileCode, class,
		)
		return err
	})
}

// WithStageSeeder runs fn inside the SAME transaction that inserts the
// approval_routes row, with the new route's id. Since ADR 0087 / migration
// 0316 a governed profile's ACTIVE route may never be stageless — not even
// between two committed statements — so "insert route, then insert stages"
// as separate autocommitted statements is no longer a legal fixture shape:
// the route insert alone would trip trg_route_profile_policy at ITS commit.
// Seeding both in one tx is the honest fixture: the DEFERRABLE INITIALLY
// DEFERRED trigger validates only the final committed shape.
func WithStageSeeder(fn func(tx *sql.Tx, routeID string) error) Opt {
	return func(s *Spec) { s.StageSeeder = fn }
}

// WithSubjectKind selects the approval-route subject kind ("document" or
// "template"). Both are profile-keyed since ADR 0086, so the factory always
// binds subject_key := profile_code.
func WithSubjectKind(kind string) Opt { return func(s *Spec) { s.SubjectKind = kind } }
func WithRevisionNumber(n int) Opt        { return func(s *Spec) { s.RevisionNumber = n } }
func WithRecipient(userID string) Opt     { return func(s *Spec) { s.RecipientUserID = userID } }
func WithEventType(t string) Opt          { return func(s *Spec) { s.EventType = t } }
func WithResourceType(t string) Opt       { return func(s *Spec) { s.ResourceType = t } }
func WithResourceID(id string) Opt        { return func(s *Spec) { s.ResourceID = id } }

func WithRevisionVersion(n int) Opt {
	return func(s *Spec) { s.RevisionVersion = n; s.hasRevisionVersion = true }
}
func WithScheduleGen(g int64) Opt {
	return func(s *Spec) { s.ScheduleGeneration = g; s.hasScheduleGeneration = true }
}
func WithEffectiveFrom(at time.Time) Opt {
	return func(s *Spec) { s.EffectiveFrom = at; s.hasEffectiveFrom = true }
}

// WithPlannedEffectiveFrom sets documents.planned_effective_from — the
// IMMUTABLE PLAN date (ADR 0085), which is a different column and a different
// concept from effective_from (the ACTUAL release instant, stamped by the
// release coordinator). A scheduled fixture states a plan; only a published one
// states an actual.
func WithPlannedEffectiveFrom(at time.Time) Opt {
	return func(s *Spec) { s.PlannedEffectiveFrom = at; s.hasPlannedEffectiveFrom = true }
}

// WithSubmitReadySnapshots forces the six enforce_snapshot_on_submit columns to
// be stubbed even for status='draft'. A real draft that is ready to submit for
// review has already had its template/content snapshot taken, so those columns
// are populated before the draft->under_review transition; the factory
// otherwise leaves them NULL for drafts. Opt into this when a test drives
// SubmitRevisionForReview (or any transition into under_review/approved/
// scheduled/published) off a factory-seeded draft.
func WithSubmitReadySnapshots() Opt {
	return func(s *Spec) { s.stubSubmitSnapshots = true }
}

func newSpec(opts []Opt) *Spec {
	s := &Spec{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// ---------------------------------------------------------------------------
// Builders
// ---------------------------------------------------------------------------

// NewTenant seeds a metaldocs.tenants row. Migration 0277 (M7 F7.2, ADR
// 0070) attached trg_require_cap_asserted to metaldocs.tenants (INSERT ->
// tenant.onboard) — so, unlike when this comment was first written, the
// INSERT now requires tenant.onboard asserted tx-locally (seedWithCaps,
// mirrors every other tripwire-guarded builder in this file).
func NewTenant(t *testing.T, db *sql.DB, opts ...Opt) Tenant {
	t.Helper()
	s := newSpec(opts)
	id := s.TenantID
	if id == "" {
		id = uuid.NewString()
	}
	seedWithCapsIdentity(t, db, id, "", `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, $2, $1::text)
			 ON CONFLICT (id) DO NOTHING`,
			id, "Test Tenant "+id,
		)
		return err
	})
	return Tenant{ID: id}
}

// NewUser seeds a metaldocs.iam_users row (user.manage tripwire on
// display_name/is_active/tenant_id/deactivated_at, migration 0259 — column-scoped
// so the login-context/presence telemetry columns stay ungated) and, when WithRole
// is given, the iam_user_roles row (user.manage tripwire, asserted tx-locally).
// Both writes share one seedWithCaps tx since they require the same capability.
func NewUser(t *testing.T, db *sql.DB, opts ...Opt) User {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	userID := s.UserID
	if userID == "" {
		userID = uuid.NewString()
	}
	display := s.DisplayName
	if display == "" {
		display = "Test User " + userID
	}

	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
			 VALUES ($1, $2, $3::uuid)
			 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
			userID, display, tenantID,
		)
		return err
	})

	if s.Role != "" {
		seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(context.Background(),
				`INSERT INTO metaldocs.iam_user_roles (user_id, role_code, tenant_id, assigned_by)
				 VALUES ($1, $2, $3::uuid, $1)
				 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code, assigned_by = EXCLUDED.assigned_by`,
				userID, s.Role, tenantID,
			)
			return err
		})
	}

	return User{ID: userID, DisplayName: display, TenantID: tenantID}
}

// NewTaxonomy seeds the governed family/process-area/profile chain with fresh,
// per-call-unique codes (each satisfying profile_code_format ^[a-z][a-z0-9_-]{1,63}$).
// All three tables carry the taxonomy.manage tripwire, asserted tx-locally.
func NewTaxonomy(t *testing.T, db *sql.DB, opts ...Opt) Taxonomy {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	suffix := randomSuffix(t)
	family := "fam-" + suffix
	processArea := "pa-" + suffix
	profile := "prof-" + suffix

	// governance_class: fixtures default to 'livre' because that is the class
	// whose route shape the factory actually seeds — NewApprovalRoute creates a
	// bare ACTIVE route with NO stages, and since ADR 0087 / migration 0316 a
	// stageless active route is DB-exclusive to livre (every governed class now
	// requires >=1 stage at the DB line, because a stageless route
	// auto-approves). The pre-0087 default was 'simples', which was valid only
	// while simples carried no route-shape constraint at all.
	//
	// Tests that seed route STAGES must therefore opt into a governed class
	// explicitly (WithGovernanceClass("controlado"/"simples")). Stage
	// INSTANCES (approval_stage_instances) are unaffected — the route-shape
	// trigger reads approval_route_stages only.
	class := s.GovernanceClass
	if class == "" {
		class = "livre"
	}

	seedWithCapsIdentity(t, db, tenantID, "", `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		ctx := context.Background()
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_families (code, name)
			 VALUES ($1, $1) ON CONFLICT (code) DO NOTHING`, family,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_process_areas (code, tenant_id, name)
			 VALUES ($1, $2::uuid, $1) ON CONFLICT (tenant_id, code) DO NOTHING`, processArea, tenantID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.document_profiles (code, tenant_id, family_code, name, review_interval_days, alias, governance_class)
			 VALUES ($1, $2::uuid, $3, $1, 365, $1, $4) ON CONFLICT (tenant_id, code) DO NOTHING`,
			profile, tenantID, family, class,
		); err != nil {
			return err
		}
		return nil
	})

	return Taxonomy{TenantID: tenantID, FamilyCode: family, ProcessAreaCode: processArea, ProfileCode: profile}
}

// NewControlledDoc seeds a public.controlled_documents row (controlled_documents.create
// tripwire), auto-wiring a tenant, taxonomy, and owner when not supplied.
func NewControlledDoc(t *testing.T, db *sql.DB, opts ...Opt) ControlledDoc {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	var tax Taxonomy
	if s.Taxonomy != nil {
		tax = *s.Taxonomy
		if tenantID == "" {
			tenantID = tax.TenantID
		}
	}
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	if s.Taxonomy == nil {
		tax = NewTaxonomy(t, db, WithTenant(tenantID), WithGovernanceClass(s.GovernanceClass))
	}

	owner := s.OwnerUserID
	if owner == "" {
		owner = NewUser(t, db, WithTenant(tenantID)).ID
	}

	id := uuid.NewString()
	code := s.Code
	if code == "" {
		code = "cd-" + randomSuffix(t)
	}
	status := s.Status
	if status == "" {
		status = "active"
	}
	title := s.Name
	if title == "" {
		title = "Test Controlled Doc"
	}

	seedWithCapsIdentity(t, db, tenantID, owner, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.controlled_documents
			   (id, tenant_id, profile_code, process_area_code, code, title, owner_user_id, status)
			 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)`,
			id, tenantID, tax.ProfileCode, tax.ProcessAreaCode, code, title, owner, status,
		)
		return err
	})

	return ControlledDoc{
		ID: id, TenantID: tenantID, Code: code,
		ProfileCode: tax.ProfileCode, ProcessAreaCode: tax.ProcessAreaCode,
		OwnerUserID: owner, Status: status,
	}
}

// NewDocument seeds a public.documents row (document.create tripwire) on the
// CD-governed lineage, auto-wiring a controlled document when not supplied. The
// template_version_id is a free minted UUID by default (the CD-governed path
// carries no template-version FK); WithTemplateVersionID overrides. Optional
// revision_version / schedule_generation / effective_from are omitted from the
// INSERT when unset, so the DB default applies (mirroring the consumers).
func NewDocument(t *testing.T, db *sql.DB, opts ...Opt) Document {
	t.Helper()
	s := newSpec(opts)

	var cd ControlledDoc
	if s.ControlledDoc != nil {
		cd = *s.ControlledDoc
	} else {
		cdOpts := []Opt{}
		if s.TenantID != "" {
			cdOpts = append(cdOpts, WithTenant(s.TenantID))
		}
		if s.OwnerUserID != "" {
			cdOpts = append(cdOpts, WithOwner(s.OwnerUserID))
		}
		cd = NewControlledDoc(t, db, cdOpts...)
	}

	owner := s.OwnerUserID
	if owner == "" {
		owner = cd.OwnerUserID
	}
	docID := uuid.NewString()
	tvID := s.TemplateVersionID
	if tvID == "" {
		tvID = uuid.NewString()
	}
	name := s.Name
	if name == "" {
		name = "Test Doc"
	}
	status := s.Status
	if status == "" {
		status = "draft"
	}

	cols := []string{"id", "tenant_id", "template_version_id", "name", "status",
		"form_data_json", "created_by", "controlled_document_id", "revision_number"}
	casts := []string{"::uuid", "::uuid", "::uuid", "", "", "::jsonb", "", "::uuid", ""}
	args := []any{docID, cd.TenantID, tvID, name, status, "{}", owner, cd.ID, s.RevisionNumber}

	if s.hasRevisionVersion {
		cols = append(cols, "revision_version")
		casts = append(casts, "")
		args = append(args, s.RevisionVersion)
	}
	if s.hasScheduleGeneration {
		cols = append(cols, "schedule_generation")
		casts = append(casts, "")
		args = append(args, s.ScheduleGeneration)
	}
	if s.hasEffectiveFrom {
		cols = append(cols, "effective_from")
		casts = append(casts, "")
		args = append(args, s.EffectiveFrom)
	}
	if s.hasPlannedEffectiveFrom {
		cols = append(cols, "planned_effective_from")
		casts = append(casts, "")
		args = append(args, s.PlannedEffectiveFrom)
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d%s", i+1, casts[i])
	}

	// ADR 0085 splits the effective date in two: planned_effective_from is the
	// immutable PLAN, effective_from the ACTUAL release instant. Migration 0310
	// makes that a DB invariant (ck_documents_published_effective_from,
	// ck_documents_scheduled_planned_effective_from), so the factory must seed
	// a coherent row for the two terminal-ish statuses exactly as it already
	// stubs the enforce_snapshot_on_submit columns below — a fixture that seeds
	// `published` is asserting the document IS effective, and one that seeds
	// `scheduled` is asserting a plan exists.
	switch {
	case status == "published" && !s.hasEffectiveFrom:
		cols = append(cols, "effective_from")
		placeholders = append(placeholders, "now()")
	case status == "scheduled":
		// A scheduled document has NOT been released, so it must carry a PLAN
		// and no ACTUAL date. Stating an actual on a scheduled fixture is a
		// contradiction, not a shorthand for the plan — say
		// WithPlannedEffectiveFrom.
		if s.hasEffectiveFrom {
			t.Fatalf("NewDocument: WithEffectiveFrom is meaningless for status=scheduled (ADR 0085: the plan lives in planned_effective_from; effective_from is stamped at actual release). Use WithPlannedEffectiveFrom.")
		}
		if !s.hasPlannedEffectiveFrom {
			cols = append(cols, "planned_effective_from")
			placeholders = append(placeholders, "now() + interval '7 days'")
		}
	}

	// enforce_snapshot_on_submit (23514) requires the six snapshot columns for
	// any non-draft status. Stub them with literal expressions (32-byte hashes
	// satisfy the *_hash_len checks), mirroring fixtures.go's supersede walk.
	// A draft that a test will drive into under_review needs them too — opt in
	// via WithSubmitReadySnapshots (a real submit-ready draft has them set).
	if status != "draft" || s.stubSubmitSnapshots {
		cols = append(cols,
			"placeholder_schema_snapshot", "placeholder_schema_hash",
			"composition_config_snapshot", "composition_config_hash",
			"body_docx_snapshot_s3_key", "body_docx_hash")
		placeholders = append(placeholders,
			`'{}'::jsonb`, `decode(repeat('ab',32),'hex')`,
			`'{}'::jsonb`, `decode(repeat('cd',32),'hex')`,
			`'tenants/test/`+docID+`/body.docx'`, `decode(repeat('ef',32),'hex')`)
	}

	query := `INSERT INTO public.documents (` + strings.Join(cols, ", ") + `)
		 VALUES (` + strings.Join(placeholders, ", ") + `)`

	seedWithCapsIdentity(t, db, cd.TenantID, owner, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), query, args...)
		return err
	})

	return Document{
		ID: docID, TenantID: cd.TenantID, ControlledDocumentID: cd.ID,
		TemplateVersionID: tvID, Owner: owner, Status: status,
		RevisionNumber: s.RevisionNumber, RevisionVersion: s.RevisionVersion,
		ScheduleGeneration: s.ScheduleGeneration,
	}
}

// NewApprovalRoute seeds a public.approval_routes row (no tripwire), auto-wiring
// a tenant and a profile when not supplied.
func NewApprovalRoute(t *testing.T, db *sql.DB, opts ...Opt) ApprovalRoute {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	profile := s.ProfileCode
	if profile == "" {
		profile = NewTaxonomy(t, db, WithTenant(tenantID), WithGovernanceClass(s.GovernanceClass)).ProfileCode
	}
	owner := s.OwnerUserID
	if owner == "" {
		owner = NewUser(t, db, WithTenant(tenantID)).ID
	}
	name := s.Name
	if name == "" {
		name = "Route"
	}
	subjectKind := s.SubjectKind
	if subjectKind == "" {
		subjectKind = "document"
	}
	id := uuid.NewString()
	seedTenantOnly(t, db, tenantID, owner, func(tx *sql.Tx) error {
		// subject_key is bound explicitly rather than left to the alias trigger:
		// since ADR 0086 BOTH kinds are profile-keyed, and the template arm is
		// constrained by approval_routes_template_subject_key_check.
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.approval_routes
			   (id, tenant_id, name, profile_code, subject_kind, subject_key, active, created_by)
			 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $4, true, $6)`,
			id, tenantID, name, profile, subjectKind, owner,
		)
		if err != nil {
			return err
		}
		if s.StageSeeder != nil {
			// Same tx as the route row (ADR 0087): a governed profile's active
			// route must never commit stageless.
			return s.StageSeeder(tx, id)
		}
		return nil
	})
	return ApprovalRoute{ID: id, TenantID: tenantID, ProfileCode: profile, SubjectKind: subjectKind}
}

// NewApprovalInstance seeds a public.approval_instances row (document.submit
// tripwire), auto-wiring the document and route when not supplied. The
// submitter defaults to the document's owner; content hash and idempotency key
// are minted to satisfy the length/uniqueness constraints.
func NewApprovalInstance(t *testing.T, db *sql.DB, opts ...Opt) ApprovalInstance {
	t.Helper()
	s := newSpec(opts)

	var doc Document
	if s.Document != nil {
		doc = *s.Document
	} else {
		docOpts := []Opt{WithStatus("approved")}
		if s.TenantID != "" {
			// Thread the caller's WithTenant through to the auto-created
			// document — without this, a caller passing WithTenant(tenantID)
			// alone (no WithDocument) silently got a document under a
			// DIFFERENT, freshly-minted tenant, and every downstream query
			// filtered on the caller's tenantID (e.g. approval_instances.tenant_id)
			// found nothing.
			docOpts = append(docOpts, WithTenant(s.TenantID))
		}
		doc = NewDocument(t, db, docOpts...)
	}
	var route ApprovalRoute
	if s.Route != nil {
		route = *s.Route
	} else {
		route = NewApprovalRoute(t, db, WithTenant(doc.TenantID), WithProfile(profileForCD(t, db, doc.ControlledDocumentID)))
	}

	submitter := s.OwnerUserID
	if submitter == "" {
		submitter = doc.Owner
	}
	status := s.Status
	if status == "" {
		status = "approved"
	}
	id := uuid.NewString()
	idemKey := "idem-" + randomSuffix(t)

	// frozen_content_hash (F6 no-fallback hash chain): signoff/publish read
	// ONLY this pin, never a fallback. A real "approved" instance always has
	// one set (F1/F5 freeze boundary), so default-seed it here unless the
	// caller explicitly overrides via WithFrozenContentHash (including
	// forcing NULL with WithFrozenContentHash("") for not-yet-frozen fixtures).
	var frozenHash sql.NullString
	switch {
	case s.FrozenContentHash != nil && *s.FrozenContentHash != "":
		frozenHash = sql.NullString{String: *s.FrozenContentHash, Valid: true}
	case s.FrozenContentHash != nil:
		frozenHash = sql.NullString{Valid: false}
	case status == string(domain.InstanceApproved):
		frozenHash = sql.NullString{String: strings.Repeat("a", 64), Valid: true}
	}

	seedWithCapsIdentity(t, db, doc.TenantID, submitter, `[{"cap":"document.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.approval_instances
			   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
			    submitted_by, submitted_at, content_hash_at_submit, idempotency_key,
			    frozen_content_hash)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, $5, $6, now(), repeat('a', 64), $7, $8)`,
			id, doc.TenantID, doc.ID, route.ID, status, submitter, idemKey, frozenHash,
		)
		return err
	})

	var frozenHashOut *string
	if frozenHash.Valid {
		v := frozenHash.String
		frozenHashOut = &v
	}
	return ApprovalInstance{ID: id, TenantID: doc.TenantID, DocumentID: doc.ID, RouteID: route.ID, Status: status, FrozenContentHash: frozenHashOut}
}

// NewNotification seeds a metaldocs.notifications row (no tripwire). Auto-wires a
// tenant and a recipient user when not supplied. status defaults to 'PENDING'.
// source_event_id is left NULL (the partial unique index does not constrain NULLs),
// so multiple fixture rows for one recipient never collide.
func NewNotification(t *testing.T, db *sql.DB, opts ...Opt) Notification {
	t.Helper()
	s := newSpec(opts)

	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = NewTenant(t, db).ID
	}
	recipient := s.RecipientUserID
	if recipient == "" {
		recipient = NewUser(t, db, WithTenant(tenantID)).ID
	}
	id := uuid.NewString()
	eventType := s.EventType
	if eventType == "" {
		eventType = "document_published"
	}
	resourceType := s.ResourceType
	if resourceType == "" {
		resourceType = "document"
	}
	resourceID := s.ResourceID
	if resourceID == "" {
		resourceID = uuid.NewString()
	}
	status := s.Status
	if status == "" {
		status = "PENDING"
	}

	seedTenantOnly(t, db, tenantID, "", func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO metaldocs.notifications
			   (id, tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, status)
			 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)`,
			id, tenantID, recipient, eventType, resourceType, resourceID,
			"Novo documento controlado para leitura", "Um documento foi publicado.", status,
		)
		return err
	})

	return Notification{
		ID: id, TenantID: tenantID, RecipientUserID: recipient,
		EventType: eventType, ResourceType: resourceType, ResourceID: resourceID, Status: status,
	}
}

// NewCDFieldReader returns the real Postgres-backed controlleddocuments
// CDFieldReader (the published cross-module read port, ADR-0039 D3(b)) for use by
// integration tests that exercise the documents/approval area-resolver COALESCE
// chains (B5, B6 parity gates). Centralizing the cross-module construction here —
// the testdb fixture-factory home — keeps the boundary crossing in one place
// instead of having each consumer test import controlleddocuments/infrastructure
// directly. The concrete reader is stateless, so the *testing.T is taken only to
// keep the accessor uniform with the other builders and future-proof for setup.
func NewCDFieldReader(t *testing.T) cddomain.CDFieldReader {
	t.Helper()
	return cdinfra.NewCDFieldReaderPG()
}

func profileForCD(t *testing.T, db *sql.DB, controlledDocID string) string {
	t.Helper()
	var profile string
	if err := db.QueryRowContext(context.Background(),
		`SELECT profile_code FROM public.controlled_documents WHERE id=$1::uuid`, controlledDocID,
	).Scan(&profile); err != nil {
		t.Fatalf("profileForCD: %v", err)
	}
	return profile
}

// ---------------------------------------------------------------------------
// Scenario composites — wire the common multi-row shapes the consumers build by
// hand (a published head, a scheduled revision at a given generation).
// ---------------------------------------------------------------------------

type Scenario struct{}

// PublishedDocument builds tenant → user → taxonomy → CD → one published
// document. Caller options thread through to NewDocument (and its inner
// builders); the published status is applied last so it always wins.
func (Scenario) PublishedDocument(t *testing.T, db *sql.DB, opts ...Opt) Document {
	t.Helper()
	return NewDocument(t, db, append(opts, WithStatus("published"))...)
}

// ScheduledRevision builds the scheduled-publish-worker shape: one scheduled
// document at schedule_generation = gen.
func (Scenario) ScheduledRevision(t *testing.T, db *sql.DB, gen int64, opts ...Opt) Document {
	t.Helper()
	return NewDocument(t, db, append(opts, WithStatus("scheduled"), WithScheduleGen(gen))...)
}
