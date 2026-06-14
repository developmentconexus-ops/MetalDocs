package wiring

import (
	"context"
	"encoding/json"
	"fmt"

	controlleddocumentsapp "metaldocs/internal/modules/controlleddocuments/application"
	docapp "metaldocs/internal/modules/documents/application"
	taxonomydomain "metaldocs/internal/modules/taxonomy/domain"
)

// controlledDocumentDuplicatorService is the narrow interface of *ControlledDocumentService
// the duplicator adapter needs; the interface seam keeps the adapter unit-testable.
type controlledDocumentDuplicatorService interface {
	Get(ctx context.Context, tenantID, id string) (*controlleddocumentsapp.ControlledDocument, error)
	Create(ctx context.Context, cmd controlleddocumentsapp.CreateControlledDocumentCmd) (*controlleddocumentsapp.CreateResult, error)
}

// controlledDocumentDuplicatorAdapter bridges *controlleddocumentsapp.ControlledDocumentService
// to docapp.ControlledDocumentDuplicator.
type controlledDocumentDuplicatorAdapter struct {
	svc controlledDocumentDuplicatorService
}

// NewControlledDocumentDuplicator returns a docapp.ControlledDocumentDuplicator backed
// by the given ControlledDocumentService. Panics if svc is nil.
func NewControlledDocumentDuplicator(svc *controlleddocumentsapp.ControlledDocumentService) docapp.ControlledDocumentDuplicator {
	if svc == nil {
		panic("controlled document duplicator service is nil")
	}
	return &controlledDocumentDuplicatorAdapter{svc: svc}
}

func (a controlledDocumentDuplicatorAdapter) DuplicateControlledDocument(ctx context.Context, in docapp.DuplicateControlledDocumentInput) (*docapp.CreateDocumentResult, error) {
	if a.svc == nil {
		return nil, fmt.Errorf("duplicate controlled document %s: service not configured", in.ControlledDocumentID)
	}
	source, err := a.svc.Get(ctx, in.TenantID, in.ControlledDocumentID)
	if err != nil {
		return nil, fmt.Errorf("duplicate controlled document %s: load source: %w", in.ControlledDocumentID, err)
	}
	var overrideReason *string
	if source.OverrideTemplateVersionID != nil {
		reason := "Duplicated from existing controlled document"
		overrideReason = &reason
	}
	var formData map[string]any
	if len(in.FormData) > 0 {
		if err := json.Unmarshal(in.FormData, &formData); err != nil {
			return nil, fmt.Errorf("duplicate controlled document %s: unmarshal form data: %w", in.ControlledDocumentID, err)
		}
	}
	res, err := a.svc.Create(ctx, controlleddocumentsapp.CreateControlledDocumentCmd{
		TenantID:                  in.TenantID,
		ProfileCode:               source.ProfileCode,
		ProcessAreaCode:           source.ProcessAreaCode,
		DepartmentCode:            source.DepartmentCode,
		Title:                     source.Title,
		OwnerUserID:               source.OwnerUserID,
		ActorUserID:               in.ActorUserID,
		OverrideTemplateVersionID: source.OverrideTemplateVersionID,
		OverrideTemplateReason:    overrideReason,
		DocumentName:              in.DocumentName,
		FormData:                  formData,
		VisibilityScope:           string(source.Visibility.Scope),
		VisibilityAreaCodes:       source.Visibility.AreaCodes,
		VisibilityUserIDs:         source.Visibility.UserIDs,
		// TemplateVersionID threads the source's override into the off-tx clone
		// resolver so the duplicate clones the source's override version (not the
		// profile default); OverrideTemplateVersionID above drives in-tx override
		// validation + persistence. Both are nil when the source has no override,
		// so the duplicate falls back to the profile's current default template.
		TemplateVersionID: source.OverrideTemplateVersionID,
	})
	if err != nil {
		return nil, fmt.Errorf("duplicate controlled document %s: create duplicate: %w", in.ControlledDocumentID, err)
	}
	if res.DocumentRef == nil {
		return nil, fmt.Errorf("duplicate controlled document: create returned no document ref")
	}
	return &docapp.CreateDocumentResult{
		DocumentID:        res.DocumentRef.ID,
		InitialRevisionID: res.DocumentRef.RevisionID,
		SessionID:         res.DocumentRef.SessionID,
	}, nil
}

// profileDefaultsAdapter bridges a taxonomy ProfileRepository to
// docapp.ProfileDefaultTemplateReader.
type profileDefaultsAdapter struct {
	profileRepo interface {
		GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
	}
}

// NewProfileDefaults returns a docapp.ProfileDefaultTemplateReader backed by the given
// profile repository.
func NewProfileDefaults(profileRepo interface {
	GetByCode(ctx context.Context, tenantID string, code taxonomydomain.ProfileCode) (*taxonomydomain.DocumentProfile, error)
}) docapp.ProfileDefaultTemplateReader {
	return &profileDefaultsAdapter{profileRepo: profileRepo}
}

func (a *profileDefaultsAdapter) GetDefaultTemplateVersionID(ctx context.Context, tenantID, profileCode string) (*string, *string, error) {
	profile, err := a.profileRepo.GetByCode(ctx, tenantID, taxonomydomain.ProfileCode(profileCode))
	if err != nil {
		return nil, nil, err
	}
	if profile.DefaultTemplateVersionID == nil {
		return nil, nil, nil
	}
	status := "published"
	return profile.DefaultTemplateVersionID, &status, nil
}
