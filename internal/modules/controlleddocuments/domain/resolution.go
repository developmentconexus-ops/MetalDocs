package domain

import "errors"

type TemplateResolutionInput struct {
	ProfileCode      string
	OverrideTemplate *TemplateVersionCandidate
	DefaultTemplate  *TemplateVersionCandidate
}

type TemplateVersionCandidate struct {
	ID          string
	ProfileCode string
	Status      *string
}

type TemplateResolutionResult struct {
	TemplateVersionID string
	Source            string
}

var (
	ErrProfileHasNoDefaultTemplate = errors.New("profile has no default template")
	ErrOverrideTemplateDeleted     = errors.New("override template deleted")
	ErrOverrideNotPublished        = errors.New("override template is not published")
	ErrDefaultObsolete             = errors.New("default template is obsolete")
	ErrTemplateProfileMismatch     = errors.New("template profile mismatch")
)

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
	if *candidate.Status != "published" {
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
	if *candidate.Status == "obsolete" {
		return TemplateResolutionResult{}, ErrDefaultObsolete
	}
	if *candidate.Status != "published" {
		return TemplateResolutionResult{}, ErrProfileHasNoDefaultTemplate
	}
	return TemplateResolutionResult{TemplateVersionID: candidate.ID, Source: "default"}, nil
}
