package messaging

import "context"

type EventID string
type EventType string
type AggregateType string
type AggregateID string
type IdempotencyKey string
type TraceID string

const EventTypePDFConvert EventType = "docgen_v2_pdf"

// EventTypeMaterializeFanout triggers async DOCX materialization (ADR 0015).
const EventTypeMaterializeFanout EventType = "docx_materialize"

type Payload interface {
	eventPayload()
}

type PDFConvertPayload struct {
	TenantID       string `json:"tenant_id"`
	RevisionID     string `json:"revision_id"`
	FinalDocxS3Key string `json:"final_docx_s3_key,omitempty"`
	ContentHash    string `json:"content_hash,omitempty"`
}

func (PDFConvertPayload) eventPayload() {}

type MaterializeFanoutPayload struct {
	TenantID   string `json:"tenant_id"`
	RevisionID string `json:"revision_id"`
}

func (MaterializeFanoutPayload) eventPayload() {}

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
