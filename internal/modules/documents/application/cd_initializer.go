package application

import (
	"context"
	"encoding/json"
	"errors"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	documentsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// CDDocumentInitializer adapts the documents Service to the controlled-document
// DocumentInitializer port. It lets the controlled-document atomic-create flow seed
// initial document rows inside the same tx as the CD insert.
type CDDocumentInitializer struct {
	svc *Service
}

func NewCDDocumentInitializer(svc *Service) *CDDocumentInitializer {
	return &CDDocumentInitializer{svc: svc}
}

func (i *CDDocumentInitializer) ResolveTemplateStorageKey(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	if i == nil || i.svc == nil {
		return "", errors.New("documents service not configured")
	}
	return i.svc.resolveTemplateStorageKey(ctx, tenantID, profileCode, templateVersionID)
}

func (i *CDDocumentInitializer) Exists(ctx context.Context, storageKey string) (bool, error) {
	if i == nil || i.svc == nil {
		return false, errors.New("documents service not configured")
	}
	return i.svc.templateArtifactExists(ctx, storageKey)
}

// ResolveTemplateVersionID resolves the effective template version id off-tx so
// the controlled-document atomic flow can pre-resolve before opening its tx.
func (i *CDDocumentInitializer) ResolveTemplateVersionID(ctx context.Context, tenantID, profileCode string, templateVersionID *string) (string, error) {
	if i == nil || i.svc == nil {
		return "", errors.New("documents service not configured")
	}
	return i.svc.resolveTemplateVersionID(ctx, tenantID, profileCode, templateVersionID)
}

// ResolveDictionaryValues resolves PHDictionary placeholder values off-tx for the
// controlled-document create/revision flow, translating the documents-domain
// not-found sentinel into the controlleddocuments-domain one the CD HTTP layer
// maps to 422 (SP-2 D7).
func (i *CDDocumentInitializer) ResolveDictionaryValues(ctx context.Context, tenantID, templateVersionID string) (map[string]string, error) {
	if i == nil || i.svc == nil {
		return nil, errors.New("documents service not configured")
	}
	vals, err := i.svc.ResolveDictionaryValues(ctx, tenantID, templateVersionID)
	if err != nil {
		if errors.Is(err, documentsdomain.ErrDictionaryTokenMissing) {
			return nil, controlleddocumentsdomain.ErrDictionaryTokenMissing
		}
		return nil, err
	}
	return vals, nil
}

// CloneTemplate threads the caller's tx into Service.cloneIntoTx so the
// document, initial revision, editor session, snapshot columns and required
// placeholder rows commit atomically with the CD row.
func (i *CDDocumentInitializer) CloneTemplate(ctx context.Context, tx db.Tx, cd *controlleddocumentsdomain.ControlledDocument, req controlleddocumentsdomain.CloneTemplateRequest) (*controlleddocumentsdomain.DocumentRef, error) {
	var formData json.RawMessage
	if req.FormData() != nil {
		raw, err := json.Marshal(req.FormData())
		if err != nil {
			return nil, err
		}
		formData = raw
	}

	docID, revID, sessionID, contentHash, err := i.svc.cloneIntoTx(ctx, tx, cloneIntoTxInput{
		TenantID:                  cd.TenantID,
		ControlledDocumentID:      cd.ID,
		ProfileCode:               cd.ProfileCode,
		ProcessAreaCode:           cd.ProcessAreaCode,
		Code:                      cd.Code,
		OverrideTemplateVersionID: req.TemplateVersionID(),
		OwnerUserID:               cd.OwnerUserID,
		Name:                      req.Name(),
		FormData:                  formData,
		DictionaryValues:          req.DictionaryValues(),
	})
	if err != nil {
		return nil, err
	}
	return &controlleddocumentsdomain.DocumentRef{ID: docID, ContentHash: contentHash, RevisionID: revID, SessionID: sessionID}, nil
}
