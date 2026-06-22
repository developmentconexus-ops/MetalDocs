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
}

type ApprovalInstance struct {
	ID         string
	TenantID   string
	DocumentID string
	RouteID    string
	Status     string
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

	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string

	RevisionNumber     int
	RevisionVersion    int
	ScheduleGeneration int64
	EffectiveFrom      time.Time

	hasRevisionVersion    bool
	hasScheduleGeneration bool
	hasEffectiveFrom      bool
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
func WithCode(code string) Opt            { return func(s *Spec) { s.Code = code } }
func WithProfile(code string) Opt         { return func(s *Spec) { s.ProfileCode = code } }
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

func newSpec(opts []Opt) *Spec {
	s := &Spec{}
	for _, o := range opts {
		o(s)
	}
	return s
}

func exec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("factory exec failed: %v\nquery: %s", err, query)
	}
}

// ---------------------------------------------------------------------------
// Builders
// ---------------------------------------------------------------------------

// NewTenant seeds a metaldocs.tenants row. tenants carries no tripwire.
func NewTenant(t *testing.T, db *sql.DB, opts ...Opt) Tenant {
	t.Helper()
	s := newSpec(opts)
	id := s.TenantID
	if id == "" {
		id = uuid.NewString()
	}
	exec(t, db,
		`INSERT INTO metaldocs.tenants (id, name, slug)
		 VALUES ($1::uuid, $2, $1::text)
		 ON CONFLICT (id) DO NOTHING`,
		id, "Test Tenant "+id,
	)
	return Tenant{ID: id}
}

// NewUser seeds a metaldocs.iam_users row (no tripwire) and, when WithRole is
// given, the iam_user_roles row (user.manage tripwire, asserted tx-locally).
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

	exec(t, db,
		`INSERT INTO metaldocs.iam_users (user_id, display_name, tenant_id)
		 VALUES ($1, $2, $3::uuid)
		 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
		userID, display, tenantID,
	)

	if s.Role != "" {
		seedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
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

	seedWithCaps(t, db, `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
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
			`INSERT INTO metaldocs.document_profiles (code, tenant_id, family_code, name, review_interval_days, alias)
			 VALUES ($1, $2::uuid, $3, $1, 365, $1) ON CONFLICT (tenant_id, code) DO NOTHING`,
			profile, tenantID, family,
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
		tax = NewTaxonomy(t, db, WithTenant(tenantID))
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

	seedWithCaps(t, db, `[{"cap":"controlled_documents.create"}]`, func(tx *sql.Tx) error {
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

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d%s", i+1, casts[i])
	}

	// enforce_snapshot_on_submit (23514) requires the six snapshot columns for
	// any non-draft status. Stub them with literal expressions (32-byte hashes
	// satisfy the *_hash_len checks), mirroring fixtures.go's supersede walk.
	if status != "draft" {
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

	seedWithCaps(t, db, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
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
		profile = NewTaxonomy(t, db, WithTenant(tenantID)).ProfileCode
	}
	owner := s.OwnerUserID
	if owner == "" {
		owner = NewUser(t, db, WithTenant(tenantID)).ID
	}
	name := s.Name
	if name == "" {
		name = "Route"
	}
	id := uuid.NewString()
	exec(t, db,
		`INSERT INTO public.approval_routes (id, tenant_id, name, profile_code, active, created_by)
		 VALUES ($1::uuid, $2::uuid, $3, $4, true, $5)`,
		id, tenantID, name, profile, owner,
	)
	return ApprovalRoute{ID: id, TenantID: tenantID, ProfileCode: profile}
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
		doc = NewDocument(t, db, WithStatus("approved"))
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

	seedWithCaps(t, db, `[{"cap":"document.submit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(),
			`INSERT INTO public.approval_instances
			   (id, tenant_id, document_id, route_id, route_version_snapshot, status,
			    submitted_by, submitted_at, content_hash_at_submit, idempotency_key)
			 VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 1, $5, $6, now(), repeat('a', 64), $7)`,
			id, doc.TenantID, doc.ID, route.ID, status, submitter, idemKey,
		)
		return err
	})

	return ApprovalInstance{ID: id, TenantID: doc.TenantID, DocumentID: doc.ID, RouteID: route.ID, Status: status}
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

	exec(t, db,
		`INSERT INTO metaldocs.notifications
		   (id, tenant_id, recipient_user_id, event_type, resource_type, resource_id, title, message, status)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)`,
		id, tenantID, recipient, eventType, resourceType, resourceID,
		"Novo documento controlado para leitura", "Um documento foi publicado.", status,
	)

	return Notification{
		ID: id, TenantID: tenantID, RecipientUserID: recipient,
		EventType: eventType, ResourceType: resourceType, ResourceID: resourceID, Status: status,
	}
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
