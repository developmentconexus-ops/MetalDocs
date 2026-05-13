package audit_integrity_validator

import (
	"context"
	"errors"
	"testing"

	auditdomain "metaldocs/internal/modules/audit/domain"
)

type fakeValidator struct {
	issues []auditdomain.IntegrityIssue
	err    error
	calls  int
}

func (v *fakeValidator) ValidateIntegrity(context.Context) ([]auditdomain.IntegrityIssue, error) {
	v.calls++
	return v.issues, v.err
}

func TestJobReturnsNilWhenChainIsValid(t *testing.T) {
	t.Parallel()

	validator := &fakeValidator{}

	err := New(validator)(context.Background(), 7)
	if err != nil {
		t.Fatalf("job returned error: %v", err)
	}
	if validator.calls != 1 {
		t.Fatalf("ValidateIntegrity calls = %d, want 1", validator.calls)
	}
}

func TestJobReturnsIntegrityViolationWhenIssuesExist(t *testing.T) {
	t.Parallel()

	validator := &fakeValidator{
		issues: []auditdomain.IntegrityIssue{
			{
				Sequence: 42,
				EventID:  "event-42",
				Kind:     auditdomain.IntegrityIssueRowHashMismatch,
			},
		},
	}

	err := New(validator)(context.Background(), 8)
	if !errors.Is(err, ErrIntegrityViolation) {
		t.Fatalf("err = %v, want ErrIntegrityViolation", err)
	}
}

func TestJobPropagatesValidatorError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("scan failed")
	validator := &fakeValidator{err: expectedErr}

	err := New(validator)(context.Background(), 9)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("err = %v, want %v", err, expectedErr)
	}
}
