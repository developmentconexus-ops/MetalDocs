package application

import (
	"metaldocs/internal/modules/approval/infrastructure"
)

// Cross-module error re-exports (M3 P3.S2b-4, R2a).
//
// application (this package) is a published cross-module surface —
// scripts/check-module-boundaries.ps1's allow-model lets any module import
// <module>/domain, <module>/application, and <module>/api, but NEVER
// <module>/infrastructure. Several kernel services (TemplateSubmitService,
// DecisionService) return infrastructure.Err* sentinels UNWRAPPED straight
// out of their published application-layer functions (e.g.
// TemplateSubmitService.SubmitTemplateVersionForReview returns
// infrastructure.ErrNoActiveApprovalRoute directly). A cross-module caller
// like templates' thin kernel HTTP entry points (submit-for-approval,
// signoff) needs errors.Is against those exact sentinel values to classify
// the failure into the right RFC 9457 problem+json response, but it cannot
// import approval/infrastructure to reference them without breaking the
// module boundary.
//
// These vars alias (not wrap) the infrastructure sentinels: same
// underlying *errors.errorString value, so errors.Is(err, ErrNoActiveInstance)
// from ANY package matches identically to errors.Is(err,
// infrastructure.ErrNoActiveInstance). This is the published-surface
// counterpart of what approval's OWN http package already does internally
// (approval/http/errors.go's MapErrorToResponse switches on the same
// infrastructure sentinels) — the alias just makes that classification
// available to an out-of-module caller through the sanctioned layer.
//
// Only the subset reachable by the CURRENT cross-module kernel entry points
// (TemplateSubmitService.SubmitTemplateVersionForReview,
// DecisionService.RecordSignoff for a template subject) is re-exported here.
// Extend this list, never approval/infrastructure imports elsewhere, if a
// future cross-module caller needs to classify another sentinel.
var (
	// ErrNoActiveInstance re-exports infrastructure.ErrNoActiveInstance —
	// returned by DecisionService.RecordSignoff (via the caller-supplied
	// InstanceID) when no in-progress instance matches.
	ErrNoActiveInstance = infrastructure.ErrNoActiveInstance
	// ErrInstanceCompleted re-exports infrastructure.ErrInstanceCompleted —
	// returned by DecisionService.RecordSignoff when the instance already
	// reached a terminal status.
	ErrInstanceCompleted = infrastructure.ErrInstanceCompleted
	// ErrStageNotActive re-exports infrastructure.ErrStageNotActive —
	// returned by DecisionService.RecordSignoff when the supplied stage
	// instance is not the active stage (race between resolve and record).
	ErrStageNotActive = infrastructure.ErrStageNotActive
	// ErrDuplicateSubmission re-exports infrastructure.ErrDuplicateSubmission
	// — returned by TemplateSubmitService.SubmitTemplateVersionForReview when
	// a concurrent submission with the same idempotency key already exists.
	ErrDuplicateSubmission = infrastructure.ErrDuplicateSubmission
	// ErrNoActiveApprovalRoute re-exports infrastructure.ErrNoActiveApprovalRoute
	// — returned by TemplateSubmitService.SubmitTemplateVersionForReview when
	// the template has no active approval route.
	ErrNoActiveApprovalRoute = infrastructure.ErrNoActiveApprovalRoute
)
