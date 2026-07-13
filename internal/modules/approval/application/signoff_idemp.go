package application

// SignoffReplay is the cached outcome envelope of a previously completed signoff
// operation, persisted in idempotency_keys.response_body and replayed verbatim to
// a retrying caller.
type SignoffReplay struct {
	Outcome string `json:"outcome"`
}

// SignoffReplayCommitter is the slot handle a winning caller must resolve with
// exactly one of Complete or Fail. It is an interface so HTTP handlers depend on
// the behaviour, not the Postgres-bound concrete type, which keeps the replay
// seam unit-testable with an in-memory double. Production impl lives in
// infrastructure/postgres_signoff_idemp_store.go.
type SignoffReplayCommitter interface {
	Complete(outcome string) error
	Fail(cause error) error
}
