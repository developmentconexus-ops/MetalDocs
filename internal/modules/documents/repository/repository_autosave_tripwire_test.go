package repository

import (
	"os"
	"strings"
	"testing"
)

func TestCommitUpload_AssertsDocumentEditBeforeDocumentsUpdate(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (r *Repository) CommitUpload")
	if start == -1 {
		t.Fatal("CommitUpload not found")
	}
	end := strings.Index(body[start:], "func (r *Repository) DeleteExpiredPending")
	if end == -1 {
		t.Fatal("DeleteExpiredPending not found")
	}
	commitUpload := body[start : start+end]

	tenantGUC := strings.Index(commitUpload, "set_config('metaldocs.tenant_id'")
	if tenantGUC == -1 {
		t.Fatal("CommitUpload must set metaldocs.tenant_id before asserting document.edit")
	}
	actorGUC := strings.Index(commitUpload, "set_config('metaldocs.actor_id'")
	if actorGUC == -1 {
		t.Fatal("CommitUpload must set metaldocs.actor_id before asserting document.edit")
	}
	editRequire := strings.Index(commitUpload, "CapDocumentEdit")
	if editRequire == -1 {
		t.Fatal("CommitUpload must assert document.edit before updating documents")
	}
	updateDocuments := strings.Index(commitUpload, "UPDATE documents SET current_revision_id")
	if updateDocuments == -1 {
		t.Fatal("CommitUpload documents pointer update not found")
	}
	if tenantGUC > editRequire || actorGUC > editRequire {
		t.Fatal("CommitUpload sets authz GUCs after document.edit assertion")
	}
	if editRequire > updateDocuments {
		t.Fatal("CommitUpload asserts document.edit after documents update; tripwire requires it before UPDATE")
	}
}

func TestSessionCleanup_AssertsOrBypassesBeforeClearingDocumentSession(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	cases := []struct {
		name       string
		fn         string
		nextFn     string
		wantMarker string
	}{
		{name: "ReleaseSession", fn: "ReleaseSession", nextFn: "ForceReleaseSession", wantMarker: "CapDocumentEdit"},
		{name: "ForceReleaseSession", fn: "ForceReleaseSession", nextFn: "ExpireStaleSessions", wantMarker: "CapDocumentEdit"},
		{name: "ExpireStaleSessions", fn: "ExpireStaleSessions", nextFn: "PresignReserve", wantMarker: "BypassSystem"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := strings.Index(body, "func (r *Repository) "+tc.fn)
			if start == -1 {
				t.Fatalf("%s not found", tc.fn)
			}
			end := strings.Index(body[start:], "func (r *Repository) "+tc.nextFn)
			if end == -1 {
				t.Fatalf("%s boundary not found", tc.nextFn)
			}
			fnBody := body[start : start+end]

			marker := strings.Index(fnBody, tc.wantMarker)
			if marker == -1 {
				t.Fatalf("%s must establish %s before clearing documents.active_session_id", tc.fn, tc.wantMarker)
			}
			updateDocuments := strings.Index(fnBody, "UPDATE documents SET active_session_id=NULL")
			if updateDocuments == -1 {
				t.Fatalf("%s documents session clear not found", tc.fn)
			}
			if marker > updateDocuments {
				t.Fatalf("%s establishes %s after documents session clear", tc.fn, tc.wantMarker)
			}
		})
	}
}

func TestAcquireSession_AssertsDocumentEditBeforeSettingDocumentSession(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (r *Repository) AcquireSession")
	if start == -1 {
		t.Fatal("AcquireSession not found")
	}
	end := strings.Index(body[start:], "func (r *Repository) HeartbeatSession")
	if end == -1 {
		t.Fatal("HeartbeatSession boundary not found")
	}
	acquire := body[start : start+end]

	tenantGUC := strings.Index(acquire, "set_config('metaldocs.tenant_id'")
	if tenantGUC == -1 {
		t.Fatal("AcquireSession must set metaldocs.tenant_id before asserting document.edit")
	}
	actorGUC := strings.Index(acquire, "set_config('metaldocs.actor_id'")
	if actorGUC == -1 {
		t.Fatal("AcquireSession must set metaldocs.actor_id before asserting document.edit")
	}
	editRequire := strings.Index(acquire, "CapDocumentEdit")
	if editRequire == -1 {
		t.Fatal("AcquireSession must assert document.edit before setting documents.active_session_id")
	}
	updateDocuments := strings.Index(acquire, "UPDATE documents SET active_session_id=$1")
	if updateDocuments == -1 {
		t.Fatal("AcquireSession documents active-session update not found")
	}
	if tenantGUC > editRequire || actorGUC > editRequire {
		t.Fatal("AcquireSession sets authz GUCs after document.edit assertion")
	}
	if editRequire > updateDocuments {
		t.Fatal("AcquireSession asserts document.edit after documents update; tripwire requires it before UPDATE")
	}
}

func TestExpireStaleSessions_ReturnsZeroWhenNothingExpires(t *testing.T) {
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}

	body := string(src)
	start := strings.Index(body, "func (r *Repository) ExpireStaleSessions")
	if start == -1 {
		t.Fatal("ExpireStaleSessions not found")
	}
	end := strings.Index(body[start:], "func (r *Repository) PresignReserve")
	if end == -1 {
		t.Fatal("PresignReserve boundary not found")
	}
	expire := body[start : start+end]

	if strings.Contains(expire, "UPDATE documents SET active_session_id=NULL") &&
		strings.Contains(expire, "RETURNING (SELECT count(*) FROM expired)") {
		t.Fatal("ExpireStaleSessions must not use UPDATE ... RETURNING as the final statement; it returns sql.ErrNoRows when no documents are cleared")
	}
	if !strings.Contains(expire, "SELECT count(*) FROM expired") {
		t.Fatal("ExpireStaleSessions must always SELECT count(*) from expired sessions, including zero")
	}
}
