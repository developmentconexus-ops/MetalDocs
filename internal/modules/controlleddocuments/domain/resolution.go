package domain

import "errors"

// tplVersionPublished and tplVersionObsolete are the wire-string values that
// controlled-documents compares template version statuses against. They mirror
// templatesdomain.VersionStatusPublished / VersionStatusObsolete without
// importing the sibling BC's domain package (BC domain independence rule).
// A parity test in resolution_test.go asserts they equal the templates-owned
// constants so a future rename is caught at the boundary.
const (
	tplVersionPublished = "published"
	tplVersionObsolete  = "obsolete"
)

// TemplateResolutionInput is the input to Resolve: the target profile plus
// the (already-fetched) override and/or default template version
// candidates. OverrideTemplate, when non-nil, takes precedence over
// DefaultTemplate.
type TemplateResolutionInput struct {
	ProfileCode      string
	OverrideTemplate *TemplateVersionCandidate
	DefaultTemplate  *TemplateVersionCandidate
}

// TemplateVersionCandidate is the minimal template-version projection
// Resolve needs to decide eligibility. Status is nil when the version row
// no longer exists (deleted).
type TemplateVersionCandidate struct {
	ID          string
	ProfileCode string
	Status      *string
}

// TemplateResolutionResult is the winning template version and where it
// came from ("override" or "default").
type TemplateResolutionResult struct {
	TemplateVersionID string
	Source            string
}

// Sentinel errors returned by Resolve, one per rejected-candidate reason.
var (
	ErrProfileHasNoDefaultTemplate = errors.New("profile has no default template")
	ErrOverrideTemplateDeleted     = errors.New("override template deleted")
	ErrOverrideNotPublished        = errors.New("override template is not published")
	ErrDefaultObsolete             = errors.New("default template is obsolete")
	ErrTemplateProfileMismatch     = errors.New("template profile mismatch")
)

// Resolve picks the effective template version for a controlled document:
// in.OverrideTemplate when set (must be published and belong to
// in.ProfileCode), otherwise in.DefaultTemplate (must be published; an
// obsolete default is a distinct error from a missing one). Pure function,
// no I/O.
func Resolve(in TemplateResolutionInput) (TemplateResolutionResult, error) {
	if in.OverrideTemplate != nil {
		return resolveOverrideTemplate(in.ProfileCode, in.OverrideTemplate)
	}

	return resolveDefaultTemplate(in.DefaultTemplate)
}

func resolveOverrideTemplate(profileCode string, candidate *TemplateVersionCandidate) (TemplateResolutionResult, error) {
	if candidate.Status == nil {
		return TemplateResolutionResult{}, ErrOverrideTemplateDeleted
	}
	if *candidate.Status != tplVersionPublished {
		return TemplateResolutionResult{}, ErrOverrideNotPublished
	}
	if candidate.ProfileCode != profileCode {
		return TemplateResolutionResult{}, ErrTemplateProfileMismatch
	}
	return TemplateResolutionResult{TemplateVersionID: candidate.ID, Source: "override"}, nil
}

func resolveDefaultTemplate(candidate *TemplateVersionCandidate) (TemplateResolutionResult, error) {
	if candidate == nil || candidate.Status == nil {
		return TemplateResolutionResult{}, ErrProfileHasNoDefaultTemplate
	}
	if *candidate.Status == tplVersionObsolete {
		return TemplateResolutionResult{}, ErrDefaultObsolete
	}
	if *candidate.Status != tplVersionPublished {
		return TemplateResolutionResult{}, ErrProfileHasNoDefaultTemplate
	}
	return TemplateResolutionResult{TemplateVersionID: candidate.ID, Source: "default"}, nil
}
