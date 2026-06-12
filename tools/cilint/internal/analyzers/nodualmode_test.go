package analyzers_test

import (
	"os"
	"path/filepath"
	"testing"

	"metaldocs/tools/cilint/internal/analyzers"
)

// noDualModeFixture writes a Go source file at the given relative path inside a
// temp directory so path-based filtering works correctly.
func noDualModeFixture(t *testing.T, relPath, src string) string {
	t.Helper()
	dir := t.TempDir()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return full
}

// ─── NoDualMode ──────────────────────────────────────────────────────────────

func TestNoDualMode_Positive_EqNilWithElse(t *testing.T) {
	src := `package application
func (s *Svc) Handle() interface{} {
	if s.db == nil {
		return "no-db result"
	} else {
		return "with-db result"
	}
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for db == nil / else dual-mode branch")
	}
}

func TestNoDualMode_Positive_NeqNilSkipAWrite(t *testing.T) {
	src := `package application
func (s *Svc) Handle() {
	if s.db != nil {
		s.db.Write()
	}
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for db != nil skip-a-write (no else)")
	}
}

// MANDATORY NEGATIVE CASE: the live auth/application/service.go pattern.
// if loginCtxPort == nil { panic("auth.NewService: loginCtxPort is required") }
// This is a required-dep fail-loud guard — must NOT be flagged.
func TestNoDualMode_Negative_PanicGuard_LoginCtxPort(t *testing.T) {
	src := `package application
func NewService(loginCtxPort interface{}) {
	if loginCtxPort == nil {
		panic("auth.NewService: loginCtxPort is required")
	}
}
`
	path := noDualModeFixture(t, "internal/modules/auth/application/service.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("MUST NOT flag required-dep panic guard (loginCtxPort pattern), got %d findings: %+v", len(findings), findings)
	}
}

func TestNoDualMode_Negative_ReturnGuard(t *testing.T) {
	src := `package application
import "errors"
func NewService(db interface{}) error {
	if db == nil {
		return errors.New("db is required")
	}
	return nil
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("MUST NOT flag required-dep return guard, got %d findings: %+v", len(findings), findings)
	}
}

func TestNoDualMode_Negative_ErrCheck(t *testing.T) {
	src := `package application
func (s *Svc) Handle() error {
	err := s.repo.Save()
	if err == nil {
		s.notify()
	}
	return err
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("err identifier should not match db/port heuristic, got %d findings: %+v", len(findings), findings)
	}
}

func TestNoDualMode_OutOfScope_TestFile(t *testing.T) {
	src := `package application
func TestDualMode(t interface{}) {
	var db interface{}
	if db == nil {
		return
	} else {
		_ = db
	}
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc_test.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("_test.go files are out of scope, got %d findings: %+v", len(findings), findings)
	}
}

func TestNoDualMode_OutOfScope_DomainLayer(t *testing.T) {
	src := `package domain
func Check(db interface{}) {
	if db == nil {
		return
	} else {
		_ = db
	}
}
`
	path := noDualModeFixture(t, "internal/modules/foo/domain/check.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("domain layer is out of scope for nodualmode, got %d findings: %+v", len(findings), findings)
	}
}

func TestNoDualMode_Negative_AllowDirective(t *testing.T) {
	src := `package application
func (s *Svc) Handle() interface{} {
	if s.db == nil { //cilint:allow-dualmode legacy migration path
		return "legacy"
	} else {
		return "normal"
	}
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/svc.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) != 0 {
		t.Fatalf("expected no findings with allow directive, got %d: %+v", len(findings), findings)
	}
}

func TestNoDualMode_Positive_SelectorDb(t *testing.T) {
	src := `package application
func (h *Handler) Handle() interface{} {
	if h.sqldb != nil {
		return h.sqldb.Query()
	}
	return nil
}
`
	path := noDualModeFixture(t, "internal/modules/foo/application/handler.go", src)
	findings := analyzers.NoDualMode([]string{path})
	if len(findings) == 0 {
		t.Fatal("expected finding for h.sqldb != nil skip-a-write")
	}
}
