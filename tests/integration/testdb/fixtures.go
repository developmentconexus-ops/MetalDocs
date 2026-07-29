//go:build integration
// +build integration

package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"metaldocs/internal/modules/iam/authz"
)

// InsertDraftDocument inserts a minimal draft document with an initial revision
// into the test schema. Returns (docID, tenantID).
// tenantID must be a valid UUID string.
//
// Guarded writes (INSERT/UPDATE on public.documents) are wrapped in seedWithCaps
// transaction-locally — pool-safe, never leaks session state.
//
// DB-01: seeds the CANONICAL templates_template / templates_template_version
// family (mirrors TST-01's internal/test/e2e_seed.go ensureTemplateVersion,
// commit 1ac5e530). Unlike the legacy public.templates / template_versions pair,
// both canonical tables carry trg_require_cap_asserted (satisfied by the
// template.create capability, asserted tx-locally via seedWithCaps) and
// templates_template_version additionally carries
// trg_template_version_tenant_consistent, which requires the tx-local
// metaldocs.tenant_id GUC to be set and match the parent template's tenant —
// set once per seed tx alongside the capability assertion.
func InsertDraftDocument(t *testing.T, db *sql.DB, schema, tenantID string) (docID, tenant string) {
	t.Helper()
	ctx := context.Background()

	userID := DeterministicID(t, "user")
	templateKey := "test-template-" + randomSuffix(t)

	// Insert minimal published template + version into the canonical family.
	// content_hash must be 64-hex for any non-draft status
	// (chk_template_version_content_hash_non_draft); docx_storage_key is
	// globally unique (uq_templates_template_version_docx_storage_key), so key
	// it off the minted template ID to stay collision-free across parallel tests.
	var tplID, tvID string
	docxStorageKey := "templates/test-seed/" + templateKey + "/body.docx"
	contentHash := strings.Repeat("0", 64)

	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"template.create"}]`, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO `+Qualified(schema, "templates_template")+
				` (id, tenant_id, doc_type_code, key, name, description, latest_version, published_version_id, created_by, created_at)
			 VALUES (gen_random_uuid(), $1::uuid, '', $2, 'Test Template', '', 1, NULL, $3::text, now())
			 RETURNING id::text`,
			tenantID, templateKey, userID,
		).Scan(&tplID); err != nil {
			return fmt.Errorf("insert templates_template: %w", err)
		}

		if err := tx.QueryRowContext(ctx,
			`INSERT INTO `+Qualified(schema, "templates_template_version")+
				` (id, tenant_id, template_id, version_number, status, docx_storage_key, content_hash,
				   metadata_schema, placeholder_schema, author_id, published_at)
			 VALUES (gen_random_uuid(), $1::uuid, $2::uuid, 1, 'published', $3, $4,
				   '{}'::jsonb, '{"placeholders":[]}'::jsonb, $5::text, now())
			 RETURNING id::text`,
			tenantID, tplID, docxStorageKey, contentHash, userID,
		).Scan(&tvID); err != nil {
			return fmt.Errorf("insert templates_template_version: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE `+Qualified(schema, "templates_template")+
				` SET published_version_id = $2::uuid WHERE id = $1::uuid`,
			tplID, tvID,
		); err != nil {
			return fmt.Errorf("update templates_template.published_version_id: %w", err)
		}
		return nil
	})

	// Insert document — guarded by document.create tripwire.
	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx,
			`INSERT INTO `+Qualified(schema, "documents")+
				` (tenant_id, template_version_id, name, status, form_data_json, created_by)
			 VALUES ($1::uuid, $2::uuid, 'Test Doc', 'draft', '{}', $3::uuid)
			 RETURNING id::text`,
			tenantID, tvID, userID,
		).Scan(&docID)
	})

	// Insert editor session (no tripwire on public.editor_sessions).
	var sessID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "editor_sessions")+
			` (tenant_id, document_id, user_id, expires_at, last_acknowledged_revision_id, status)
		 VALUES ($1::uuid, $2::uuid, $3::uuid, now() + interval '1 hour',
		         '00000000-0000-0000-0000-000000000000', 'active')
		 RETURNING id::text`,
		tenantID, docID, userID,
	).Scan(&sessID); err != nil {
		t.Fatalf("InsertDraftDocument: insert session: %v", err)
	}

	// Insert initial revision (no tripwire on public.document_revisions).
	var revID string
	if err := db.QueryRowContext(ctx,
		`INSERT INTO `+Qualified(schema, "document_revisions")+
			` (document_id, parent_revision_id, session_id, storage_key, content_hash, form_data_snapshot)
		 VALUES ($1::uuid, NULL, $2::uuid, '', 'aabbcc', '{}')
		 RETURNING id::text`,
		docID, sessID,
	).Scan(&revID); err != nil {
		t.Fatalf("InsertDraftDocument: insert revision: %v", err)
	}

	// Update document pointers — guarded by document.edit tripwire.
	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`UPDATE `+Qualified(schema, "documents")+
				` SET current_revision_id=$1::uuid, active_session_id=$2::uuid, updated_at=now()
			 WHERE id=$3::uuid`,
			revID, sessID, docID,
		)
		return err
	})

	// Update session ack (no tripwire on public.editor_sessions).
	if _, err := db.ExecContext(ctx,
		`UPDATE `+Qualified(schema, "editor_sessions")+
			` SET last_acknowledged_revision_id=$1::uuid WHERE id=$2::uuid AND tenant_id=$3::uuid`,
		revID, sessID, tenantID,
	); err != nil {
		t.Fatalf("InsertDraftDocument: update session ack: %v", err)
	}

	return docID, tenantID
}

// SeedWithCaps is the exported wrapper over seedWithCaps, for consumer tests
// that must run a raw guarded write (INSERT/UPDATE on a tripwire table) as their
// system-under-test without an inline set_config. The capability is asserted
// transaction-locally (pool-safe), exactly as the production authz layer does.
func SeedWithCaps(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	seedWithCaps(t, db, capsJSON, fn)
}

// SetCapsOnTx asserts capabilities transaction-locally on an already-open tx.
// Use this when the test owns the tx (e.g., passes it to a repo method that
// requires a caller-managed transaction) and must satisfy the tripwire without
// inline set_config in the test file. Mirrors what the production authz layer
// does: set_config with is_local=true so the assertion is discarded on commit.
func SetCapsOnTx(t *testing.T, tx *sql.Tx, capsJSON string) {
	t.Helper()
	if _, err := tx.ExecContext(context.Background(),
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, capsJSON,
	); err != nil {
		t.Fatalf("SetCapsOnTx: %v", err)
	}
}

// SetCapsOnDB asserts capabilities at session level on the given *sql.DB. Safe
// only for isolated per-test databases (MaxOpenConns=1, DB dropped after test).
// Use when the system-under-test takes a *sql.DB directly and cannot be wrapped
// in a SeedWithCaps transaction (e.g., repo methods with no DBTX variadic).
func SetCapsOnDB(t *testing.T, db *sql.DB, capsJSON string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`SELECT set_config('metaldocs.asserted_caps', $1, false)`, capsJSON,
	); err != nil {
		t.Fatalf("SetCapsOnDB: %v", err)
	}
}

// seedWithCaps runs fn inside its own transaction with the given capabilities
// asserted transaction-locally (set_config third arg true), exactly as the
// production authz layer asserts them (authz.appendAssertedCap). Doing the
// set_config and the governed write in one transaction makes the helper safe
// against connection pools of any size: the tripwire reads the caps on the same
// connection that performs the write, and the assertion is discarded on commit
// so it never leaks into the caller's session. capsJSON is the JSONB array
// literal, e.g. `[{"cap":"taxonomy.manage"}]`.
func seedWithCaps(t *testing.T, db *sql.DB, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedWithCaps: begin tx: %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`SELECT set_config('metaldocs.asserted_caps', $1, true)`, capsJSON,
	); err != nil {
		t.Fatalf("seedWithCaps: assert caps %s: %v", capsJSON, err)
	}
	if err := fn(tx); err != nil {
		t.Fatalf("seedWithCaps: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedWithCaps: commit: %v", err)
	}
}

// seedTx is the identity-aware sibling of seedWithCaps: same tx-local,
// pool-safe shape, but ALSO seeds the metaldocs.tenant_id (and, when actorID
// is non-empty, metaldocs.actor_id) GUC on the same tx before fn runs — via
// the production authz.SeedTxIdentity / authz.SeedTxTenant functions, never
// hand-written set_config. This satisfies both the tenant-consistency
// triggers (e.g. trg_template_version_tenant_consistent) and any authz.Require
// call fn makes on its way to a guarded write. capsJSON == "" skips the
// asserted_caps assertion entirely, for builders that write tables with no
// trg_require_cap_asserted tripwire.
func seedTx(t *testing.T, db *sql.DB, tenantID, actorID, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seedTx: begin tx: %v", err)
	}
	defer tx.Rollback()

	if actorID != "" {
		if err := authz.SeedTxIdentity(ctx, tx, tenantID, actorID); err != nil {
			t.Fatalf("seedTx: seed identity: %v", err)
		}
	} else {
		if err := authz.SeedTxTenant(ctx, tx, tenantID); err != nil {
			t.Fatalf("seedTx: seed tenant: %v", err)
		}
	}

	if capsJSON != "" {
		if _, err := tx.ExecContext(ctx,
			`SELECT set_config('metaldocs.asserted_caps', $1, true)`, capsJSON,
		); err != nil {
			t.Fatalf("seedTx: assert caps %s: %v", capsJSON, err)
		}
	}
	if err := fn(tx); err != nil {
		t.Fatalf("seedTx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("seedTx: commit: %v", err)
	}
}

// seedWithCapsIdentity is seedWithCaps + tenant/actor GUC seeding. Use for any
// guarded (trg_require_cap_asserted) write whose builder Spec knows the
// tenant, and the actor when one exists.
func seedWithCapsIdentity(t *testing.T, db *sql.DB, tenantID, actorID, capsJSON string, fn func(tx *sql.Tx) error) {
	t.Helper()
	seedTx(t, db, tenantID, actorID, capsJSON, fn)
}

// seedTenantOnly seeds tenant/actor identity on a tx-local seed transaction
// for writes to tables with NO trg_require_cap_asserted tripwire (so there is
// no capability to assert), but that still need metaldocs.tenant_id (and
// optionally metaldocs.actor_id) set — either for a tenant-consistency
// trigger or for an authz.Require call inside fn.
func seedTenantOnly(t *testing.T, db *sql.DB, tenantID, actorID string, fn func(tx *sql.Tx) error) {
	t.Helper()
	seedTx(t, db, tenantID, actorID, "", fn)
}

// SeedGovernedTaxonomy seeds the governed FK-parent chain required before any
// controlled_documents insert: a shared document_families row, plus the
// (tenantID, processAreaCode) document_process_areas row and the
// (tenantID, profileCode) document_profiles row the controlled document
// references. All three tables carry the metaldocs authz tripwire
// (trg_require_cap_asserted -> taxonomy.manage), so the writes run inside a
// transaction that asserts taxonomy.manage transaction-locally (pool-safe; the
// assertion never leaks into the caller's session).
//
// Idempotent (ON CONFLICT DO NOTHING) so repeated calls within a test are safe.
func SeedGovernedTaxonomy(t *testing.T, db *sql.DB, tenantID, profileCode, processAreaCode string) {
	t.Helper()
	ctx := context.Background()

	const familyCode = "test-family"

	seedWithCapsIdentity(t, db, tenantID, "", `[{"cap":"taxonomy.manage"}]`, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_families")+` (code, name)
			 VALUES ($1, $1)
			 ON CONFLICT (code) DO NOTHING`,
			familyCode,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_process_areas")+` (code, tenant_id, name)
			 VALUES ($1, $2::uuid, $1)
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			processAreaCode, tenantID,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "document_profiles")+`
			    (code, tenant_id, family_code, name, review_interval_days, alias, governance_class)
			 VALUES ($1, $2::uuid, $3, $1, 365, $1, 'simples')
			 ON CONFLICT (tenant_id, code) DO NOTHING`,
			profileCode, tenantID, familyCode,
		); err != nil {
			return err
		}
		return nil
	})
}

// SupersedeActiveDocumentForCD walks the single active document of the given
// controlled document through the legal status lifecycle to 'superseded', so a
// successor revision can be created. ux_documents_cd_active permits only one
// active document (draft/under_review/approved/rejected/scheduled) per CD, and
// trg_documents_legal_transition allows no shortcut out of 'draft' — the only
// exit is the full path draft -> under_review -> approved -> published ->
// superseded. This mirrors production: a prior revision is published and then
// superseded before its successor exists. enforce_snapshot_on_submit requires
// the six snapshot columns for under_review/approved/published, so they are
// stubbed (32-byte hashes satisfy the *_hash_len checks).
//
// Each UPDATE carries the documents-UPDATE tripwire (document.edit), asserted
// transaction-locally so the helper is pool-safe.
func SupersedeActiveDocumentForCD(t *testing.T, db *sql.DB, controlledDocumentID string) {
	t.Helper()
	ctx := context.Background()

	var tenantID string
	if err := db.QueryRowContext(ctx,
		`SELECT tenant_id::text FROM `+Qualified("", "controlled_documents")+` WHERE id = $1::uuid`,
		controlledDocumentID,
	).Scan(&tenantID); err != nil {
		t.Fatalf("SupersedeActiveDocumentForCD: resolve tenant: %v", err)
	}

	seedWithCapsIdentity(t, db, tenantID, "", `[{"cap":"document.edit"}]`, func(tx *sql.Tx) error {
		var docID string
		if err := tx.QueryRowContext(ctx,
			`SELECT id::text FROM `+Qualified("", "documents")+`
			  WHERE controlled_document_id = $1::uuid
			    AND status = ANY (ARRAY['draft','under_review','approved','rejected','scheduled'])`,
			controlledDocumentID,
		).Scan(&docID); err != nil {
			return err
		}

		// Stub the snapshot columns (still 'draft', so no transition / no snapshot
		// gate yet), then walk the legal transition graph to 'superseded'.
		if _, err := tx.ExecContext(ctx,
			`UPDATE `+Qualified("", "documents")+` SET
			    placeholder_schema_snapshot = '{}'::jsonb,
			    placeholder_schema_hash     = decode(repeat('ab',32),'hex'),
			    composition_config_snapshot = '{}'::jsonb,
			    composition_config_hash     = decode(repeat('cd',32),'hex'),
			    body_docx_snapshot_s3_key   = 'tenants/test/superseded/body.docx',
			    body_docx_hash              = decode(repeat('ef',32),'hex')
			  WHERE id = $1::uuid`,
			docID,
		); err != nil {
			return err
		}
		// ADR 0085: a published document without effective_from is exactly
		// F-QA4-13 (it escapes the periodic-review surfacer) and is now rejected
		// by ck_documents_published_effective_from. The walk therefore stamps the
		// ACTUAL release instant when it enters 'published' — and leaves it in
		// place through 'superseded', where it remains the historical record of
		// when the document was in force.
		for _, status := range []string{"under_review", "approved", "published", "superseded"} {
			set := `status = $2`
			if status == "published" {
				set += `, effective_from = COALESCE(effective_from, now())`
			}
			if _, err := tx.ExecContext(ctx,
				`UPDATE `+Qualified("", "documents")+` SET `+set+` WHERE id = $1::uuid`,
				docID, status,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

// SeedSystemAdmin seeds the actor as a tenant-scoped system_admin so the
// application-level authz check inside the create path (authz.Require, ADR 0022
// tier-2) is satisfied via the system_admin inheritance bypass — without
// granting per-area memberships. It upserts the iam_users row (display_name is
// preserved for the iam UserDisplayNameReader port) and the iam_user_roles row.
// SEC-05 / migration 0259: iam_users now carries trg_require_cap_asserted
// (user.manage) on INSERT and on UPDATE of privileged columns (display_name,
// tenant_id among them), same as iam_user_roles — both writes run inside a
// user.manage transaction (pool-safe, tx-local).
func SeedSystemAdmin(t *testing.T, db *sql.DB, tenantID, userID, displayName string) {
	t.Helper()
	ctx := context.Background()

	// iam_users.tenant_id FK -> metaldocs.tenants; seed the tenant parent first.
	// slug = tenantID guarantees uniqueness/non-blank. ON CONFLICT (id) DO
	// NOTHING leaves an already-seeded reference tenant (e.g. the dev tenant)
	// untouched. Migration 0277 (M7 F7.2, ADR 0070) attached
	// trg_require_cap_asserted to metaldocs.tenants (INSERT -> tenant.onboard)
	// — the INSERT now requires tenant.onboard asserted tx-locally
	// (seedWithCaps, mirrors testdb.NewTenant / every other tripwire-guarded
	// builder in this package).
	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"tenant.onboard"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO metaldocs.tenants (id, name, slug)
			 VALUES ($1::uuid, $2, $1::text)
			 ON CONFLICT (id) DO NOTHING`,
			tenantID, "Test Tenant "+tenantID,
		)
		return err
	})

	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "iam_users")+` (user_id, display_name, tenant_id)
			 VALUES ($1, $2, $3::uuid)
			 ON CONFLICT (user_id) DO UPDATE SET display_name = EXCLUDED.display_name, tenant_id = EXCLUDED.tenant_id`,
			userID, displayName, tenantID,
		)
		return err
	})

	seedWithCapsIdentity(t, db, tenantID, userID, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO `+Qualified("", "iam_user_roles")+` (user_id, role_code, tenant_id, assigned_by)
			 VALUES ($1, 'system_admin', $2::uuid, $1)
			 ON CONFLICT (tenant_id, user_id) DO UPDATE SET role_code = EXCLUDED.role_code, assigned_by = EXCLUDED.assigned_by`,
			userID, tenantID,
		)
		return err
	})
}

// ---------------------------------------------------------------------------
// Caller-controlled-ID seeders (folded in from tests/integration/fixtures/seed.go,
// TST-12) — deterministic-ID variants used by callers that pin exact IDs (e.g.
// via DeterministicID) instead of taking factory-minted UUIDs. Prefer the New*
// builders above for new tests; these exist for callers that need a specific
// ID/tenant pairing without the CD-governed lineage NewDocument requires.
// ---------------------------------------------------------------------------

// SeedUser inserts an iam_users row with a caller-supplied user ID. SEC-05 /
// migration 0259: iam_users carries trg_require_cap_asserted (user.manage) on
// INSERT — asserted tx-locally via seedWithCaps (pool-safe, matches the
// production authz layer).
func SeedUser(t *testing.T, ctx context.Context, db *sql.DB, schema, userID, displayName string) {
	t.Helper()
	seedWithCaps(t, db, `[{"cap":"user.manage"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s (user_id, display_name, is_active, created_at, updated_at)
			VALUES ($1, $2, true, now(), now())
			ON CONFLICT (user_id) DO NOTHING`,
			Qualified(schema, "iam_users")),
			userID, displayName,
		)
		return err
	})
}

// SeedDocument inserts a minimal documents row with a caller-supplied document
// ID (no CD-governed lineage — use NewDocument when the test needs a
// controlled_documents parent). public.documents carries trg_require_cap_asserted
// (document.create), same tripwire NewDocument asserts — so this goes through
// the identity-aware chokepoint, not the tenant-only sibling.
func SeedDocument(t *testing.T, ctx context.Context, db *sql.DB, schema, docID, tenantID, createdBy string) {
	t.Helper()
	// template_version_id (uuid NOT NULL) and form_data_json (jsonb NOT NULL)
	// carry no FK to templates_template_version on public.documents, so a
	// freshly minted uuid satisfies the NOT NULL without an extra seed
	// (mirrors factory.go's NewDocument free-mint idiom for the same column).
	seedWithCapsIdentity(t, db, tenantID, createdBy, `[{"cap":"document.create"}]`, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s (id, tenant_id, template_version_id, name, status, form_data_json, created_by, revision_version, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, gen_random_uuid(), 'Test Document', 'draft', '{}'::jsonb, $3, 1, now(), now())
			ON CONFLICT (id) DO NOTHING`,
			Qualified(schema, "documents")),
			docID, tenantID, createdBy,
		)
		return err
	})
}

// SeedRouteConfig inserts a minimal approval_routes row with a caller-supplied
// route ID. approval_routes carries FK approval_routes_document_profile_fk
// (tenant_id, profile_code) -> document_profiles(tenant_id, code), so this
// seeds the minimal governed profile the FK requires via SeedGovernedTaxonomy
// (document_families + document_process_areas + document_profiles, asserting
// taxonomy.manage) before inserting the route. profileCode must satisfy
// document_profiles' CHECK constraint (lowercase, ^[a-z][a-z0-9_-]{1,63}$) —
// callers passing an uppercase/mixed-case code will fail that CHECK.
func SeedRouteConfig(t *testing.T, ctx context.Context, db *sql.DB, schema, routeID, tenantID, profileCode string) {
	t.Helper()
	SeedGovernedTaxonomy(t, db, tenantID, profileCode, profileCode)
	seedTenantOnly(t, db, tenantID, "", func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO %s (id, tenant_id, name, profile_code, created_by, active, created_at)
			VALUES ($1::uuid, $2::uuid, 'Test Route', $3, 'system', true, now())
			ON CONFLICT (id) DO NOTHING`,
			Qualified(schema, "approval_routes")),
			routeID, tenantID, profileCode,
		)
		return err
	})
}
