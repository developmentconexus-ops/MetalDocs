// Package maintenance holds the shared River periodic-job definitions for the
// maintenance-queue jobs (stuck-instance watchdog, idempotency janitor, audit
// integrity validator, the M6 F6.2 document review-due surfacer, the F8
// approval SLA surfacer, the ADR 0085 release-hold reconciler, and the
// outbox_events terminal-row retention purge). Both
// metaldocs-api and metaldocs-jobs define these periodic jobs on their River
// client config so that whichever process wins leader election is the one
// that enqueues them (River only enqueues periodic jobs from the elected
// leader's own Config.PeriodicJobs). Only metaldocs-jobs subscribes the
// "maintenance" queue and registers the actual Workers — this package holds
// ONLY the schedule/args wiring, no worker construction, no DB (ADR 0067).
package maintenance

import (
	"time"

	"github.com/riverqueue/river"

	"metaldocs/internal/modules/jobs/approval_sla_surfacer"
	"metaldocs/internal/modules/jobs/audit_integrity_validator"
	"metaldocs/internal/modules/jobs/document_review_surfacer"
	"metaldocs/internal/modules/jobs/idempotency_janitor"
	"metaldocs/internal/modules/jobs/outbox_retention"
	"metaldocs/internal/modules/jobs/release_hold_reconciler"
	"metaldocs/internal/modules/jobs/stuck_instance_watchdog"
)

// PeriodicJobs returns the River periodic-job definitions for the 3 janitors
// plus the M6 F6.2 document review-due surfacer, the F8 approval SLA surfacer,
// the ADR 0085 release-hold reconciler and the outbox_events terminal-row
// retention purge. It must be included in the
// Config.PeriodicJobs of every River client that
// participates in leader election for these jobs (currently metaldocs-api and
// metaldocs-jobs), regardless of whether that client subscribes the
// "maintenance" queue or registers the corresponding Workers.
func PeriodicJobs() []*river.PeriodicJob {
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(5*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return stuck_instance_watchdog.StuckInstanceWatchdogArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "stuck-instance-watchdog", RunOnStart: false},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(15*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return idempotency_janitor.IdempotencyJanitorArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "idempotency-janitor", RunOnStart: false},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return audit_integrity_validator.AuditIntegrityValidatorArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "audit-integrity-validator", RunOnStart: false},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return document_review_surfacer.DocumentReviewSurfacerArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "document-review-surfacer", RunOnStart: false},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return approval_sla_surfacer.ApprovalSLASurfacerArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "approval-sla-surfacer", RunOnStart: false},
		),
		// ADR 0085 Stage C W2 reconciliation sweep. 15 minutes against a
		// 30-minute stuck threshold: a hold that crosses the threshold is
		// surfaced within one tick of crossing it, and no generation can cross
		// the threshold and be released again between two ticks unobserved.
		river.NewPeriodicJob(
			river.PeriodicInterval(15*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return release_hold_reconciler.ReleaseHoldReconcilerArgs{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: "release-hold-reconciler", RunOnStart: false},
		),
		// Terminal-row retention for metaldocs.outbox_events. Daily, matching
		// the staging-outbox purge cadence (both stages of the two-stage
		// outbox chain age out on the same clock) and far finer than either
		// retention window, so a tick missed to a restart costs nothing.
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return outbox_retention.Args{}, &river.InsertOpts{Queue: "maintenance"}
			},
			&river.PeriodicJobOpts{ID: outbox_retention.JobName, RunOnStart: false},
		),
	}
}
