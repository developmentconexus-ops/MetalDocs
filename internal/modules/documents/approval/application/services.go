package application

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"metaldocs/internal/modules/documents/approval/repository"
)

// Clock abstracts time so services can be tested deterministically.
type Clock interface {
	Now() time.Time
}

// RealClock is the production Clock implementation.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

// Services is the top-level application service container for the approval
// subsystem. Each field is a focused service; all share the same repo,
// emitter, and clock.
type Services struct {
	Submit     *SubmitService
	Decision   *DecisionService
	Publish    *PublishService
	Scheduler  *SchedulerService
	Supersede  *SupersedeService
	Obsolete   *ObsoleteService
	Cancel     *CancelService
	Read       *ReadService
	RouteAdmin *RouteAdminService
	clock      Clock
}

type ScheduledPublishJobInput struct {
	TenantID                string
	DocumentID              string
	ExpectedRevisionVersion int
	ScheduledEffectiveAt    time.Time
	ScheduleGeneration      int64
}

type ScheduledPublishEnqueuer interface {
	EnqueueScheduledPublishTx(ctx context.Context, tx *sql.Tx, input ScheduledPublishJobInput) error
}

var ErrContentHashMismatch = errors.New("approval: content hash mismatch")

// NewServices constructs a fully wired Services value.
func NewServices(repo repository.ApprovalRepository, emitter EventEmitter, clock Clock) *Services {
	return &Services{
		Submit:     &SubmitService{repo: repo, emitter: emitter, clock: clock},
		Decision:   &DecisionService{repo: repo, emitter: emitter, clock: clock},
		Publish:    &PublishService{repo: repo, emitter: emitter, clock: clock},
		Scheduler:  &SchedulerService{repo: repo, emitter: emitter, clock: clock},
		Supersede:  &SupersedeService{repo: repo, emitter: emitter, clock: clock},
		Obsolete:   &ObsoleteService{repo: repo, emitter: emitter, clock: clock},
		Cancel:     newCancelService(repo, emitter, clock),
		Read:       newReadService(repo),
		RouteAdmin: &RouteAdminService{repo: repo, emitter: emitter, clock: clock},
		clock:      clock,
	}
}

func (s *Services) WithScheduledPublishEnqueuer(enqueuer ScheduledPublishEnqueuer) *Services {
	if s != nil && s.Publish != nil {
		s.Publish = s.Publish.WithScheduledPublishEnqueuer(enqueuer)
	}
	return s
}

// ValidateEventPayload returns ErrFloatInFormData if any value in payload is a
// float64. JSON unmarshal defaults numeric values to float64, which breaks
// canonical hashing; callers must use strings or ints instead.
func ValidateEventPayload(payload map[string]any) error {
	for _, v := range payload {
		if _, ok := v.(float64); ok {
			return ErrFloatInFormData
		}
	}
	return nil
}
