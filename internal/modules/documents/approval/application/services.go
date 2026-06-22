package application

import (
	"context"
	"errors"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/modules/documents/approval/repository"
	"metaldocs/internal/platform/db"
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
	EnqueueScheduledPublishTx(ctx context.Context, tx db.Tx, input ScheduledPublishJobInput) error
}

var ErrContentHashMismatch = errors.New("approval: content hash mismatch")

// NewServices constructs a fully wired Services value. cdRead is the
// controlleddocuments read-port (M2/F2.1) used by the area-grade authz checks in
// Submit/Publish/Supersede/Decision; a nil reader fail-closes the CD area to "".
func NewServices(repo repository.ApprovalRepository, emitter EventEmitter, clock Clock, cdRead controlleddocumentsdomain.CDFieldReader) *Services {
	return &Services{
		Submit:     &SubmitService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Decision:   &DecisionService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Publish:    &PublishService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Scheduler:  &SchedulerService{repo: repo, emitter: emitter, clock: clock},
		Supersede:  &SupersedeService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Obsolete:   &ObsoleteService{repo: repo, emitter: emitter, clock: clock},
		Cancel:     newCancelService(repo, emitter, clock),
		Read:       newReadService(repo, cdRead),
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

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer to all four emit
// services (Publish, Supersede, Obsolete, Decision). Call after NewServices.
func (s *Services) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *Services {
	if s == nil {
		return s
	}
	if s.Publish != nil {
		s.Publish = s.Publish.WithLifecycleEnqueuer(e)
	}
	if s.Supersede != nil {
		s.Supersede = s.Supersede.WithLifecycleEnqueuer(e)
	}
	if s.Obsolete != nil {
		s.Obsolete = s.Obsolete.WithLifecycleEnqueuer(e)
	}
	if s.Decision != nil {
		s.Decision = s.Decision.WithLifecycleEnqueuer(e)
	}
	return s
}

// WithRouteAdminIdempStore returns the Services with route-admin idempotency
// wired. The store is optional — unset leaves the no-store path in place
// (acceptable in tests; production composition root MUST set it).
func (s *Services) WithRouteAdminIdempStore(store RouteAdminIdempStore) *Services {
	if s != nil && s.RouteAdmin != nil {
		s.RouteAdmin = s.RouteAdmin.WithIdempStore(store)
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
