package domain

import (
	"errors"
	"testing"

	templatesdomain "metaldocs/internal/modules/templates/domain"
)

// TestResolve_UsesTemplatesVocabulary is a boundary regression guard: it feeds the templates-owned
// VersionStatus constants directly as Resolve inputs and asserts the resolver agrees with that
// vocabulary. If the templates published/obsolete wire values ever drift from what resolution.go
// compares against, this fails at CD's boundary instead of passing silently.
func TestResolve_UsesTemplatesVocabulary(t *testing.T) {
	published := string(templatesdomain.VersionStatusPublished)
	obsolete := string(templatesdomain.VersionStatusObsolete)

	res, err := Resolve(TemplateResolutionInput{
		ProfileCode:     "po",
		DefaultTemplate: &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &published},
	})
	if err != nil || res.TemplateVersionID != "def-1" || res.Source != "default" {
		t.Fatalf("published default should resolve, got res=%+v err=%v", res, err)
	}

	if _, err := Resolve(TemplateResolutionInput{
		ProfileCode:     "po",
		DefaultTemplate: &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &obsolete},
	}); !errors.Is(err, ErrDefaultObsolete) {
		t.Fatalf("obsolete default should be ErrDefaultObsolete, got %v", err)
	}

	if _, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "po", Status: &obsolete},
	}); !errors.Is(err, ErrOverrideNotPublished) {
		t.Fatalf("non-published override should be ErrOverrideNotPublished, got %v", err)
	}
}

func TestResolve_Override_Wins(t *testing.T) {
	published := "published"
	obsolete := "obsolete"

	res, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "po", Status: &published},
		DefaultTemplate:  &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &obsolete},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TemplateVersionID != "ovr-1" || res.Source != "override" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestResolve_Override_Deleted(t *testing.T) {
	_, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "po", Status: nil},
	})
	if !errors.Is(err, ErrOverrideTemplateDeleted) {
		t.Fatalf("expected ErrOverrideTemplateDeleted, got %v", err)
	}
}

func TestResolve_Override_NotPublished(t *testing.T) {
	draft := "draft"
	_, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "po", Status: &draft},
	})
	if !errors.Is(err, ErrOverrideNotPublished) {
		t.Fatalf("expected ErrOverrideNotPublished, got %v", err)
	}
}

func TestResolve_Override_ProfileMismatch(t *testing.T) {
	published := "published"
	_, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "it", Status: &published},
	})
	if !errors.Is(err, ErrTemplateProfileMismatch) {
		t.Fatalf("expected ErrTemplateProfileMismatch, got %v", err)
	}
}

func TestResolve_Default_Only(t *testing.T) {
	published := "published"
	res, err := Resolve(TemplateResolutionInput{
		ProfileCode:     "po",
		DefaultTemplate: &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &published},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TemplateVersionID != "def-1" || res.Source != "default" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestResolve_BothNull(t *testing.T) {
	_, err := Resolve(TemplateResolutionInput{ProfileCode: "po"})
	if !errors.Is(err, ErrProfileHasNoDefaultTemplate) {
		t.Fatalf("expected ErrProfileHasNoDefaultTemplate, got %v", err)
	}
}

func TestResolve_Default_Obsolete(t *testing.T) {
	obsolete := "obsolete"
	_, err := Resolve(TemplateResolutionInput{
		ProfileCode:     "po",
		DefaultTemplate: &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &obsolete},
	})
	if !errors.Is(err, ErrDefaultObsolete) {
		t.Fatalf("expected ErrDefaultObsolete, got %v", err)
	}
}

func TestResolve_Default_Deleted(t *testing.T) {
	_, err := Resolve(TemplateResolutionInput{
		ProfileCode:     "po",
		DefaultTemplate: &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: nil},
	})
	if !errors.Is(err, ErrProfileHasNoDefaultTemplate) {
		t.Fatalf("expected ErrProfileHasNoDefaultTemplate, got %v", err)
	}
}

func TestResolve_DefaultObsolete_Override_Exists(t *testing.T) {
	published := "published"
	obsolete := "obsolete"
	res, err := Resolve(TemplateResolutionInput{
		ProfileCode:      "po",
		OverrideTemplate: &TemplateVersionCandidate{ID: "ovr-1", ProfileCode: "po", Status: &published},
		DefaultTemplate:  &TemplateVersionCandidate{ID: "def-1", ProfileCode: "po", Status: &obsolete},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TemplateVersionID != "ovr-1" {
		t.Fatalf("expected override id, got %+v", res)
	}
}
