package contracts

// InstanceStatus is the wire representation of an approval instance's top-level status.
type InstanceStatus string

// InstanceStatus values.
const (
	InstanceStatusApproved         InstanceStatus = "approved"
	InstanceStatusCancelled        InstanceStatus = "cancelled"
	InstanceStatusInProgress       InstanceStatus = "in_progress"
	InstanceStatusRejected         InstanceStatus = "rejected"
	InstanceStatusChangesRequested InstanceStatus = "changes_requested"
)

// IsValidInstanceStatus reports whether s is a wire-representable instance
// status. The wire enum must cover every domain status (domain/instance.go):
// changes_requested (F4/W11) was the last carried gap — a review verdict of
// request_changes transitions the instance to this non-terminal status, so a
// read of that instance must serialize a value the FE can parse.
func IsValidInstanceStatus(s string) bool {
	switch InstanceStatus(s) {
	case InstanceStatusApproved, InstanceStatusCancelled, InstanceStatusInProgress,
		InstanceStatusRejected, InstanceStatusChangesRequested:
		return true
	default:
		return false
	}
}

// SignoffDecision is the wire representation of an approve/reject vote.
type SignoffDecision string

// SignoffDecision values.
const (
	SignoffDecisionApprove SignoffDecision = "approve"
	SignoffDecisionReject  SignoffDecision = "reject"
)

// SignatureMethod is the wire representation of the method used to sign a decision.
type SignatureMethod string

// SignatureMethod values.
const (
	SignatureMethodPasswordReauth SignatureMethod = "password_reauth"
)

// InstanceResponse is the response body for a GET on a single approval instance.
// ETag carries the OCC token clients must echo via If-Match on subsequent writes.
type InstanceResponse struct {
	ID          string          `json:"id"`
	DocumentID  string          `json:"document_id"`
	RouteID     string          `json:"route_id"`
	TenantID    string          `json:"tenant_id"`
	Status      InstanceStatus  `json:"status"`
	SubmittedBy string          `json:"submitted_by"`
	SubmittedAt string          `json:"submitted_at"`
	CompletedAt *string         `json:"completed_at,omitempty"`
	Stages      []StageInstance `json:"stages"`
	ETag        string          `json:"etag"`
	// FrozenContentHash (F6/F8): the SHA-256 hex digest pinned at freeze time
	// (no-fallback principle, spec §11). Required-and-nullable on the wire
	// (present as explicit null, never omitted) — nil before the freeze boundary.
	FrozenContentHash *string `json:"frozen_content_hash"`
	// Viewer (F2d.1, ADR 0078) is the server-derived eligibility-truth block
	// for the authenticated caller. Present only on the by-document view
	// (GetInstanceByDocumentHandler); omitted on the by-id read
	// (GetInstanceHandler), which is not scoped to a single viewer's workspace.
	Viewer *ViewerFacts `json:"viewer,omitempty"`
	// Verdicts (F2d.2, ADR 0079) is the instance's review-verdict history across
	// ALL stages, chronological (verdict_at asc), empty ⇒ []. A pointer so the
	// by-document view emits it as a present (possibly empty) array while the
	// by-id read omits it entirely (nil pointer) — mirroring the Viewer split.
	// actor_display_name is the immutable snapshot at action time.
	Verdicts *[]VerdictRecord `json:"verdicts,omitempty"`
}

// VerdictRecord is the wire representation of one recorded review verdict
// (F2d.2, ADR 0079). Reason carries the verdict comment (required and non-empty
// for request_changes, optional for ready).
type VerdictRecord struct {
	ID               string  `json:"id"`
	StageInstanceID  string  `json:"stage_instance_id"`
	ActorUserID      string  `json:"actor_user_id"`
	ActorDisplayName string  `json:"actor_display_name"`
	Verdict          string  `json:"verdict"`
	Reason           *string `json:"reason"`
	VerdictAt        string  `json:"verdict_at"`
}

// ViewerFacts (F2d.1, ADR 0078) is the server-derived eligibility truth for
// the authenticated caller against the instance's active stage. Display
// facts only — the DTO never gates the server; enforcement stays at
// authz.Require (tier-2) plus the write-path CheckEligibility/CheckSoD.
type ViewerFacts struct {
	IsAuthor               bool                  `json:"is_author"`
	EligibleForActiveStage bool                  `json:"eligible_for_active_stage"`
	HasSignedActiveStage   bool                  `json:"has_signed_active_stage"`
	ViaDelegationFrom      *ViewerDelegationFrom `json:"via_delegation_from"`
}

// ViewerDelegationFrom identifies the delegator granting the viewer
// eligibility, when eligibility is satisfied purely via delegation.
type ViewerDelegationFrom struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// StageInstance is the wire representation of one stage's runtime state within an InstanceResponse.
type StageInstance struct {
	ID         string          `json:"id"`
	StageIndex int             `json:"stage_index"`
	Label      string          `json:"label"`
	Status     string          `json:"status"`
	Signoffs   []SignoffRecord `json:"signoffs"`
	Actors     []StageActor    `json:"actors"`
	// StageKind (F8, spec.md §4/W4): review or approval.
	StageKind string `json:"stage_kind,omitempty"`
	// DueAt (F8, spec.md §4/W4): RFC3339 UTC SLA due date. Required-and-nullable
	// on the wire (present as explicit null, never omitted) — nil while pending
	// or when no SLA is configured (no-fallback principle, spec §11).
	DueAt *string `json:"due_at"`
}

// StageActor is one eligible actor's status within a StageInstance.
type StageActor struct {
	UserID      string           `json:"user_id"`
	DisplayName string           `json:"display_name"`
	Status      string           `json:"status"`
	Decision    *SignoffDecision `json:"decision,omitempty"`
}

// SignoffRecord is the wire representation of a recorded signoff.
type SignoffRecord struct {
	ID               string          `json:"id"`
	ActorUserID      string          `json:"actor_user_id"`
	Decision         SignoffDecision `json:"decision"`
	Reason           string          `json:"reason,omitempty"`
	SignatureMethod  SignatureMethod `json:"signature_method"`
	SignedAt         string          `json:"signed_at"`
	SignatureMeaning string          `json:"signature_meaning"`
}

// InboxItem is one row of an approver's inbox listing. Unit 4.2 made this
// subject-generic: SubjectKind/SubjectKey/SubjectTitle/SubjectRef cover both
// document and template instances; ControlledDocumentID stays document-only
// (explicit null for template rows). document_id/document_title were DROPPED (hard
// break, no relax-to-optional compat — the sole consumer, the approval inbox
// FE, is updated in the same unit).
type InboxItem struct {
	InstanceID string `json:"instance_id"`
	// SubjectKind is the instance's subject discriminator: "document" or
	// "template".
	SubjectKind string `json:"subject_kind"`
	// SubjectKey is the kernel subject key: documentId for document
	// instances, template version id for template instances.
	SubjectKey string `json:"subject_key"`
	// SubjectTitle is the display name: document name or template name.
	SubjectTitle string `json:"subject_title"`
	// SubjectRef is the FE navigation id: documentId for document instances,
	// templateId for template instances (NOT the version id).
	SubjectRef string `json:"subject_ref"`
	// ControlledDocumentID is document-instance-only. Required-and-nullable on
	// the wire (present as explicit null, never omitted) — null for template
	// rows (legacy-fallback-extermination: no relax-to-optional / dropped key).
	ControlledDocumentID *string `json:"controlled_document_id"`
	AreaCode             string  `json:"area_code"`
	SubmittedBy          string  `json:"submitted_by"`
	SubmittedAt          string  `json:"submitted_at"`
	StageLabel           string  `json:"stage_label"`
	QuorumProgress       string  `json:"quorum_progress"`
	// StageKind (F8, spec.md §4/W4): review or approval. Required on the wire
	// (unit 4.2) — approval_stage_instances.stage_kind is NOT NULL DEFAULT
	// 'approval' (0286), so the service always populates it from the stage
	// snapshot. No omitempty: dropping the key would contradict the openapi
	// `required` list.
	StageKind string `json:"stage_kind"`
	// DueAt (F8, spec.md §4/W4): RFC3339 UTC SLA due date. Required-and-nullable
	// on the wire (present as explicit null, never omitted) — nil when no SLA is
	// configured for the stage (no-fallback principle, spec §11).
	DueAt *string `json:"due_at"`
}

// InboxResponse is the response body for the approver inbox listing endpoint.
type InboxResponse struct {
	Items []InboxItem `json:"items"`
	Total int         `json:"total"`
}
