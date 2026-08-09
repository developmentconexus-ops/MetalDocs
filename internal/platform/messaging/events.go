package messaging

import "context"

// EventID uniquely identifies an outbox event.
type EventID string

// EventType discriminates the payload shape carried by an outbox event.
type EventType string

// AggregateType identifies the domain aggregate kind an event is about.
type AggregateType string

// AggregateID identifies the specific domain aggregate instance an event is about.
type AggregateID string

// IdempotencyKey deduplicates event publication.
type IdempotencyKey string

// TraceID carries the originating request's trace identifier for correlation.
type TraceID string

// EventTypePDFConvert requests DOCX->PDF conversion for a revision.
const EventTypePDFConvert EventType = "docgen_v2_pdf"

// EventTypeMaterializeFanout triggers async DOCX materialization (ADR 0015).
const EventTypeMaterializeFanout EventType = "docx_materialize"

// Payload is implemented by every concrete outbox event payload type.
type Payload interface {
	eventPayload()
}

// PDFConvertPayload is the payload for EventTypePDFConvert.
type PDFConvertPayload struct {
	TenantID       string `json:"tenant_id"`
	RevisionID     string `json:"revision_id"`
	FinalDocxS3Key string `json:"final_docx_s3_key"`
	ContentHash    string `json:"content_hash,omitempty"`
	// ReleaseGenerationID keys the produced artifact to one ADR 0085 release
	// generation. Empty on legacy/non-approval renders, which produce no
	// artifact fact.
	ReleaseGenerationID string `json:"release_generation_id,omitempty"`
}

func (PDFConvertPayload) eventPayload() {}

// MaterializeFanoutPayload is the payload for EventTypeMaterializeFanout.
type MaterializeFanoutPayload struct {
	TenantID   string `json:"tenant_id"`
	RevisionID string `json:"revision_id"`
	// ReleaseGenerationID keys the produced artifact to one ADR 0085 release
	// generation. Empty on legacy/non-approval renders.
	ReleaseGenerationID string `json:"release_generation_id,omitempty"`
}

func (MaterializeFanoutPayload) eventPayload() {}

// UnknownPayload is the fallback payload for event types with no registered
// concrete decoder, preserved as raw key/value pairs.
type UnknownPayload map[string]any

func (UnknownPayload) eventPayload() {}

// Event is the stable envelope used by internal domain event publishers.
type Event struct {
	EventID           EventID
	EventType         EventType
	AggregateType     AggregateType
	AggregateID       AggregateID
	OccurredAtRFC3339 string
	Version           int
	AttemptCount      int
	IdempotencyKey    IdempotencyKey
	Producer          string
	TraceID           TraceID
	Payload           Payload
}

// Publisher abstracts internal event delivery.
type Publisher interface {
	Publish(ctx context.Context, event Event) error
}
