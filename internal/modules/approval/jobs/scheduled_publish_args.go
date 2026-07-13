package jobs

import "time"

// ScheduledPublishArgs is the River job payload for the deferred publish cutover.
type ScheduledPublishArgs struct {
	TenantID                string    `json:"tenant_id"`
	DocumentID              string    `json:"document_id"`
	ExpectedRevisionVersion int       `json:"expected_revision_version"`
	ScheduledEffectiveAt    time.Time `json:"scheduled_effective_at"`
	ScheduleGeneration      int64     `json:"schedule_generation"`
}

// Kind implements river.JobArgs, identifying this job type to River.
func (ScheduledPublishArgs) Kind() string { return "scheduled_publish_cutover" }
