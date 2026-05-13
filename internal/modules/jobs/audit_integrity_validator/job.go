package audit_integrity_validator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	auditdomain "metaldocs/internal/modules/audit/domain"
	"metaldocs/internal/modules/jobs/scheduler"
)

const JobName = "audit_integrity_validator"

var ErrIntegrityViolation = errors.New("audit integrity violation")

func New(validator auditdomain.IntegrityValidator) scheduler.JobFunc {
	return func(ctx context.Context, epoch int64) error {
		issues, err := validator.ValidateIntegrity(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "audit_integrity_validator: validation failed",
				"job", JobName, "epoch", epoch, "error", err)
			return err
		}
		if len(issues) > 0 {
			first := issues[0]
			slog.ErrorContext(ctx, "audit_integrity_validator: integrity violation detected",
				"job", JobName,
				"epoch", epoch,
				"issue_count", len(issues),
				"first_sequence", first.Sequence,
				"first_event_id", first.EventID,
				"first_kind", first.Kind)
			return fmt.Errorf("%w: %d issue(s)", ErrIntegrityViolation, len(issues))
		}

		slog.InfoContext(ctx, "audit_integrity_validator: tick complete",
			"job", JobName, "epoch", epoch, "issue_count", 0)
		return nil
	}
}
