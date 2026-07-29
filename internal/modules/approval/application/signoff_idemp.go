package application

// SignoffReplay is the cached outcome envelope of a previously completed signoff
// operation, persisted in idempotency_keys.response_body and replayed verbatim to
// a retrying caller.
type SignoffReplay struct {
	Outcome string `json:"outcome"`
	// SignoffID is the approval_signoffs row id produced by the original
	// (winning) call, so an Idempotency-Key retry replays the SAME identifier
	// instead of an empty string (F-QA4-7). Envelopes persisted before this
	// field existed decode with an empty value — the pre-fix behaviour — which
	// is why it is omitempty rather than required.
	//
	// The review-verdict route reuses this SAME slot for the
	// approval_review_verdicts row id (it replays through the same envelope and
	// store) rather than introducing a parallel id field.
	SignoffID string `json:"signoff_id,omitempty"`
}

// SignoffReplayCommitter is the slot handle a winning caller must resolve with
// exactly one of Complete or Fail. It is an interface so HTTP handlers depend on
// the behaviour, not the Postgres-bound concrete type, which keeps the replay
// seam unit-testable with an in-memory double. Production impl lives in
// infrastructure/postgres_signoff_idemp_store.go.
type SignoffReplayCommitter interface {
	Complete(replay SignoffReplay) error
	Fail(cause error) error
}
