//go:build integration
// +build integration

package scenarios_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	approvalapp "metaldocs/internal/modules/approval/application"
	approvaldomain "metaldocs/internal/modules/approval/domain"
	approvalinfra "metaldocs/internal/modules/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	iampg "metaldocs/internal/modules/iam/infrastructure/postgres"
	platformdb "metaldocs/internal/platform/db"
)

func TestE2E_HappyPath_HTTP(t *testing.T) {
	baseURL := strings.TrimSpace(os.Getenv("METALDOCS_E2E_URL"))
	if baseURL == "" {
		t.Skip("requires running server - set METALDOCS_E2E_URL")
	}

	// Authenticate the way a real client does. The IAM tier-1 middleware
	// deliberately strips inbound X-User-ID/X-User-Roles and resolves identity
	// from the authenticated session context only
	// (internal/modules/iam/delivery/http/middleware.go:81), so header-supplied
	// identity yields 401 AUTH_UNAUTHORIZED against any real stack. The session
	// cookie issued by POST /api/v1/auth/login is the only identity this test
	// may carry; the cookie jar replays it on every subsequent request.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("new cookie jar: %v", err)
	}
	client := &http.Client{Timeout: 30 * time.Second, Jar: jar}

	userID, tenantID := loginE2E(t, client, baseURL)

	// Profile and area come from the server's own creation context rather than
	// from constants: `areas` is already narrowed to the ones where THIS caller
	// holds controlled_documents.create, and `has_active_route` is the flag that
	// decides whether the later submit can resolve a route at all. Hardcoding
	// them makes the test assert against a taxonomy the tenant may not have.
	profileCodes, areaCode := creationContext(t, client, baseURL, userID)

	// Optional pin for environments with more than one active route on the
	// profile; empty means "let the server resolve it in-tx" (ADR 0073).
	routeID := strings.TrimSpace(os.Getenv("METALDOCS_E2E_ROUTE_ID"))

	var documentID string
	var instanceID string
	var contentHash string
	var preSubmitETag string
	var submitETag string
	var instanceETag string
	var stageIDs []string
	var stageActors [][]string

	// Per-run keys. A fixed literal is stored by the idempotency layer against
	// the FIRST run's payload, so every later run against the same database is
	// rejected with 422 IDEMPOTENCY_KEY_REUSED before the handler runs. Replay
	// is still proven: step 2b re-sends submitKey verbatim.
	submitKey := newIdempotencyKey()
	stage1Key := newIdempotencyKey()
	stage2Key := newIdempotencyKey()

	// Carried from step 7 into step 7b: the generation the coordinator settled
	// on, and whether it settled by HOLDING (no artifacts in this harness) — the
	// only case where step 7b has work to do.
	var releaseGenerationID string
	var releaseHeldOnMaterializing bool

	// 1) POST /api/v1/controlled-documents -> atomic create (CD + document)
	t.Run("CreateControlledDocument", func(t *testing.T) {
		// Field names and required set are contract truth
		// (CreateAtomicRequest, api/openapi/v1/openapi.yaml:6429): snake_case,
		// with document_name and visibility required.
		title := fmt.Sprintf("E2E Happy %d", time.Now().UnixNano())

		// Walk the route-bearing profiles until one also has a default template.
		// A 409 state.profile_no_default_template is a property of the tenant's
		// taxonomy, not a failure of the happy path; any other non-201 is.
		var raw string
		var attempts []string
		for _, candidate := range profileCodes {
			body := map[string]any{
				"profile_code":      candidate,
				"process_area_code": areaCode,
				"title":             title,
				"owner_user_id":     userID,
				"document_name":     title,
				"visibility": map[string]any{
					"scope":      "company",
					"area_codes": []string{},
					"user_ids":   []string{},
				},
			}

			var resp *http.Response
			resp, raw = doJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/controlled-documents", body, map[string]string{
				"Idempotency-Key": newIdempotencyKey(),
			})
			if resp.StatusCode == http.StatusCreated {
				break
			}
			attempts = append(attempts, fmt.Sprintf("%s -> %d %s", candidate, resp.StatusCode, raw))
			if resp.StatusCode == http.StatusConflict && strings.Contains(raw, "state.profile_no_default_template") {
				raw = ""
				continue
			}
			t.Fatalf("atomic create status=%d body=%s", resp.StatusCode, raw)
		}
		if raw == "" {
			t.Fatalf("no route-bearing profile also carries a default template; attempts: %s", strings.Join(attempts, "; "))
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("decode create response: %v", err)
		}

		docRef, _ := payload["document"].(map[string]any)
		if docRef == nil {
			t.Fatalf("missing document ref in create response: %s", raw)
		}
		documentID = asString(docRef["id"])
		if documentID == "" {
			t.Fatalf("missing document.id in create response: %s", raw)
		}
		contentHash = asString(docRef["content_hash"])
	})

	// 1b) GET /api/v1/documents/{id} and verify storage_key is populated
	t.Run("VerifyStorageKeyViaDB", func(t *testing.T) {
		db := openRequiredDirectDB(t)
		defer db.Close()

		// public.document_revisions (db/baseline/0001_current_schema.sql:2205),
		// not metaldocs.
		var storageKey string
		if err := db.QueryRowContext(context.Background(), `
			SELECT COALESCE(r.storage_key, '')
			  FROM public.document_revisions r
			 WHERE r.document_id = $1::uuid
			 ORDER BY r.created_at DESC
			 LIMIT 1`,
			documentID,
		).Scan(&storageKey); err != nil {
			t.Fatalf("query revision storage_key: %v", err)
		}
		if storageKey == "" {
			t.Fatalf("expected non-empty storage_key on first revision of atomic-created document %s", documentID)
		}
	})

	// 2) POST /api/v1/documents/{id}/submit with Idempotency-Key + If-Match
	t.Run("SubmitForReview", func(t *testing.T) {
		// route_id and content_hash are both optional (ADR 0073): omitting them
		// makes the server resolve the active route and the head revision hash
		// inside the submit transaction, which is what a real client does and
		// what closes the wrapper-era TOCTOU.
		submitBody := map[string]any{}
		if routeID != "" {
			submitBody["route_id"] = routeID
		}

		// If-Match is the document's CURRENT revision version ("v<N>"); a fresh
		// atomic-created draft is v0, so a hardcoded "v1" is a guaranteed
		// 409 conflict.stale_revision.
		preSubmitETag = documentETag(t, client, baseURL, documentID)

		resp, raw := doJSONRequest(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/documents/%s/submit", baseURL, documentID), submitBody, map[string]string{
			"Idempotency-Key":  submitKey,
			"If-Match":         preSubmitETag,
			"Content-Type":     "application/json",
			"Accept":           "application/json",
			"X-Request-Source": "integration-e2e",
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("submit status=%d body=%s", resp.StatusCode, raw)
		}

		submitETag = strings.TrimSpace(resp.Header.Get("ETag"))
		if submitETag == "" {
			t.Fatalf("submit response missing ETag")
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("decode submit response: %v", err)
		}
		instanceID = asString(payload["instance_id"])
		if instanceID == "" {
			t.Fatalf("missing instance_id in submit response: %s", raw)
		}
	})

	// 2b) Replay submit with same idempotency key; expect replay marker header.
	t.Run("SubmitReplayHeader", func(t *testing.T) {
		// route_id and content_hash are both optional (ADR 0073): omitting them
		// makes the server resolve the active route and the head revision hash
		// inside the submit transaction, which is what a real client does and
		// what closes the wrapper-era TOCTOU.
		submitBody := map[string]any{}
		if routeID != "" {
			submitBody["route_id"] = routeID
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/documents/%s/submit", baseURL, documentID), submitBody, map[string]string{
			// Same key AND same precondition as the original request — a replay
			// is a byte-identical retry, not a fresh submit at the new version.
			"Idempotency-Key": submitKey,
			"If-Match":        preSubmitETag,
		})

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
			t.Fatalf("submit replay unexpected status=%d body=%s", resp.StatusCode, raw)
		}

		replayHeader := strings.TrimSpace(resp.Header.Get("Idempotent-Replay"))
		if replayHeader == "" {
			t.Fatalf("expected Idempotent-Replay header on replay submit; status=%d body=%s", resp.StatusCode, raw)
		}
	})

	// 3) GET /api/v1/documents/{id}/approval-instance (fallback to approval instance route)
	t.Run("GetApprovalInstanceAfterSubmit", func(t *testing.T) {
		status, raw := getApprovalInstance(t, client, baseURL, documentID, instanceID)
		if status != http.StatusOK {
			t.Fatalf("get approval instance status=%d body=%s", status, raw)
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("decode approval instance response: %v", err)
		}

		gotStatus := asString(payload["status"])
		if gotStatus == "" {
			t.Fatalf("missing status in approval instance response: %s", raw)
		}
		if gotStatus != "in_progress" && gotStatus != "approved" {
			t.Fatalf("unexpected instance status after submit: %q body=%s", gotStatus, raw)
		}

		// The signoffs below sign the hash the instance FROZE, not a client-side
		// guess: frozen_content_hash is the pinned digest (no-fallback principle),
		// and the instance etag is the concurrency token those writes must carry.
		if frozen := asString(payload["frozen_content_hash"]); frozen != "" {
			contentHash = frozen
		}
		instanceETag = asString(payload["etag"])

		// Stage identity is `id` per ApprovalStageInstanceResponse
		// (api/openapi/v1/openapi.yaml:6636) — reading "stage_id" silently
		// produced an empty list and skipped both signoff steps.
		if stages, ok := payload["stages"].([]any); ok {
			for _, stage := range stages {
				stageMap, ok := stage.(map[string]any)
				if !ok {
					continue
				}
				stageID := asString(stageMap["id"])
				if stageID == "" {
					continue
				}
				stageIDs = append(stageIDs, stageID)

				// The stage's eligible actors decide WHO may sign it. The
				// submitting session is the author and is SoD-barred from its own
				// approval stage (viewer.eligible_for_active_stage is false), so
				// each signoff runs under the actor's own session.
				var actors []string
				if rawActors, ok := stageMap["actors"].([]any); ok {
					for _, actor := range rawActors {
						actorMap, ok := actor.(map[string]any)
						if !ok {
							continue
						}
						if id := asString(actorMap["user_id"]); id != "" {
							actors = append(actors, id)
						}
					}
				}
				stageActors = append(stageActors, actors)
			}
		}
	})

	// 4) POST signoff stage 1
	t.Run("SignoffStage1", func(t *testing.T) {
		stageID := stageIDAt(stageIDs, 0, os.Getenv("METALDOCS_E2E_STAGE1_ID"))
		if stageID == "" {
			t.Skip("no stage 1 id found in instance response; set METALDOCS_E2E_STAGE1_ID to force this step")
		}

		status, raw := signoffStage(t, baseURL, instanceID, stageID, stageActorsAt(stageActors, 0),
			contentHash, firstNonEmpty(instanceETag, submitETag), stage1Key)
		if status != http.StatusOK {
			t.Fatalf("stage1 signoff status=%d body=%s", status, raw)
		}
	})

	// 5) POST signoff stage 2
	t.Run("SignoffStage2", func(t *testing.T) {
		stageID := stageIDAt(stageIDs, 1, os.Getenv("METALDOCS_E2E_STAGE2_ID"))
		if stageID == "" {
			t.Skip("no stage 2 id found in instance response; set METALDOCS_E2E_STAGE2_ID to force this step")
		}

		status, raw := signoffStage(t, baseURL, instanceID, stageID, stageActorsAt(stageActors, 1),
			contentHash, firstNonEmpty(instanceETag, submitETag), stage2Key)
		if status != http.StatusOK {
			t.Fatalf("stage2 signoff status=%d body=%s", status, raw)
		}
	})

	// 6) GET approval instance and expect completion
	t.Run("GetApprovalInstanceCompleted", func(t *testing.T) {
		status, raw := getApprovalInstance(t, client, baseURL, documentID, instanceID)
		if status != http.StatusOK {
			t.Fatalf("get approval instance status=%d body=%s", status, raw)
		}

		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("decode final instance response: %v", err)
		}

		gotStatus := asString(payload["status"])
		if gotStatus != "approved" && gotStatus != "completed" {
			t.Fatalf("expected completed/approved instance after signoffs, got %q body=%s", gotStatus, raw)
		}
	})

	// 7) Release is NOT an endpoint (ADR 0085): terminal approval writes the
	// approval fact onto the release generation and the coordinator reacts.
	// There is nothing to POST — the assertion is that the durable fact
	// landed and the coordinator reached a legal outcome for this path.
	//
	// This e2e never renders the final DOCX+PDF set, so the artifact fact is
	// absent by construction and the coordinator's honest outcome is the
	// `materializing` hold. Asserting "published" here would assert a bug.
	t.Run("ReleaseGenerationApprovalFactDB", func(t *testing.T) {
		db := openRequiredDirectDB(t)
		defer db.Close()

		var (
			generationID   string
			approvalFactAt sql.NullTime
			artifactFactAt sql.NullTime
			releasedAt     sql.NullTime
			holdReason     sql.NullString
			docStatus      string
		)

		// The coordinator runs asynchronously (release_evaluate River job), so
		// poll rather than read once.
		deadline := time.Now().Add(30 * time.Second)
		for {
			err := db.QueryRowContext(context.Background(), `
				SELECT g.id::text,
				       g.approval_fact_at,
				       g.artifact_fact_at,
				       g.released_at,
				       g.hold_reason,
				       d.status
				  FROM public.release_generations g
				  JOIN public.documents d
				    ON d.id = g.document_id
				   AND d.tenant_id = g.tenant_id
				 WHERE g.tenant_id = $1::uuid
				   AND g.document_id = $2::uuid
				 ORDER BY g.generation_seq DESC
				 LIMIT 1`,
				tenantID, documentID,
			).Scan(&generationID, &approvalFactAt, &artifactFactAt, &releasedAt, &holdReason, &docStatus)
			if err == nil && approvalFactAt.Valid && (holdReason.Valid || releasedAt.Valid) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("release generation never settled for document %s (err=%v approval_fact=%v hold=%v released=%v)",
					documentID, err, approvalFactAt.Valid, holdReason.String, releasedAt.Valid)
			}
			time.Sleep(500 * time.Millisecond)
		}

		if releasedAt.Valid {
			// Only legal if the artifact fact really did land (a renderer is
			// running against this server); then the document must be published.
			if !artifactFactAt.Valid {
				t.Fatalf("released without an artifact fact: released_at=%v", releasedAt.Time)
			}
			if docStatus != "published" {
				t.Fatalf("released generation but document status=%q, want published", docStatus)
			}
			return
		}

		if holdReason.String != "materializing" {
			t.Fatalf("hold_reason=%q, want materializing (no artifact rendered in this e2e path)", holdReason.String)
		}
		if docStatus != "approved" && docStatus != "scheduled" {
			t.Fatalf("document status=%q while held on %q, want approved/scheduled", docStatus, holdReason.String)
		}

		releaseGenerationID = generationID
		releaseHeldOnMaterializing = true
	})

	// 7b) The other half of ADR 0085, and the half step 7 structurally cannot
	// reach: the coordinator RELEASES once the generation is artifact-ready.
	//
	// Step 7 stops at the honest `materializing` hold because this harness runs
	// no renderer, so asserting "published" there would assert a bug. That makes
	// the happy path's terminal state unproven end-to-end. This subtest closes
	// it deterministically: it completes the readiness predicate through the
	// SAME production fact producer the render workers use
	// (application.RecordArtifactFactTx — generation-guarded and idempotent, not
	// a raw UPDATE), then drives one evaluation through the real coordinator.
	//
	// The coordinator is constructed here rather than driven through River: the
	// River leg is a thin InsertTx/Work wrapper already covered by
	// tests/integration/approval/release_coordinator_integration_test.go, and
	// invoking Evaluate directly makes this proof deterministic instead of
	// racing a background queue.
	t.Run("CoordinatorReleasesOnceArtifactFactsLand", func(t *testing.T) {
		if !releaseHeldOnMaterializing {
			t.Skip("generation already released (this environment renders artifacts); the materializing half of step 7 did not apply, so there is nothing to complete")
		}

		database := openRequiredDirectDB(t)
		defer database.Close()

		ctx := context.Background()
		runner := platformdb.NewTxRunner(database)
		repo := approvalinfra.NewPostgresApprovalRepository(database, iampg.NewUserDisplayNameRepository(database))
		enqueuer := &e2eReleaseEnqueuer{}
		lifecycle := &e2eLifecycleEnqueuer{}
		coordinator := approvalapp.NewReleaseCoordinator(repo, approvalapp.NewSQLEmitter(), approvalapp.RealClock{}).
			WithEvaluationEnqueuer(enqueuer).
			WithLifecycleEnqueuer(lifecycle)

		// Complete the artifact set. Each half lands in its own transaction,
		// the way the DOCX and PDF workers write them.
		for _, kind := range []approvalapp.ArtifactKind{approvalapp.ArtifactFinalDocx, approvalapp.ArtifactFinalPDF} {
			kind := kind
			if err := runner.Do(ctx, func(tx *sql.Tx) error {
				_, err := approvalapp.RecordArtifactFactTx(ctx, tx, enqueuer, approvalapp.ArtifactFactInput{
					TenantID:            tenantID,
					ReleaseGenerationID: releaseGenerationID,
					Kind:                kind,
					S3Key:               fmt.Sprintf("tenants/%s/%s/%s", tenantID, documentID, kind),
				})
				return err
			}); err != nil {
				t.Fatalf("record %s artifact fact for generation %s: %v", kind, releaseGenerationID, err)
			}
		}

		key, artifactFactAt := loadReleaseGenerationIdentity(t, database, tenantID, documentID)
		if !artifactFactAt.Valid {
			t.Fatal("artifact_fact_at not stamped after both halves landed — artifact-ready(G) requires DOCX AND PDF")
		}

		eval, err := coordinator.Evaluate(ctx, runner, key)
		if err != nil {
			t.Fatalf("Evaluate (artifact-ready): %v", err)
		}
		if eval.Action != approvaldomain.ReleaseActionRelease {
			t.Fatalf("evaluation = %+v, want release (predicate: approval fact x artifact facts x effective date x supersession head)", eval)
		}

		// Document truth.
		var (
			docStatus     string
			effectiveFrom sql.NullTime
		)
		if err := database.QueryRowContext(ctx,
			`SELECT status, effective_from FROM public.documents WHERE id = $1::uuid AND tenant_id = $2::uuid`,
			documentID, tenantID,
		).Scan(&docStatus, &effectiveFrom); err != nil {
			t.Fatalf("load released document: %v", err)
		}
		if docStatus != "published" {
			t.Fatalf("document status=%q, want published", docStatus)
		}
		if !effectiveFrom.Valid {
			t.Fatal("published document has no effective_from — the release must stamp the ACTUAL release instant")
		}

		// Generation truth.
		var (
			releasedAt sql.NullTime
			holdReason sql.NullString
		)
		if err := database.QueryRowContext(ctx,
			`SELECT released_at, hold_reason FROM public.release_generations WHERE id = $1::uuid`,
			releaseGenerationID,
		).Scan(&releasedAt, &holdReason); err != nil {
			t.Fatalf("load released generation: %v", err)
		}
		if !releasedAt.Valid {
			t.Fatal("release generation not stamped released_at")
		}
		if holdReason.Valid {
			t.Fatalf("released generation still carries hold %q", holdReason.String)
		}

		// Exactly one governance event for the transition, attributed to the
		// system principal — publication has no human actor under ADR 0085.
		var governanceEvents int
		if err := database.QueryRowContext(ctx, `
			SELECT count(*) FROM public.governance_events
			 WHERE tenant_id = $1::uuid
			   AND resource_id = $2
			   AND event_type = 'document_published'`,
			tenantID, documentID,
		).Scan(&governanceEvents); err != nil {
			t.Fatalf("count document_published governance events: %v", err)
		}
		if governanceEvents != 1 {
			t.Fatalf("document_published governance events = %d, want exactly 1", governanceEvents)
		}
		var eventActor string
		if err := database.QueryRowContext(ctx, `
			SELECT actor_user_id FROM public.governance_events
			 WHERE tenant_id = $1::uuid AND resource_id = $2 AND event_type = 'document_published'
			 ORDER BY created_at DESC LIMIT 1`,
			tenantID, documentID,
		).Scan(&eventActor); err != nil {
			t.Fatalf("load document_published governance event: %v", err)
		}
		if eventActor != approvalapp.ReleaseSystemPrincipal {
			t.Fatalf("release event actor = %q, want %q", eventActor, approvalapp.ReleaseSystemPrincipal)
		}

		// Exactly one lifecycle (notification-fanout) event for the same
		// transition — one governance record and one reader notification, never
		// two of either.
		if n := lifecycle.count(documentID, docsdomain.EventTypeDocumentPublished); n != 1 {
			t.Fatalf("published lifecycle events for %s = %d, want exactly 1", documentID, n)
		}
	})

	// 8) DB-level governance_events assertion. METALDOCS_E2E_URL is set (the
	// outer skip already fired otherwise), so the DB DSN pointing at the same
	// server's database is required, not optional.
	t.Run("GovernanceEventsCountDB", func(t *testing.T) {
		db := openRequiredDirectDB(t)
		defer db.Close()

		// public.governance_events (db/baseline/0001_current_schema.sql:2268),
		// not metaldocs.
		var count int
		if err := db.QueryRowContext(context.Background(), `
			SELECT count(*)
			  FROM public.governance_events
			 WHERE tenant_id = $1::uuid
			   AND resource_type = 'document'
			   AND resource_id = $2`,
			tenantID, documentID,
		).Scan(&count); err != nil {
			t.Fatalf("query governance_events count: %v", err)
		}
		if count < 1 {
			t.Fatalf("expected at least one governance event for document %s, got %d", documentID, count)
		}
	})
}

// ── release-coordinator test doubles (step 7b) ──────────────────────────────

// e2eReleaseEnqueuer stands in for the River evaluation enqueuer. Step 7b
// invokes Evaluate directly, so the outbox leg only needs to not be the reason
// a fact write fails; what it records is used for nothing but debuggability.
type e2eReleaseEnqueuer struct {
	mu   sync.Mutex
	keys []approvaldomain.ReleaseGenerationKey
}

func (e *e2eReleaseEnqueuer) EnqueueReleaseEvaluationTx(_ context.Context, _ platformdb.Tx, key approvaldomain.ReleaseGenerationKey, _ time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.keys = append(e.keys, key)
	return nil
}

// e2eLifecycleEnqueuer captures the notification-fanout rows the release
// transaction enqueues. The River leg is a thin InsertTx wrapper; what step 7b
// asserts is that the release emits exactly one lifecycle event per affected
// transition.
type e2eLifecycleEnqueuer struct {
	mu     sync.Mutex
	events []docsdomain.LifecycleEventArgs
}

func (e *e2eLifecycleEnqueuer) EnqueueLifecycleEventTx(_ context.Context, _ platformdb.Tx, args docsdomain.LifecycleEventArgs) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, args)
	return nil
}

func (e *e2eLifecycleEnqueuer) count(resourceID, eventType string) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	n := 0
	for _, ev := range e.events {
		if ev.ResourceID == resourceID && ev.EventType == eventType {
			n++
		}
	}
	return n
}

// loadReleaseGenerationIdentity reads the document's newest release generation
// and returns the coordinator key for it, plus artifact_fact_at. The key is
// read back from the row rather than reconstructed from the document, because
// the release itself bumps revision_version — driving Evaluate from the
// document's CURRENT version would key on a generation that does not exist.
func loadReleaseGenerationIdentity(t *testing.T, database *sql.DB, tenantID, documentID string) (approvaldomain.ReleaseGenerationKey, sql.NullTime) {
	t.Helper()

	var (
		key            approvaldomain.ReleaseGenerationKey
		artifactFactAt sql.NullTime
	)
	if err := database.QueryRowContext(context.Background(), `
		SELECT tenant_id::text, document_id::text, approval_instance_id::text,
		       revision_id::text, revision_version, frozen_content_hash, artifact_fact_at
		  FROM public.release_generations
		 WHERE tenant_id = $1::uuid
		   AND document_id = $2::uuid
		 ORDER BY generation_seq DESC
		 LIMIT 1`,
		tenantID, documentID,
	).Scan(&key.TenantID, &key.DocumentID, &key.ApprovalInstanceID,
		&key.RevisionID, &key.RevisionVersion, &key.FrozenContentHash, &artifactFactAt); err != nil {
		t.Fatalf("load release generation identity for %s: %v", documentID, err)
	}
	key.SubjectKind = approvaldomain.SubjectKindDocument
	return key, artifactFactAt
}

// newIdempotencyKey returns a fresh UUID. The API-wide Idempotency-Key wire
// rule is UUID-only (F-QA4-6: idempotency.ValidateKey + `format: uuid` on every
// spec'd Idempotency-Key parameter), so a free-string key like "e2e-create-<ns>"
// is rejected with 400 before the handler runs. Steps that must not replay
// across runs use this; steps that deliberately re-send the same key to prove
// replay use a fixed UUID literal.
func newIdempotencyKey() string {
	return uuid.NewString()
}

// loginE2E authenticates against the live server the way real clients do —
// POST /api/v1/auth/login with the dev-seed credentials
// (wiki/references/local-dev-startup.md "Credentials"; body field is
// `identifier`, NOT `username`) — and returns the session's user id and tenant
// id. The session cookie lands in client.Jar and is replayed on every later
// request, which is the ONLY identity the IAM middleware honors.
//
// The returned tenant id is also what the DB-verification subtests key on:
// reading it back from the session guarantees the rows asserted belong to the
// tenant the server actually wrote under, instead of a guessed constant.
func loginE2E(t *testing.T, client *http.Client, baseURL string) (userID, tenantID string) {
	t.Helper()

	identifier := envOrDefault("METALDOCS_E2E_IDENTIFIER", "admin")
	password := envOrDefault("METALDOCS_E2E_PASSWORD", "AdminMetalDocs123!")

	resp, raw := doJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
		"identifier": identifier,
		"password":   password,
	}, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login as %q status=%d body=%s", identifier, resp.StatusCode, raw)
	}

	var payload struct {
		User struct {
			UserID   string `json:"user_id"`
			TenantID string `json:"tenant_id"`
		} `json:"user"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	userID = strings.TrimSpace(payload.User.UserID)
	tenantID = strings.TrimSpace(payload.User.TenantID)
	if userID == "" || tenantID == "" {
		t.Fatalf("login response missing user_id/tenant_id: %s", raw)
	}

	// A 200 without a session cookie is a broken environment, not a pass: every
	// subsequent request would fall back to anonymous and 401.
	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse base URL %q: %v", baseURL, err)
	}
	if len(client.Jar.Cookies(parsed)) == 0 {
		t.Fatalf("login succeeded but no session cookie was stored for %s", baseURL)
	}
	return userID, tenantID
}

// creationContext asks the server which profile/area this session may actually
// create a controlled document under. `areas` is already filtered server-side to
// the areas where the caller holds controlled_documents.create, and
// has_active_route tells whether a submit under that profile can resolve a
// route at all — so picking from this response is the only way to choose inputs
// that are true for the tenant AND for the authenticated caller.
// It returns EVERY route-bearing profile, not just the first: an active route
// is necessary but not sufficient — the atomic create also needs the profile to
// carry a default template (409 state.profile_no_default_template otherwise), and the
// creation-context contract exposes no flag for that.
func creationContext(t *testing.T, client *http.Client, baseURL, author string) (profileCodes []string, areaCode string) {
	t.Helper()

	resp, raw := doJSONRequest(t, client, http.MethodGet, baseURL+"/api/v1/controlled-documents/creation-context", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("creation-context status=%d body=%s", resp.StatusCode, raw)
	}

	var payload struct {
		Profiles []struct {
			Code           string `json:"code"`
			HasActiveRoute bool   `json:"has_active_route"`
		} `json:"profiles"`
		Areas []struct {
			Code string `json:"code"`
		} `json:"areas"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode creation-context: %v", err)
	}

	if forced := strings.TrimSpace(os.Getenv("METALDOCS_E2E_PROFILE_CODE")); forced != "" {
		profileCodes = []string{forced}
	} else {
		for _, p := range payload.Profiles {
			if p.HasActiveRoute && strings.TrimSpace(p.Code) != "" {
				profileCodes = append(profileCodes, strings.TrimSpace(p.Code))
			}
		}
	}
	if len(profileCodes) == 0 {
		t.Fatalf("no profile with an active approval route in creation-context: %s", raw)
	}

	if forced := strings.TrimSpace(os.Getenv("METALDOCS_E2E_AREA_CODE")); forced != "" {
		areaCode = forced
	} else {
		// The area decides who may SIGN: document.signoff is capability x area,
		// so an area where no credentialed approver holds a membership produces a
		// route whose named actor is denied at tier-2 (403 authz.capability_denied)
		// and a happy path that can never complete. Prefer an area that has one.
		signable := areasWithSignableApprover(t, client, baseURL, author)
		for _, area := range payload.Areas {
			if _, ok := signable[strings.TrimSpace(area.Code)]; ok {
				areaCode = strings.TrimSpace(area.Code)
				break
			}
		}
		if areaCode == "" && len(payload.Areas) > 0 {
			areaCode = strings.TrimSpace(payload.Areas[0].Code)
			t.Logf("no creation-context area has a credentialed approver membership; falling back to %q", areaCode)
		}
	}
	if areaCode == "" {
		t.Fatalf("session holds controlled_documents.create in no process area: %s", raw)
	}
	return profileCodes, areaCode
}

// devSeedCredentials is the sanctioned local QA login set
// (wiki/references/local-dev-startup.md "Credentials" / the same values seeded
// by db/dev-seeds/0001_local_dev_seed.sql). Never read from .env — these are
// documented dev-only accounts, not deployment secrets. Override for a stack
// whose personas differ with METALDOCS_E2E_CREDENTIALS="user:pass,user:pass".
var devSeedCredentials = map[string]string{
	"admin":         "AdminMetalDocs123!",
	"approver":      "ApproverMetalDocs123!",
	"author-test":   "AuthorTest123!",
	"approver-test": "ApproverMetalDocs456!@",
}

func credentialFor(identifier string) (string, bool) {
	for _, pair := range strings.Split(os.Getenv("METALDOCS_E2E_CREDENTIALS"), ",") {
		user, pass, ok := strings.Cut(strings.TrimSpace(pair), ":")
		if ok && strings.TrimSpace(user) == identifier {
			return pass, true
		}
	}
	pass, ok := devSeedCredentials[identifier]
	return pass, ok
}

// signoffStage records one stage signoff as one of the stage's OWN eligible
// actors. The submitting session cannot do this: it is the author, and SoD bars
// an author from signing their own document — so the signer gets its own client,
// its own login, and its own session cookie. password_token is that signer's
// real re-authentication password (password_reauth method); a placeholder string
// is rejected with 401 authn.signature_invalid.
func signoffStage(t *testing.T, baseURL, instanceID, stageID string, actors []string, contentHash, fallbackETag, idempotencyKey string) (int, string) {
	t.Helper()

	for _, actor := range actors {
		password, ok := credentialFor(actor)
		if !ok {
			continue
		}

		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatalf("new cookie jar for %s: %v", actor, err)
		}
		signer := &http.Client{Timeout: 30 * time.Second, Jar: jar}

		resp, raw := doJSONRequest(t, signer, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
			"identifier": actor,
			"password":   password,
		}, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("login as stage actor %q status=%d body=%s", actor, resp.StatusCode, raw)
		}

		// Re-read the instance through the signer's session: the If-Match token
		// must be the version THIS writer observed, not one carried over from a
		// stage another actor already advanced.
		etag := fallbackETag
		if status, instanceRaw := getApprovalInstance(t, signer, baseURL, "", instanceID); status == http.StatusOK {
			var payload map[string]any
			if err := json.Unmarshal([]byte(instanceRaw), &payload); err == nil {
				etag = firstNonEmpty(asString(payload["etag"]), fallbackETag)
			}
		}

		resp, raw = doJSONRequest(t, signer, http.MethodPost,
			fmt.Sprintf("%s/api/v1/approval/instances/%s/stages/%s/signoffs", baseURL, instanceID, stageID),
			map[string]any{
				"decision":       "approve",
				"password_token": password,
				"content_hash":   contentHash,
			},
			map[string]string{
				"Idempotency-Key": idempotencyKey,
				"If-Match":        etag,
			},
		)
		return resp.StatusCode, raw
	}

	t.Fatalf("no credentials available for any eligible actor of stage %s (actors=%v); set METALDOCS_E2E_CREDENTIALS", stageID, actors)
	return 0, ""
}

func stageActorsAt(actors [][]string, idx int) []string {
	if idx >= 0 && idx < len(actors) {
		return actors[idx]
	}
	return nil
}

// documentETag reads the document's current revision_version and renders it as
// the If-Match token the write endpoints expect (`"v<N>"`, N >= 0 — a fresh
// draft is v0).
func documentETag(t *testing.T, client *http.Client, baseURL, documentID string) string {
	t.Helper()

	resp, raw := doJSONRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/documents/%s", baseURL, documentID), nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get document %s status=%d body=%s", documentID, resp.StatusCode, raw)
	}

	var payload struct {
		RevisionVersion *int `json:"revision_version"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode document %s: %v", documentID, err)
	}
	if payload.RevisionVersion == nil {
		t.Fatalf("document %s carries no revision_version: %s", documentID, raw)
	}
	return fmt.Sprintf("%q", fmt.Sprintf("v%d", *payload.RevisionVersion))
}

// areasWithSignableApprover returns the areas that hold an active approver
// membership for a user this test can actually log in as, excluding the author
// (SoD bars an author from signing their own document).
func areasWithSignableApprover(t *testing.T, client *http.Client, baseURL, author string) map[string]struct{} {
	t.Helper()

	out := map[string]struct{}{}
	resp, raw := doJSONRequest(t, client, http.MethodGet, baseURL+"/api/v1/iam/area-memberships?role=approver", nil, nil)
	if resp.StatusCode != http.StatusOK {
		t.Logf("area-memberships lookup status=%d body=%s — area choice falls back to creation-context order", resp.StatusCode, raw)
		return out
	}

	var payload struct {
		Items []struct {
			UserID   string `json:"user_id"`
			AreaCode string `json:"area_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode area-memberships: %v", err)
	}
	for _, item := range payload.Items {
		if item.UserID == author {
			continue
		}
		if _, ok := credentialFor(item.UserID); !ok {
			continue
		}
		out[strings.TrimSpace(item.AreaCode)] = struct{}{}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// requestOrigin reduces a request URL to its scheme://host origin.
func requestOrigin(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func doJSONRequest(t *testing.T, client *http.Client, method, url string, body any, headers map[string]string) (*http.Response, string) {
	t.Helper()

	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.Fatalf("new request %s %s: %v", method, url, err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Origin protection (security.OriginProtection) rejects any cookie-bearing
	// POST/PUT/PATCH/DELETE that arrives without a same-origin Origin/Referer.
	// A real browser client always sends one; so must this test.
	if origin := requestOrigin(url); origin != "" {
		req.Header.Set("Origin", origin)
	}
	for k, v := range headers {
		if strings.TrimSpace(v) != "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("http %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return resp, strings.TrimSpace(string(raw))
}

func getApprovalInstance(t *testing.T, client *http.Client, baseURL, documentID, instanceID string) (int, string) {
	t.Helper()

	// Identity rides the session cookie on client's jar — no identity headers.
	headers := map[string]string{}

	// Callers that hold only the instance id (a stage signer) skip the
	// by-document lookup entirely rather than requesting an empty path segment.
	if strings.TrimSpace(documentID) == "" && strings.TrimSpace(instanceID) != "" {
		resp, raw := doJSONRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/approval/instances/%s", baseURL, instanceID), nil, headers)
		return resp.StatusCode, raw
	}

	resp, raw := doJSONRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/documents/%s/approval-instance", baseURL, documentID), nil, headers)
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
		return resp.StatusCode, raw
	}

	if strings.TrimSpace(instanceID) == "" {
		return resp.StatusCode, raw
	}

	resp2, raw2 := doJSONRequest(t, client, http.MethodGet, fmt.Sprintf("%s/api/v1/approval/instances/%s", baseURL, instanceID), nil, headers)
	return resp2.StatusCode, raw2
}

// openRequiredDirectDB opens a raw DSN connection to the database the live
// E2E server under test actually writes to. It deliberately does NOT go
// through testdb.Open: a leased factory database is a different physical
// database than the one METALDOCS_E2E_URL's server connected to at startup,
// so routing these DB-verification subtests through the factory would assert
// against rows the server never wrote. TestE2E_HappyPath_HTTP already gated
// the whole test on METALDOCS_E2E_URL being set, so once a subtest here runs,
// the DB DSN naming that same server's database is required, not optional —
// a missing or unreachable DSN is a real gap in the E2E environment and must
// fail loud, not read as a skip-shaped green.
func openRequiredDirectDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("METALDOCS_DATABASE_URL"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		t.Fatal("METALDOCS_E2E_URL is set but DATABASE_URL/METALDOCS_DATABASE_URL is not: DB-verification subtests require a direct connection to the server's database")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		t.Fatalf("integration DB unreachable: %v", err)
	}
	return db
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func asString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func stageIDAt(stageIDs []string, idx int, fallback string) string {
	if idx >= 0 && idx < len(stageIDs) && strings.TrimSpace(stageIDs[idx]) != "" {
		return strings.TrimSpace(stageIDs[idx])
	}
	return strings.TrimSpace(fallback)
}
