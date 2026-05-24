package fanout

import (
	"context"

	"github.com/google/uuid"

	"metaldocs/internal/platform/messaging"
)

// DispatchInput carries the payload for the docgen_v2_pdf job.
type DispatchInput struct {
	TenantID       string
	RevisionID     string
	FinalDocxS3Key string
}

// PDFDispatcher enqueues docgen_v2_pdf jobs onto the platform event bus.
// Dispatch is called AFTER the approval transaction commits — failures are
// best-effort and never roll back the freeze.
type PDFDispatcher struct {
	pub messaging.Publisher
}

func NewPDFDispatcher(pub messaging.Publisher) *PDFDispatcher {
	return &PDFDispatcher{pub: pub}
}

func (d *PDFDispatcher) Dispatch(ctx context.Context, in DispatchInput) error {
	return d.pub.Publish(ctx, messaging.Event{
		EventID:        messaging.EventID(uuid.NewString()),
		EventType:      messaging.EventTypePDFConvert,
		AggregateType:  messaging.AggregateType("document_revision"),
		AggregateID:    messaging.AggregateID(in.RevisionID),
		IdempotencyKey: messaging.IdempotencyKey("docgen_v2_pdf:" + in.RevisionID),
		Payload: messaging.PDFConvertPayload{
			TenantID:       in.TenantID,
			RevisionID:     in.RevisionID,
			FinalDocxS3Key: in.FinalDocxS3Key,
		},
	})
}
