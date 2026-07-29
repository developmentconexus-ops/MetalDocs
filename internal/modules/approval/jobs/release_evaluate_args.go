package jobs

// ReleaseEvaluateArgs is the River payload for one release-coordinator
// evaluation (ADR 0085).
//
// The payload IS the release generation key — the same identity the facts,
// the projection and the emitted events carry. Nothing else is needed: the
// coordinator re-reads every input under lock, so a job that sat in the queue
// for a week decides on current state, not on state captured at enqueue time.
//
// One job kind serves both triggers:
//   - immediate re-evaluation after a fact lands (ScheduledAt = now);
//   - the effective-date timer (ScheduledAt = planned_effective_from).
//
// Evaluate is idempotent, so duplicate deliveries are free.
type ReleaseEvaluateArgs struct {
	TenantID           string `json:"tenant_id"`
	SubjectKind        string `json:"subject_kind"`
	DocumentID         string `json:"document_id"`
	ApprovalInstanceID string `json:"approval_instance_id"`
	RevisionID         string `json:"revision_id"`
	RevisionVersion    int    `json:"revision_version"`
	FrozenContentHash  string `json:"frozen_content_hash"`
}

// Kind implements river.JobArgs, identifying this job type to River.
func (ReleaseEvaluateArgs) Kind() string { return "release_evaluate" }
