//go:build integration
// +build integration

// Package application_test contains SP-2 D8 integration proof:
// a pinned source='dictionary' placeholder value (a) substitutes by NAME into the
// FanoutRequest and (b) is captured in the documents.values_hash column after Freeze.
package application_test

import (
	"context"
	"database/sql"
	"testing"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/application"
	docrepo "metaldocs/internal/modules/documents/repository"
	"metaldocs/internal/modules/render/fanout"
	"metaldocs/internal/modules/render/resolvers"
	"metaldocs/tests/integration/testdb"
)

// integFakeFanoutClient captures the last FanoutRequest without hitting the renderer.
type integFakeFanoutClient struct {
	req  fanout.FanoutRequest
	resp fanout.FanoutResponse
}

func (f *integFakeFanoutClient) Fanout(_ context.Context, req fanout.FanoutRequest) (fanout.FanoutResponse, error) {
	f.req = req
	return f.resp, nil
}

// integFakeCtxBuilder is a minimal ResolverContextBuilder for the integration test.
// The placeholder under test is non-computed so Build is never called to drive a
// resolver — but FreezeService always calls Build once before the resolver loop.
type integFakeCtxBuilder struct{}

func (integFakeCtxBuilder) Build(_ context.Context, tenantID, revisionID string, _ application.ApproverContext) (resolvers.ResolveInput, error) {
	return resolvers.ResolveInput{TenantID: resolvers.TenantID(tenantID), RevisionID: resolvers.RevisionID(revisionID)}, nil
}

func (integFakeCtxBuilder) BuildForDraft(_ context.Context, tenantID, revisionID string) (resolvers.ResolveInput, error) {
	return resolvers.ResolveInput{TenantID: resolvers.TenantID(tenantID), RevisionID: resolvers.RevisionID(revisionID)}, nil
}

// TestSP2_D8_DictionarySubstitutionViaFreezePath is the integration proof of SP-2 D8:
//   - A revision that has a persisted document_placeholder_values row with
//     source='dictionary' for placeholder "company_name" (pinned at creation time)
//     flows through Freeze so that:
//     (a) fanout.FanoutRequest.PlaceholderValues["company_name"] == pinned value, and
//     (b) the values_hash column on the revision is non-empty (dictionary value was
//     captured in the deterministic hash).
//
// A fake FanoutClient is injected — no real docx-renderer call is made.
func TestSP2_D8_DictionarySubstitutionViaFreezePath(t *testing.T) {
	ctx := context.Background()
	db, dbName := testdb.Open(t)

	// Mirror the search_path + pool setup from openDictDB (defined in sp2_dictionary_pin_integration_test.go).
	if _, err := db.ExecContext(ctx, `ALTER DATABASE "`+dbName+`" SET search_path TO public, metaldocs`); err != nil {
		t.Fatalf("alter search_path: %v", err)
	}
	db.SetMaxIdleConns(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	tenant := testdb.NewTenant(t, db)
	tenantID := tenant.ID
	actorID := testdb.DeterministicID(t, "actor-d8-subst")
	testdb.SeedSystemAdmin(t, db, tenantID, actorID, "D8 Author")
	testdb.SeedGovernedTaxonomy(t, db, tenantID, "po", "quality")

	cdFixture := testdb.NewControlledDoc(t, db,
		testdb.WithTenant(tenantID),
		testdb.WithTaxonomy(testdb.Taxonomy{TenantID: tenantID, ProfileCode: "po", ProcessAreaCode: "quality"}),
		testdb.WithCode("PO-D8-001"),
		testdb.WithOwner(actorID),
	)

	const phID = "ph-d8-company"
	const phName = "company_name"
	const pinnedValue = "Acme Corp"

	// Seed template with a single non-computed placeholder named "company_name".
	tplVersionID := seedDictTemplate(t, db, tenantID, actorID, "d8-subst-test", phID, phName)

	// Pin the dictionary value at creation time (D1 path, same as Task 5).
	dictReader := newFakeDictLookup()
	dictReader.set(tenantID, phName, pinnedValue)

	svc := newDictionarySvcForIntegration(db, dictReader, tplVersionID)
	initializer := application.NewCDDocumentInitializer(svc)

	cd := &controlleddocumentsdomain.ControlledDocument{
		ID:              cdFixture.ID,
		TenantID:        tenantID,
		ProfileCode:     cdFixture.ProfileCode,
		ProcessAreaCode: cdFixture.ProcessAreaCode,
		Code:            cdFixture.Code,
		OwnerUserID:     actorID,
		Status:          controlleddocumentsdomain.CDStatusActive,
	}

	dv, err := initializer.ResolveDictionaryValues(ctx, tenantID, tplVersionID)
	if err != nil {
		t.Fatalf("ResolveDictionaryValues: %v", err)
	}
	req, err := controlleddocumentsdomain.NewCloneTemplateRequest(&tplVersionID, "D8 Substitution Test", nil)
	if err != nil {
		t.Fatalf("NewCloneTemplateRequest: %v", err)
	}
	req = req.WithDictionaryValues(dv)

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin create tx: %v", err)
	}
	ref, err := initializer.CloneTemplate(ctx, tx1, cd, req)
	if err != nil {
		tx1.Rollback()
		t.Fatalf("CloneTemplate: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit create: %v", err)
	}
	revisionID := ref.RevisionID

	// Confirm the pin row exists before freeze.
	source, vt := queryPlaceholderRow(t, db, tenantID, revisionID, phID)
	if source != "dictionary" || !vt.Valid || vt.String != pinnedValue {
		t.Fatalf("pre-freeze: want source=dictionary value=%s, got source=%q value=%+v", pinnedValue, source, vt)
	}

	// Wire a real FreezeService with real repos but a fake fanout (no renderer).
	// The fake fanout returns a deterministic content_hash so WriteFinalDocx can decode it.
	snapRepo := docrepo.NewSnapshotRepository(db)
	fillInRepo := docrepo.NewFillInRepository(db)
	schemaReader := application.NewSnapshotSchemaReader(db)
	capturedFanout := &integFakeFanoutClient{
		resp: fanout.FanoutResponse{
			ContentHash:    "aabbccdd00000000000000000000000000000000000000000000000000000000",
			FinalDocxS3Key: "final/" + revisionID + ".docx",
			UnreplacedVars: []string{},
		},
	}
	reg := resolvers.NewRegistry()
	resolvers.RegisterBuiltins(reg)

	freezeSvc := application.NewFreezeService(
		schemaReader,
		fillInRepo,
		fillInRepo,
		reg,
		snapRepo,
		integFakeCtxBuilder{},
		snapRepo,
		snapRepo,
		capturedFanout,
	)

	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin freeze tx: %v", err)
	}
	defer tx2.Rollback()

	if err := freezeSvc.Freeze(ctx, tx2, tenantID, revisionID, application.ApproverContext{}); err != nil {
		t.Fatalf("Freeze: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit freeze: %v", err)
	}

	// D8-substitution: fanout request must carry the pinned dictionary value keyed by NAME.
	if got := capturedFanout.req.PlaceholderValues[phName]; got != pinnedValue {
		t.Errorf("D8 substitution: fanout PlaceholderValues[%q] = %q, want %q", phName, got, pinnedValue)
	}

	// D8-hash: values_hash column must be non-empty after Freeze (dictionary value captured).
	var valuesHash []byte
	if err := db.QueryRowContext(ctx,
		`SELECT values_hash FROM public.documents WHERE tenant_id = $1::uuid AND id = $2::uuid`,
		tenantID, revisionID,
	).Scan(&valuesHash); err != nil {
		t.Fatalf("query values_hash: %v", err)
	}
	if len(valuesHash) == 0 {
		t.Error("D8 hash: values_hash is empty after Freeze — dictionary value not captured")
	}
}

// queryValuesHash reads values_hash from the documents row for the given revision.
// Exposed as a helper for readability; same super-user read pattern as queryPlaceholderRow.
func queryValuesHash(t *testing.T, db *sql.DB, tenantID, revisionID string) []byte {
	t.Helper()
	var h []byte
	if err := db.QueryRowContext(context.Background(),
		`SELECT values_hash FROM public.documents WHERE tenant_id = $1::uuid AND id = $2::uuid`,
		tenantID, revisionID,
	).Scan(&h); err != nil {
		t.Fatalf("queryValuesHash(%s, %s): %v", tenantID, revisionID, err)
	}
	return h
}
