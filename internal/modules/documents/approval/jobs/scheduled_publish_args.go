package jobs

import "time"

type ScheduledPublishArgs struct {
	TenantID                string    `json:"tenant_id"`
	DocumentID              string    `json:"document_id"`
	ExpectedRevisionVersion int       `json:"expected_revision_version"`
	ScheduledEffectiveAt    time.Time `json:"scheduled_effective_at"`
	ScheduleGeneration      int64     `json:"schedule_generation"`
}

func (ScheduledPublishArgs) Kind() string { return "scheduled_publish_cutover" }
