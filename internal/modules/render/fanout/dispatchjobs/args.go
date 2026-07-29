package dispatchjobs

// dispatchFields is the shared payload shape for the staging dispatch River
// jobs. It mirrors fanout.OutboxRow (TenantID, RevisionID, ContentHash) plus
// the staging outbox row's own ID, which the worker needs to mark the row
// dispatched or dead-lettered after the event is published.
type dispatchFields struct {
	TenantID    string `json:"tenant_id"`
	RevisionID  string `json:"revision_id"`
	ContentHash []byte `json:"content_hash"`
	OutboxID    string `json:"outbox_id"`
	// FinalDocxS3Key is the renderer-produced frozen-docx key, carried from the
	// pdf_dispatch_outbox snapshot into the PDFConvertPayload (F-QA2-2). Set only
	// on the pdf dispatch path; empty on the materialize dispatch path (which has
	// no docx key at enqueue time).
	FinalDocxS3Key string `json:"final_docx_s3_key,omitempty"`
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
}

// Kind implements river.JobArgs, identifying this job type to River.
func (PDFDispatchArgs) Kind() string { return "pdf_dispatch" }

// MaterializeDispatchArgs is the River job payload for dispatching one
// materialize staging outbox row as a docx_materialize event.
type MaterializeDispatchArgs struct {
	dispatchFields
}

// Kind implements river.JobArgs, identifying this job type to River.
func (MaterializeDispatchArgs) Kind() string { return "materialize_dispatch" }
