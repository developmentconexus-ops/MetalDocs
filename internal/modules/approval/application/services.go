// Package application holds the approval subsystem's application services —
// submit, decision, obsolete, cancel, read, and route-admin — each a thin
// orchestration layer over infrastructure.ApprovalRepository
// and domain. All mutating operations run inside a single caller-owned
// transaction (state-write + governance-event emit + outbox enqueue never
// span more than one tx), and callers are expected to seed authz identity and
// call authz.Require before any write.
package application

import (
	"errors"
	"time"

	approvaldomain "metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docsdomain "metaldocs/internal/modules/documents/domain"
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
	Submit         *SubmitService
	TemplateSubmit *TemplateSubmitService
	Decision       *DecisionService
	Obsolete       *ObsoleteService
	Cancel         *CancelService
	Read           *ReadService
	RouteAdmin     *RouteAdminService
	MarkReviewed   *MarkReviewedService
	ReviewVerdict  *ReviewVerdictService
	Delegation     *DelegationService
	FastForward    *FastForwardService
	clock          Clock
}

// ErrContentHashMismatch is returned when the caller-supplied content hash does not
// match the document's current content hash, indicating a stale read (OCC guard).
var ErrContentHashMismatch = errors.New("approval: content hash mismatch")

// NewServices constructs a fully wired Services value. cdRead is the
// controlleddocuments read-port (M2/F2.1) used by the area-grade authz checks in
// Submit/Decision; a nil reader fail-closes the CD area to "".
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
		Submit:         &SubmitService{repo: repo, emitter: emitter, clock: clock, cdRead: cdRead, resolver: resolver},
		TemplateSubmit: NewTemplateSubmitService(repo, emitter, clock, nil, nil),
		Decision:       decision,
		Obsolete:       &ObsoleteService{repo: repo, emitter: emitter, clock: clock},
		Cancel:         newCancelService(repo, emitter, clock),
		Read:           newReadService(repo, cdRead, nil),
		RouteAdmin:     &RouteAdminService{repo: repo, emitter: emitter, clock: clock},
		MarkReviewed:   NewMarkReviewedService(emitter, clock),
		ReviewVerdict:  reviewVerdict,
		Delegation:     newDelegationService(repo, emitter, clock),
		FastForward:    newFastForwardService(reviewVerdict, decision),
		clock:          clock,
	}
}

// WithLifecycleEnqueuer wires the F3.3 domain-event enqueuer to the emit
// services (Obsolete, Decision, ReviewVerdict). Call after NewServices.
func (s *Services) WithLifecycleEnqueuer(e docsdomain.LifecycleEventEnqueuer) *Services {
	if s == nil {
		return s
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
	// ADR 0087: Submit is a fourth emit site — a zero-stage (livre) route
	// reaches terminal approval inside the submit tx, so it must enqueue the
	// same document.approved lifecycle event the other routes do.
	if s.Submit != nil {
		s.Submit = s.Submit.WithLifecycleEnqueuer(e)
	}
	return s
}

// WithReleaseRecorder wires the ADR 0085 terminal-approval seam into BOTH
// routes to terminal approval — signoff (Decision) and review verdict
// (ReviewVerdict). Before ADR 0085 only the signoff route pinned, and the
// review-verdict route silently produced an approved document that could never
// materialize (F-QA4-14). One collaborator wired from one call site is what
// makes that class of asymmetry unrepresentable.
func (s *Services) WithReleaseRecorder(recorder TerminalApprovalReleaseRecorder) *Services {
	if s == nil {
		return s
	}
	if s.Decision != nil {
		s.Decision = s.Decision.WithReleaseRecorder(recorder)
	}
	if s.ReviewVerdict != nil {
		s.ReviewVerdict = s.ReviewVerdict.WithReleaseRecorder(recorder)
	}
	// ADR 0087 adds a THIRD route to terminal approval: submit against a
	// zero-stage (livre) route. Same one-call-site rule — wiring it here is
	// what keeps the auto-approve leg from being the next F-QA4-14.
	if s.Submit != nil {
		s.Submit = s.Submit.WithReleaseRecorder(recorder)
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

// WithTemplateVersionReader wires the approval-owned TemplateVersionReader
// port into TemplateSubmit (M3 P3.S2b-3b-ii). Call after NewServices. A nil
// reader leaves the draft-status guard fail-closed (ErrTemplateVersionNotFound
// on every submit) rather than unenforced.
func (s *Services) WithTemplateVersionReader(reader TemplateVersionReader) *Services {
	if s == nil {
		return s
	}
	if s.TemplateSubmit != nil {
		// Set the field in place (rather than reconstructing) so a previously
		// wired versionWriter is preserved regardless of call order.
		s.TemplateSubmit.versionReader = reader
	}
	// M3 P3.S2b-3b-iii-b: DecisionService reuses the same port to read a
	// template version's content_hash in place of the document-only freeze
	// pin during signoff.
	if s.Decision != nil {
		s.Decision = s.Decision.WithTemplateVersionReader(reader)
	}
	// Unit 4.2: Read's ListWorklist reuses the same port's
	// LoadTemplateInboxMeta to resolve template-subject rows' display
	// metadata. The composition root's single WithTemplateVersionReader call
	// (apps/api/cmd/metaldocs-api/main.go) wires this with no change needed
	// there.
	if s.Read != nil {
		s.Read.templateRead = reader
	}
	return s
}

// WithTemplateCompletionWriter wires the M3 P3.S2b-3b-iii-b completion port
// into Decision. Call after NewServices. A nil writer leaves a
// template-subject terminal decision fail-closed (an explicit error) rather
// than silently skipping the templates_template_version transition.
func (s *Services) WithTemplateCompletionWriter(writer TemplateCompletionWriter) *Services {
	if s == nil {
		return s
	}
	if s.Decision != nil {
		s.Decision = s.Decision.WithTemplateCompletionWriter(writer)
	}
	// The same templates-infra adapter (ApprovalCompletionWriter) also
	// satisfies the submit-lock port (M3 P3.S2b-3b-iii-b, hub Option (a)):
	// one concrete adapter owns the full template-version lifecycle under
	// kernel governance (under_review at submit, approved/draft at
	// completion). Wire the submit-lock leg when the writer implements it.
	if s.TemplateSubmit != nil {
		if sw, ok := writer.(TemplateVersionSubmitWriter); ok {
			s.TemplateSubmit.versionWriter = sw
		}
		// ADR 0087: a livre profile's template route carries zero stages, so
		// template submit itself reaches terminal approval and needs the SAME
		// completion port DecisionService uses — one adapter, one lifecycle.
		s.TemplateSubmit.templateCompletion = writer
	}
	return s
}

// WithApprovalNotificationEnqueuer wires the Task 6 approval.pending
// notification port into BOTH submit routes — Submit (document) and
// TemplateSubmit (template). Call after NewServices. A nil enqueuer (or an
// unset field) leaves each service fail-closed to
// domain.NoopApprovalNotificationEnqueuer: it enqueues nothing and reports
// success, so a composition root that forgets to call this degrades to
// silent-no-notification rather than a panic — the same fail-closed shape
// every other optional port on Services uses.
func (s *Services) WithApprovalNotificationEnqueuer(notifier approvaldomain.ApprovalNotificationEnqueuer) *Services {
	if s == nil {
		return s
	}
	if s.Submit != nil {
		s.Submit = s.Submit.WithApprovalNotificationEnqueuer(notifier)
	}
	if s.TemplateSubmit != nil {
		s.TemplateSubmit.notifier = notifier
		if s.TemplateSubmit.notifier == nil {
			s.TemplateSubmit.notifier = approvaldomain.NoopApprovalNotificationEnqueuer{}
		}
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
