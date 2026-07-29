package dispatchjobs

// dispatchFields is the payload shape SHARED by both staging dispatch River
// jobs: the tenant/revision identity plus the staging outbox row's own ID,
// which the worker needs to mark the row dispatched or dead-lettered after the
// event is published.
//
// The payload hash is deliberately NOT here. The two outbox tables carry
// DIFFERENT hashes (F-QA4-10, migration 0312) — materialize carries the
// resolved-placeholder values hash, pdf carries the materialized frozen-docx
// hash — so a single shared field could only be named after one of them, or
// after neither. Each Args type declares its own, correctly-named field.
type dispatchFields struct {
	TenantID   string `json:"tenant_id"`
	RevisionID string `json:"revision_id"`
	OutboxID   string `json:"outbox_id"`
	// ReleaseGenerationID is the ADR 0085 release generation this render
	// belongs to. It widens the dedup identity from revision-only to
	// generation-aware: a re-approval of the same revision is a NEW generation
	// and must re-materialize rather than dedup against the previous run.
	ReleaseGenerationID string `json:"release_generation_id,omitempty"`
}

// PDFDispatchArgs is the River job payload for dispatching one pdf staging
// outbox row as a docgen_v2_pdf event.
type PDFDispatchArgs struct {
	dispatchFields
	// FrozenDocxHash is the MATERIALIZED frozen-docx hash produced by the
	// renderer fanout (pdf_dispatch_outbox.frozen_docx_hash, and the same value
	// written to documents.content_hash).
	FrozenDocxHash []byte `json:"frozen_docx_hash"`
	// FinalDocxS3Key is the renderer-produced frozen-docx key, carried from the
	// pdf_dispatch_outbox snapshot into the PDFConvertPayload (F-QA2-2).
	FinalDocxS3Key string `json:"final_docx_s3_key,omitempty"`
}

// Kind implements river.JobArgs, identifying this job type to River.
func (PDFDispatchArgs) Kind() string { return "pdf_dispatch" }

// MaterializeDispatchArgs is the River job payload for dispatching one
// materialize staging outbox row as a docx_materialize event.
type MaterializeDispatchArgs struct {
	dispatchFields
	// ValuesHash is the resolved-placeholder values hash pinned at freeze
	// (materialize_dispatch_outbox.values_hash, and the same value written to
	// documents.values_hash).
	ValuesHash []byte `json:"values_hash"`
}

// Kind implements river.JobArgs, identifying this job type to River.
func (MaterializeDispatchArgs) Kind() string { return "materialize_dispatch" }
