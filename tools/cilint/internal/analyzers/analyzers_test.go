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
