// Package application holds the approval subsystem's application services —
// submit, decision, publish, scheduler, supersede, obsolete, cancel, read, and
// route-admin — each a thin orchestration layer over infrastructure.ApprovalRepository
// and domain. All mutating operations run inside a single caller-owned
// transaction (state-write + governance-event emit + outbox enqueue never
// span more than one tx), and callers are expected to seed authz identity and
// call authz.Require before any write.
package application

import (
	"context"
	"errors"
	"time"

	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	"metaldocs/internal/modules/documents/approval/infrastructure"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// Clock abstracts time so services can be tested deterministically.
type Clock interface {
	Now() time.Time
}

// RealClock is the production Clock implementation.
type RealClock struct{}

// Now returns the current UTC time.
func (RealClock) Now() time.Time { return time.Now().UTC() }

// Services is the top-level application service container for the approval
// subsystem. Each field is a focused service; all share the same repo,
// emitter, and clock.
type Services struct {
	Submit        *SubmitService
	Decision      *DecisionService
	Publish       *PublishService
	Scheduler     *SchedulerService
	Supersede     *SupersedeService
	Obsolete      *ObsoleteService
	Cancel        *CancelService
	Read          *ReadService
	RouteAdmin    *RouteAdminService
	MarkReviewed  *MarkReviewedService
	ReviewVerdict *ReviewVerdictService
	Delegation    *DelegationService
	FastForward   *FastForwardService
	clock         Clock
}

// ScheduledPublishJobInput carries the parameters needed to enqueue a scheduled-publish job.
type ScheduledPublishJobInput struct {
	TenantID                string
	DocumentID              string
	ExpectedRevisionVersion int
	ScheduledEffectiveAt    time.Time
	ScheduleGeneration      int64
}

// ScheduledPublishEnqueuer enqueues a scheduled-publish job in the same transaction
// as the state write that scheduled it (transactional outbox pattern).
type ScheduledPublishEnqueuer interface {
	EnqueueScheduledPublishTx(ctx context.Context, tx db.Tx, input ScheduledPublishJobInput) error
}

// ErrContentHashMismatch is returned when the caller-supplied content hash does not
// match the document's current content hash, indicating a stale read (OCC guard).
var ErrContentHashMismatch = errors.New("approval: content hash mismatch")

// NewServices constructs a fully wired Services value. cdRead is the
// controlleddocuments read-port (M2/F2.1) used by the area-grade authz checks in
// Submit/Publish/Supersede/Decision; a nil reader fail-closes the CD area to "".
func NewServices(repo infrastructure.ApprovalRepository, emitter EventEmitter, clock Clock, cdRead controlleddocumentsdomain.CDFieldReader) *Services {
	// The concrete Postgres repository also satisfies SubmitDefaultsResolver
	// (in-tx route/hash resolution, ADR 0073). Wire it via assertion so the
	// broad ApprovalRepository interface stays untouched (ISP); a repo that does
	// not implement it leaves submit resolution disabled (nil), which fail-closes
	// the omitted-route path to ErrApprovalRouteMissing.
	resolver, _ := repo.(SubmitDefaultsResolver)
	decision := &DecisionService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead}
	reviewVerdict := &ReviewVerdictService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead}
	return &Services{
		Submit:        &SubmitService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead, resolver: resolver},
		Decision:      decision,
		Publish:       &PublishService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Scheduler:     &SchedulerService{repo: repo, emitter: emitter, clock: clock},
		Supersede:     &SupersedeService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead},
		Obsolete:      &ObsoleteService{repo: repo, emitter: emitter, clock: clock},
		Cancel:        newCancelService(repo, emitter, clock),
		Read:          newReadService(repo, cdRead),
		RouteAdmin:    &RouteAdminService{repo: repo, emitter: emitter, clock: clock},
		MarkReviewed:  NewMarkReviewedService(emitter, clock),
		ReviewVerdict: reviewVerdict,
		Delegation:    newDelegationService(repo, emitter, clock),
		FastForward:   newFastForwardService(reviewVerdict, decision),
		clock:         clock,
	}
}

// WithScheduledPublishEnqueuer wires the scheduled-publish job enqueuer into the
// Publish service. Call after NewServices.
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
	if s.ReviewVerdict != nil {
		s.ReviewVerdict = s.ReviewVerdict.WithLifecycleEnqueuer(e)
	}
	return s
}

// WithProfilePolicyReader wires the G1 approval→taxonomy profile-policy reader
// into the two route-shape enforcement sites (Submit and RouteAdmin). Call after
// NewServices. A nil reader leaves the friendly route-shape check disabled; the
// DB deferrable constraint trigger remains the authoritative last line, so a
// misconfigured composition root degrades to DB-enforced, never to unenforced.
func (s *Services) WithProfilePolicyReader(reader ProfilePolicyReader) *Services {
	if s == nil {
		return s
	}
	if s.Submit != nil {
		s.Submit = s.Submit.WithPolicyReader(reader)
	}
	if s.RouteAdmin != nil {
		s.RouteAdmin = s.RouteAdmin.WithPolicyReader(reader)
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
