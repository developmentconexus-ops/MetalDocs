package application

import (
	"context"
	"database/sql"
	"encoding/json"

	registrydomain "metaldocs/internal/modules/registry/domain"
)

// CDDocumentInitializer adapts the documents Service to the registry-owned
// DocumentInitializer port. It lets the registry atomic-create flow seed
// initial document rows inside the same tx as the CD insert.
type CDDocumentInitializer struct {
	svc *Service
}

func NewCDDocumentInitializer(svc *Service) *CDDocumentInitializer {
	return &CDDocumentInitializer{svc: svc}
}

// CloneTemplate threads the caller's tx into Service.cloneIntoTx so the
// document, initial revision, editor session, snapshot columns and required
// placeholder rows commit atomically with the CD row.
func (i *CDDocumentInitializer) CloneTemplate(ctx context.Context, tx *sql.Tx, cd *registrydomain.ControlledDocument, req registrydomain.CloneTemplateRequest) (*registrydomain.DocumentRef, error) {
	var formData json.RawMessage
	if req.FormData != nil {
		raw, err := json.Marshal(req.FormData)
		if err != nil {
			return nil, err
		}
		formData = raw
	}

	docID, contentHash, err := i.svc.cloneIntoTx(ctx, tx, cloneIntoTxInput{
		TenantID:                  cd.TenantID,
		ControlledDocumentID:      cd.ID,
		ProfileCode:               cd.ProfileCode,
		ProcessAreaCode:           cd.ProcessAreaCode,
		Code:                      cd.Code,
		OverrideTemplateVersionID: req.TemplateVersionID,
		OwnerUserID:               cd.OwnerUserID,
		Name:                      req.Name,
		FormData:                  formData,
	})
	if err != nil {
		return nil, err
	}
	return &registrydomain.DocumentRef{ID: docID, ContentHash: contentHash}, nil
}
