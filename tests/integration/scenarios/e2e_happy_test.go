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

	client := &http.Client{Timeout: 30 * time.Second}
	tenantID := envOrDefault("METALDOCS_E2E_TENANT_ID", "00000000-0000-0000-0000-000000000001")
	userID := envOrDefault("METALDOCS_E2E_USER_ID", "e2e-user")
	userRoles := envOrDefault("METALDOCS_E2E_USER_ROLES", "admin,document_filler,reviewer,approver")
	routeID := envOrDefault("METALDOCS_E2E_ROUTE_ID", "22222222-2222-2222-2222-222222222222")
	contentHash := strings.Repeat("a", 64)

	var documentID string
	var instanceID string
	var submitETag string
	var stageIDs []string

	// Carried from step 7 into step 7b: the generation the coordinator settled
	// on, and whether it settled by HOLDING (no artifacts in this harness) — the
	// only case where step 7b has work to do.
	var releaseGenerationID string
	var releaseHeldOnMaterializing bool

	// 1) POST /api/v1/controlled-documents -> atomic create (CD + document)
	t.Run("CreateControlledDocument", func(t *testing.T) {
		body := map[string]any{
			"profileCode":     "PO",
			"processAreaCode": "quality",
			"title":           fmt.Sprintf("E2E Happy %d", time.Now().UnixNano()),
			"ownerUserId":     userID,
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost, baseURL+"/api/v1/controlled-documents", body, map[string]string{
			"X-Tenant-ID":     tenantID,
			"X-User-ID":       userID,
			"X-User-Roles":    userRoles,
			"Idempotency-Key": newIdempotencyKey(),
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("atomic create status=%d body=%s", resp.StatusCode, raw)
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
		submitBody := map[string]any{
			"route_id":     routeID,
			"content_hash": contentHash,
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/documents/%s/submit", baseURL, documentID), submitBody, map[string]string{
			"X-Tenant-ID":      tenantID,
			"X-User-ID":        userID,
			"Idempotency-Key":  "11111111-1111-4111-8111-111111111111",
			"If-Match":         "\"v1\"",
			"X-User-Roles":     userRoles,
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
		submitBody := map[string]any{
			"route_id":     routeID,
			"content_hash": contentHash,
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/documents/%s/submit", baseURL, documentID), submitBody, map[string]string{
			"X-Tenant-ID":     tenantID,
			"X-User-ID":       userID,
			"Idempotency-Key": "11111111-1111-4111-8111-111111111111",
			"If-Match":        "\"v1\"",
			"X-User-Roles":    userRoles,
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
		status, raw := getApprovalInstance(t, client, baseURL, tenantID, userID, userRoles, documentID, instanceID)
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

		if stages, ok := payload["stages"].([]any); ok {
			for _, stage := range stages {
				stageMap, ok := stage.(map[string]any)
				if !ok {
					continue
				}
				stageID := asString(stageMap["stage_id"])
				if stageID != "" {
					stageIDs = append(stageIDs, stageID)
				}
			}
		}
	})

	// 4) POST signoff stage 1
	t.Run("SignoffStage1", func(t *testing.T) {
		stageID := stageIDAt(stageIDs, 0, os.Getenv("METALDOCS_E2E_STAGE1_ID"))
		if stageID == "" {
			t.Skip("no stage 1 id found in instance response; set METALDOCS_E2E_STAGE1_ID to force this step")
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost,
			fmt.Sprintf("%s/api/v1/approval/instances/%s/stages/%s/signoffs", baseURL, instanceID, stageID),
			map[string]any{
				"decision":       "approve",
				"password_token": "e2e-token-1",
				"content_hash":   contentHash,
			},
			map[string]string{
				"X-Tenant-ID":     tenantID,
				"X-User-ID":       userID,
				"Idempotency-Key": "22222222-2222-4222-8222-222222222222",
				"If-Match":        submitETag,
				"X-User-Roles":    userRoles,
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("stage1 signoff status=%d body=%s", resp.StatusCode, raw)
		}
	})

	// 5) POST signoff stage 2
	t.Run("SignoffStage2", func(t *testing.T) {
		stageID := stageIDAt(stageIDs, 1, os.Getenv("METALDOCS_E2E_STAGE2_ID"))
		if stageID == "" {
			t.Skip("no stage 2 id found in instance response; set METALDOCS_E2E_STAGE2_ID to force this step")
		}

		resp, raw := doJSONRequest(t, client, http.MethodPost,
			fmt.Sprintf("%s/api/v1/approval/instances/%s/stages/%s/signoffs", baseURL, instanceID, stageID),
			map[string]any{
				"decision":       "approve",
				"password_token": "e2e-token-2",
				"content_hash":   contentHash,
			},
			map[string]string{
				"X-Tenant-ID":     tenantID,
				"X-User-ID":       userID,
				"Idempotency-Key": "33333333-3333-4333-8333-333333333333",
				"If-Match":        submitETag,
				"X-User-Roles":    userRoles,
			},
		)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("stage2 signoff status=%d body=%s", resp.StatusCode, raw)
		}
	})

	// 6) GET approval instance and expect completion
	t.Run("GetApprovalInstanceCompleted", func(t *testing.T) {
		status, raw := getApprovalInstance(t, client, baseURL, tenantID, userID, userRoles, documentID, instanceID)
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

func getApprovalInstance(t *testing.T, client *http.Client, baseURL, tenantID, userID, userRoles, documentID, instanceID string) (int, string) {
	t.Helper()

	headers := map[string]string{
		"X-Tenant-ID":  tenantID,
		"X-User-ID":    userID,
		"X-User-Roles": userRoles,
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
