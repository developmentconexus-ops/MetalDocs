package retention

// PurgeArgs is the (empty) River job payload for the staging-outbox purge
// tick. The job carries no per-run parameters; retention window and batch
// bounds are fixed constants (contract-bound, M5 F5.4).
type PurgeArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (PurgeArgs) Kind() string { return "staging-outbox-purge" }
