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
