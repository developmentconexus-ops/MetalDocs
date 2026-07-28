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
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	// 7) POST /api/v1/documents/{id}/publish
	t.Run("Publish", func(t *testing.T) {
		resp, raw := doJSONRequest(t, client, http.MethodPost, fmt.Sprintf("%s/api/v1/documents/%s/publish", baseURL, documentID), nil, map[string]string{
			"X-Tenant-ID":     tenantID,
			"X-User-ID":       userID,
			"Idempotency-Key": "44444444-4444-4444-8444-444444444444",
			"If-Match":        submitETag,
			"X-User-Roles":    userRoles,
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("publish status=%d body=%s", resp.StatusCode, raw)
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
