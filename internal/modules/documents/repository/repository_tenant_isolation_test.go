package repository

import (
	"os"
	"strings"
	"testing"
)

func TestSessionQueriesIncludeTenantScope(t *testing.T) {
	src := readRepositorySource(t)

	cases := []struct {
		name   string
		fn     string
		nextFn string
		want   string
	}{
		{name: "CreateDocumentTx", fn: "CreateDocumentTx", nextFn: "SetRevisionStorageKey", want: "INSERT INTO editor_sessions (tenant_id"},
		{name: "AcquireSession", fn: "AcquireSession", nextFn: "HeartbeatSession", want: "tenant_id=$1::uuid"},
		{name: "HeartbeatSession", fn: "HeartbeatSession", nextFn: "ReleaseSession", want: "AND tenant_id=$2::uuid"},
		{name: "ReleaseSession", fn: "ReleaseSession", nextFn: "ForceReleaseSession", want: "AND tenant_id=$2::uuid"},
		{name: "ForceReleaseSession", fn: "ForceReleaseSession", nextFn: "ExpireStaleSessions", want: "AND tenant_id=$2::uuid"},
		{name: "PresignReserve", fn: "PresignReserve", nextFn: "GetPendingForCommit", want: "AND tenant_id=$2::uuid"},
		{name: "CommitUpload", fn: "CommitUpload", nextFn: "SyncCurrentRevisionArtifactMetadata", want: "AND tenant_id=$2::uuid"},
		{name: "SyncCurrentRevisionArtifactMetadata", fn: "SyncCurrentRevisionArtifactMetadata", nextFn: "DeleteExpiredPending", want: "tenant_id = $2::uuid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := repositoryFuncBody(t, src, tc.fn, tc.nextFn)
			if !strings.Contains(body, tc.want) {
				t.Fatalf("%s must scope editor_sessions by tenant_id", tc.name)
			}
		})
	}
}

func TestCheckpointAndRevisionQueriesIncludeTenantScope(t *testing.T) {
	src := readRepositorySource(t)

	cases := []struct {
		name   string
		fn     string
		nextFn string
	}{
		{name: "GetPendingForCommit", fn: "GetPendingForCommit", nextFn: "CommitUpload"},
		{name: "CreateCheckpoint", fn: "CreateCheckpoint", nextFn: "ListCheckpoints"},
		{name: "ListCheckpoints", fn: "ListCheckpoints", nextFn: "ListRevisionHistory"},
		{name: "GetRevision", fn: "GetRevision", nextFn: "RestoreCheckpoint"},
		{name: "RestoreCheckpoint", fn: "RestoreCheckpoint", nextFn: "IsDocumentOwner"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := repositoryFuncBody(t, src, tc.fn, tc.nextFn)
			if !strings.Contains(body, "tenant_id") {
				t.Fatalf("%s must include tenant_id scope", tc.name)
			}
			if !strings.Contains(body, "documents") {
				t.Fatalf("%s must derive tenant scope from documents", tc.name)
			}
		})
	}
}

func readRepositorySource(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}
	return string(src)
}

func repositoryFuncBody(t *testing.T, src, fn, nextFn string) string {
	t.Helper()
	start := strings.Index(src, "func (r *Repository) "+fn)
	if start == -1 {
		t.Fatalf("%s not found", fn)
	}
	end := strings.Index(src[start:], "func (r *Repository) "+nextFn)
	if end == -1 {
		t.Fatalf("%s boundary not found", nextFn)
	}
	return src[start : start+end]
}
