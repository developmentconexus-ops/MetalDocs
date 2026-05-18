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
