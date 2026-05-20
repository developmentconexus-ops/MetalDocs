package contracts

import (
	"fmt"
	"time"
)

type PublishRequest struct {
	IdempotencyKey string
	IfMatchVersion int
}

type SchedulePublishRequest struct {
	EffectiveFrom        string `json:"effective_from"`
	SupersededDocumentID string `json:"superseded_document_id,omitempty"`
	IdempotencyKey       string
	IfMatchVersion       int
}

func (r SchedulePublishRequest) Validate() error {
	if err := validateRequired("effective_from", r.EffectiveFrom); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, r.EffectiveFrom)
	if err != nil {
		return fmt.Errorf("effective_from must be parseable RFC3339: %w", err)
	}
	_, offset := t.Zone()
	if offset != 0 {
		return fmt.Errorf("effective_from must be UTC")
	}
	if r.SupersededDocumentID != "" {
		if err := validateUUID("superseded_document_id", r.SupersededDocumentID); err != nil {
			return err
		}
	}
	return nil
}

type PublishResponse struct {
	DocumentID    string `json:"document_id"`
	NewStatus     string `json:"new_status"`
	EffectiveFrom string `json:"effective_from,omitempty"`
}
