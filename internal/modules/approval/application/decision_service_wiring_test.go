package application

import (
	"context"
	"strings"
	"testing"

	"metaldocs/internal/modules/approval/domain"
	"metaldocs/internal/modules/approval/infrastructure/signature"
	controlleddocumentsdomain "metaldocs/internal/modules/controlleddocuments/domain"
	docsdomain "metaldocs/internal/modules/documents/domain"
	"metaldocs/internal/platform/db"
)

// wiringStubTemplateReader is an inert TemplateVersionReader used only to
// assert wiring/identity, not behavior.
type wiringStubTemplateReader struct{}

func (wiringStubTemplateReader) LoadTemplateVersionStatus(context.Context, db.Tx, string, string) (string, bool, error) {
	return "", false, nil
}

func (wiringStubTemplateReader) LoadTemplateVersionContentHash(context.Context, db.Tx, string, string) (string, bool, error) {
	return "", false, nil
}

func (wiringStubTemplateReader) LoadTemplateInboxMeta(context.Context, db.Tx, string, []string) (map[string]domain.TemplateInboxMeta, error) {
	return nil, nil
}

// wiringStubTemplateCompletion is an inert TemplateCompletionWriter used only
// to assert wiring/identity, not behavior.
type wiringStubTemplateCompletion struct{}

func (wiringStubTemplateCompletion) MarkTemplateVersionApproved(context.Context, db.Tx, string, string, string) error {
	return nil
}

func (wiringStubTemplateCompletion) MarkTemplateVersionRejected(context.Context, db.Tx, string, string) error {
	return nil
}

// wiringStubLifecycleEnqueuer is an inert docsdomain.LifecycleEventEnqueuer
// used only to assert wiring/identity, not behavior.
type wiringStubLifecycleEnqueuer struct{}

func (wiringStubLifecycleEnqueuer) EnqueueLifecycleEventTx(context.Context, db.Tx, docsdomain.LifecycleEventArgs) error {
	return nil
}

// TestServicesWithTemplateVersionReader_PreservesDecisionPointerIdentity pins
// that WithTemplateVersionReader mutates the existing Decision in place
// rather than replacing it — a composition root that rebuilds Decision after
// this call (main.go's former mistake) would silently drop the wiring.
func TestServicesWithTemplateVersionReader_PreservesDecisionPointerIdentity(t *testing.T) {
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, nil)
	decisionBefore := svc.Decision

	svc = svc.WithTemplateVersionReader(wiringStubTemplateReader{})

	if svc.Decision != decisionBefore {
		t.Fatalf("WithTemplateVersionReader replaced the Decision pointer instead of mutating it in place")
	}
}

// TestServicesWithTemplateCompletionWriter_PreservesDecisionPointerIdentity
// mirrors the above for WithTemplateCompletionWriter.
func TestServicesWithTemplateCompletionWriter_PreservesDecisionPointerIdentity(t *testing.T) {
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, nil)
	decisionBefore := svc.Decision

	svc = svc.WithTemplateCompletionWriter(wiringStubTemplateCompletion{})

	if svc.Decision != decisionBefore {
		t.Fatalf("WithTemplateCompletionWriter replaced the Decision pointer instead of mutating it in place")
	}
}

// TestServicesWithLifecycleEnqueuer_PreservesDecisionPointerIdentity mirrors
// the above for WithLifecycleEnqueuer.
func TestServicesWithLifecycleEnqueuer_PreservesDecisionPointerIdentity(t *testing.T) {
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, nil)
	decisionBefore := svc.Decision

	svc = svc.WithLifecycleEnqueuer(wiringStubLifecycleEnqueuer{})

	if svc.Decision != decisionBefore {
		t.Fatalf("WithLifecycleEnqueuer replaced the Decision pointer instead of mutating it in place")
	}
}

// TestServices_FastForwardSharesDecisionInstance pins the single-instance
// invariant FastForwardService depends on: after every With* call a
// composition root would run (template ports, lifecycle enqueuer, then the
// DecisionService-level ports wired via the same pointer), FastForward's
// decisions field must still be the exact *DecisionService held at
// Services.Decision. If they diverge, FastForward silently uses a
// differently-configured DecisionService than the rest of the system.
func TestServices_FastForwardSharesDecisionInstance(t *testing.T) {
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, controlleddocumentsdomain.NoopCDFieldReader{})
	svc = svc.WithTemplateVersionReader(wiringStubTemplateReader{})
	svc = svc.WithTemplateCompletionWriter(wiringStubTemplateCompletion{})
	svc = svc.WithLifecycleEnqueuer(wiringStubLifecycleEnqueuer{})

	// The composition-root-only ports (pin invoker, signature registry) are
	// wired directly on the DecisionService pointer, exactly as main.go now
	// does: in place, via the returned same pointer, never via a fresh
	// NewDecisionService call.
	svc.Decision = svc.Decision.WithReleaseRecorder(&fakeReleaseRecorder{})

	if svc.FastForward.decisions != svc.Decision {
		t.Fatalf("FastForward.decisions and Services.Decision diverged after wiring: FastForward observes a different *DecisionService than the rest of the system")
	}
}

// TestServices_RebuildingDecisionDivergesFastForward is a regression test
// documenting main.go's former mistake: reassigning Services.Decision to a
// brand-new *DecisionService (via NewDecisionService) instead of mutating the
// existing pointer in place. FastForward was constructed once, in
// NewServices, against the original Decision pointer — a later rebuild
// leaves FastForward silently pointing at the old, differently-configured
// instance while the rest of the composition root (templates kernel wiring,
// approvalHandler) observes the new one.
func TestServices_RebuildingDecisionDivergesFastForward(t *testing.T) {
	svc := NewServices(nil, &noopSQLEmitter{}, RealClock{}, nil)
	originalDecision := svc.Decision

	// Simulate the bug: rebuild Decision from scratch instead of chaining
	// With* on the existing pointer.
	svc.Decision = NewDecisionService(nil, &noopSQLEmitter{}, RealClock{})

	if svc.FastForward.decisions == svc.Decision {
		t.Fatalf("expected FastForward.decisions to diverge from the rebuilt Decision pointer")
	}
	if svc.FastForward.decisions != originalDecision {
		t.Fatalf("expected FastForward.decisions to still reference the original Decision pointer")
	}
}

// TestDecisionService_Ready_ReportsEveryMissingRequiredPort pins that Ready
// names each missing required port in one error, so a composition-root
// regression that drops a port fails fast at boot instead of surfacing as a
// runtime 500 or a silent skip of a state transition.
func TestDecisionService_Ready_ReportsEveryMissingRequiredPort(t *testing.T) {
	s := &DecisionService{}

	err := s.Ready()

	if err == nil {
		t.Fatalf("expected Ready to report missing ports on a zero-value DecisionService")
	}
	for _, port := range []string{"templateVersionReader", "templateCompletion", "releaseRecorder", "sigRegistry", "cdRead"} {
		if !strings.Contains(err.Error(), port) {
			t.Errorf("Ready() error %q does not name missing port %q", err.Error(), port)
		}
	}
}

// TestDecisionService_Ready_NilWhenAllRequiredPortsWired pins that Ready does
// not require lifecycleEnqueuer (best-effort/nil-tolerant per
// WithLifecycleEnqueuer's doc comment) once every other required port is set.
func TestDecisionService_Ready_NilWhenAllRequiredPortsWired(t *testing.T) {
	s := &DecisionService{}
	s = s.WithTemplateVersionReader(wiringStubTemplateReader{}).
		WithTemplateCompletionWriter(wiringStubTemplateCompletion{}).
		WithReleaseRecorder(&fakeReleaseRecorder{}).
		WithCDFieldReader(controlleddocumentsdomain.NoopCDFieldReader{})
	s.sigRegistry = signature.NewRegistry()

	if err := s.Ready(); err != nil {
		t.Fatalf("expected Ready to be nil with all required ports wired, got: %v", err)
	}
}
