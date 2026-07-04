// Package retention holds the River periodic-job definition and worker for
// purging aged, already-dispatched rows out of the pdf/materialize staging
// outbox tables (internal/modules/render/fanout.StagingOutboxRepository).
// Dead-lettered rows are never purged (contract §4.1) — the purge query's
// status='dispatched' filter excludes them by construction.
//
// Mirrors internal/modules/jobs/maintenance/periodic.go's convention: both
// metaldocs-api and metaldocs-jobs include PeriodicJob() in their River
// client Config.PeriodicJobs so that whichever process wins leader election
// enqueues the purge tick. Only metaldocs-jobs subscribes the "maintenance"
// queue and registers the actual Worker — this file holds ONLY the
// schedule/args wiring, no worker construction, no DB (ADR 0067).
package retention

import (
	"time"

	"github.com/riverqueue/river"
)

// PeriodicJob returns the River periodic-job definition for the staging
// outbox purge tick. It must be included in the Config.PeriodicJobs of every
// River client that participates in leader election for this job (currently
// metaldocs-api and metaldocs-jobs), regardless of whether that client
// subscribes the "maintenance" queue or registers the PurgeWorker.
func PeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(24*time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return PurgeArgs{}, &river.InsertOpts{Queue: "maintenance"}
		},
		&river.PeriodicJobOpts{ID: "staging-outbox-purge", RunOnStart: false},
	)
}
