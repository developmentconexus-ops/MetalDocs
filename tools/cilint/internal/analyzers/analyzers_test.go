package analyzers_test

import (
	"os"
	"path/filepath"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

func writeFixture(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

// ─── TxOwnership ────────────────────────────────────────────────────────────

func TestTxOwnership_Positive(t *testing.T) {
	src := `package foo
import "database/sql"
func Bad(db *sql.DB) {
	tx, _ := db.BeginTx(nil, nil) // violation
	_ = tx
}
`
	path := writeFixture(t, "bad_tx.go", src)
	findings := analyzers.TxOwnership([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for BeginTx outside allowed package")
	}
}

func TestTxOwnership_Negative_AllowedPackage(t *testing.T) {
	src := `package application
import "database/sql"
func Good(db *sql.DB) {
	tx, _ := db.BeginTx(nil, nil)
	_ = tx
}
`
	// Place in allowed package path
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "approval", "application")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "good.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.TxOwnership([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for allowed package, got %d", len(findings))
	}
}

// ─── LegacyVocab ────────────────────────────────────────────────────────────

func TestLegacyVocab_Positive(t *testing.T) {
	src := `package foo
const badName = "finalized" // legacy
`
	path := writeFixture(t, "legacy.go", src)
	findings := analyzers.LegacyVocab([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for 'finalized'")
	}
}

func TestLegacyVocab_Negative_AllowDirective(t *testing.T) {
	src := `package foo
const ok = "finalized" // cilint:allow-legacy historical constant
`
	path := writeFixture(t, "legacy_allowed.go", src)
	findings := analyzers.LegacyVocab([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings with allow directive, got %d", len(findings))
	}
}

// ─── TxOwnership (recalibrated allowlist) ────────────────────────────────────

// Transactions in the presentation layer (delivery/http) must STILL be flagged —
// that is the real architectural smell the rule guards after the allowlist was
// widened to the data/service/platform layers.
func TestTxOwnership_Positive_DeliveryHTTPStillFlagged(t *testing.T) {
	src := `package httpdelivery
import "database/sql"
func (h *Handler) bad(db *sql.DB) {
	tx, _ := db.BeginTx(nil, nil)
	_ = tx
}
`
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "modules", "documents", "delivery", "http")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "handler.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.TxOwnership([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for BeginTx in delivery/http handler")
	}
}

func TestTxOwnership_Negative_InfrastructureLayer(t *testing.T) {
	src := `package postgres
import "database/sql"
func (r *Repository) ok(db *sql.DB) {
	tx, _ := db.BeginTx(nil, nil)
	_ = tx
}
`
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "modules", "auth", "infrastructure", "postgres")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "repository.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.TxOwnership([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for infrastructure layer, got %d", len(findings))
	}
}

// ─── PlatformBoundary (REQ-TOP-2 import direction) ───────────────────────────

func platformBoundaryFixture(t *testing.T, pkgSegments []string, src string) string {
	t.Helper()
	dir := t.TempDir()
	pkgDir := filepath.Join(append([]string{dir, "internal", "platform"}, pkgSegments...)...)
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "code.go")
	_ = os.WriteFile(path, []byte(src), 0o644)
	return path
}

func TestPlatformBoundary_Positive_ModuleImport(t *testing.T) {
	src := `package cache
import (
	"context"
	authdomain "metaldocs/internal/modules/auth/domain"
)
func Bad(ctx context.Context) { _ = authdomain.CurrentUserFromContext; _ = ctx }
`
	path := platformBoundaryFixture(t, []string{"cache"}, src)
	findings := analyzers.PlatformBoundary([]string{path})
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding for module import from platform package, got %d: %+v", len(findings), findings)
	}
}

func TestPlatformBoundary_Negative_PlatformOnlyImports(t *testing.T) {
	src := `package observability
import (
	"net/http"
	"metaldocs/internal/platform/problem"
)
func OK(w http.ResponseWriter) { _ = problem.Write }
`
	path := platformBoundaryFixture(t, []string{"observability"}, src)
	findings := analyzers.PlatformBoundary([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for platform-only imports, got %d: %+v", len(findings), findings)
	}
}

func TestPlatformBoundary_Negative_FrozenBaselinePackage(t *testing.T) {
	src := `package bootstrap
import documents "metaldocs/internal/modules/documents"
var _ = documents.Module{}
`
	path := platformBoundaryFixture(t, []string{"bootstrap"}, src)
	findings := analyzers.PlatformBoundary([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for frozen-baseline package, got %d: %+v", len(findings), findings)
	}
}

func TestPlatformBoundary_Negative_ModuleFilesNotInScope(t *testing.T) {
	src := `package application
import other "metaldocs/internal/modules/iam/domain"
var _ = other.Role("")
`
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "internal", "modules", "documents", "application")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "svc.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.PlatformBoundary([]string{path})
	if len(findings) != 0 {
		t.Fatalf("module-to-module imports are out of scope, got %d findings", len(findings))
	}
}

// ─── OutboxPair (struct-field emitter detection) ─────────────────────────────

func TestOutboxPair_Negative_StructFieldEmitter(t *testing.T) {
	src := `package application
type Svc struct{ emitter Emitter }
func (s *Svc) RecordSignoff(tx interface{}) {
	s.repo.UpdateInstanceStatus(tx)
	s.emitter.Emit(tx, nil)
}
`
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "documents", "approval", "application")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "decision_service.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.OutboxPair([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings when s.emitter.Emit is paired with mutation, got %d: %+v", len(findings), findings)
	}
}

// ─── PostCommitAudit ─────────────────────────────────────────────────────────

// postCommitAuditFixture writes a Go source file inside internal/modules/...
// so PostCommitAudit picks it up.
func postCommitAuditFixture(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	pkgDir := dir + "/internal/modules/foo/application"
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := pkgDir + "/svc.go"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestPostCommitAudit_Positive_LogAfterCommit(t *testing.T) {
	src := `package application
func (s *Svc) Create(ctx interface{}, tx interface{}) error {
	s.repo.Update(tx)
	tx.Commit()
	s.govLogger.Log(ctx, nil) // post-commit — violation
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for Log after Commit")
	}
}

func TestPostCommitAudit_Positive_WriteAfterCommit(t *testing.T) {
	src := `package application
func (s *Svc) Archive(ctx interface{}, tx interface{}) error {
	s.repo.MarkArchived(ctx)
	tx.Commit()
	s.audit.Write(ctx, "t", "a", "doc.archived", "d", nil) // post-commit — violation
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for Write after Commit")
	}
}

func TestPostCommitAudit_Negative_LogBeforeCommit(t *testing.T) {
	src := `package application
func (s *Svc) Create(ctx interface{}, tx interface{}) error {
	s.repo.Update(tx)
	s.govLogger.Log(ctx, nil) // before commit — OK
	tx.Commit()
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for Log before Commit, got %d: %+v", len(findings), findings)
	}
}

func TestPostCommitAudit_Negative_LogTxBeforeCommit(t *testing.T) {
	src := `package application
func (s *Svc) Create(ctx interface{}, tx interface{}) error {
	s.repo.Update(tx)
	s.govLogger.LogTx(ctx, tx, nil) // Tx variant before commit — OK
	tx.Commit()
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings for LogTx before Commit, got %d: %+v", len(findings), findings)
	}
}

func TestPostCommitAudit_Negative_NoCommit(t *testing.T) {
	src := `package application
func (s *Svc) Create(ctx interface{}) error {
	s.repo.Update(ctx)
	s.govLogger.Log(ctx, nil)
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings when there is no Commit, got %d: %+v", len(findings), findings)
	}
}

func TestPostCommitAudit_Negative_AllowDirective(t *testing.T) {
	src := `package application
func (s *Svc) Create(ctx interface{}, tx interface{}) error {
	s.repo.Update(tx)
	tx.Commit()
	s.govLogger.Log(ctx, nil) //cilint:allow-post-commit-audit legacy best-effort path
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings with allow directive, got %d: %+v", len(findings), findings)
	}
}

func TestPostCommitAudit_Negative_OutsideModulesPath(t *testing.T) {
	src := `package platform
func (s *Svc) Create(ctx interface{}, tx interface{}) error {
	s.repo.Update(tx)
	tx.Commit()
	s.govLogger.Log(ctx, nil) // outside modules — not scoped
	return nil
}
`
	dir := t.TempDir()
	pkgDir := dir + "/internal/platform/foo"
	_ = os.MkdirAll(pkgDir, 0o755)
	path := pkgDir + "/svc.go"
	_ = os.WriteFile(path, []byte(src), 0o644)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings outside internal/modules, got %d: %+v", len(findings), findings)
	}
}

func TestPostCommitAudit_Positive_AppendAuditAfterCommit(t *testing.T) {
	src := `package application
func (s *Svc) Submit(ctx interface{}) error {
	tx := s.db.BeginTx(ctx, nil)
	s.repo.UpdateVersionTx(ctx, tx, nil)
	tx.Commit()
	s.repo.AppendAudit(ctx, nil) // post-commit — violation
	return nil
}
`
	path := postCommitAuditFixture(t, src)
	findings := analyzers.PostCommitAudit([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for AppendAudit after Commit")
	}
}

func TestOutboxPair_Positive_MutationWithoutEmit(t *testing.T) {
	src := `package application
type Svc struct{}
func (s *Svc) MutateOnly(tx interface{}) {
	s.repo.UpdateInstanceStatus(tx)
}
`
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "documents", "approval", "application")
	_ = os.MkdirAll(pkgDir, 0o755)
	path := filepath.Join(pkgDir, "x_service.go")
	_ = os.WriteFile(path, []byte(src), 0o644)

	findings := analyzers.OutboxPair([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for mutation without paired Emit")
	}
}
