package audit_integrity_validator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/riverqueue/river"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/jobs/scheduler"
)

const JobName = "audit_integrity_validator"

var ErrIntegrityViolation = errors.New("audit integrity violation")

// AuditIntegrityValidatorArgs is the (empty) River job payload for the audit
// integrity validation tick. The job carries no per-run parameters.
type AuditIntegrityValidatorArgs struct{}

// Kind implements river.JobArgs, identifying this job type to River.
func (AuditIntegrityValidatorArgs) Kind() string { return JobName }

// AuditIntegrityValidatorWorker is the River worker that runs the audit
// integrity validation tick.
type AuditIntegrityValidatorWorker struct {
	river.WorkerDefaults[AuditIntegrityValidatorArgs]

	validator auditdomain.IntegrityValidator
}

// NewWorker constructs an AuditIntegrityValidatorWorker.
func NewWorker(validator auditdomain.IntegrityValidator) *AuditIntegrityValidatorWorker {
	return &AuditIntegrityValidatorWorker{validator: validator}
}

// Work runs one audit-integrity-validation tick.
func (w *AuditIntegrityValidatorWorker) Work(ctx context.Context, job *river.Job[AuditIntegrityValidatorArgs]) error {
	return run(ctx, w.validator)
}

func New(validator auditdomain.IntegrityValidator) scheduler.JobFunc {
	return func(ctx context.Context, epoch int64) error {
		return run(ctx, validator)
	}
}

// run executes one audit-integrity-validation tick. Extracted so both the
// legacy lease-scheduler New closure and the River Worker delegate to the
// same behavior.
func run(ctx context.Context, validator auditdomain.IntegrityValidator) error {
	issues, err := validator.ValidateIntegrity(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "audit_integrity_validator: validation failed",
			"job", JobName, "error", err)
		return err
	}
	if len(issues) > 0 {
		first := issues[0]
		slog.ErrorContext(ctx, "audit_integrity_validator: integrity violation detected",
			"job", JobName,
			"issue_count", len(issues),
			"first_sequence", first.Sequence,
			"first_event_id", first.EventID,
			"first_kind", first.Kind)
		return fmt.Errorf("%w: %d issue(s)", ErrIntegrityViolation, len(issues))
	}

	slog.InfoContext(ctx, "audit_integrity_validator: tick complete",
		"job", JobName, "issue_count", 0)
	return nil
}
