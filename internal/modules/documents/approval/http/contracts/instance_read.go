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

// InboxItem is one row of an approver's inbox listing.
type InboxItem struct {
	InstanceID           string `json:"instance_id"`
	DocumentID           string `json:"document_id"`
	ControlledDocumentID string `json:"controlled_document_id"`
	DocumentTitle        string `json:"document_title"`
	AreaCode             string `json:"area_code"`
	SubmittedBy          string `json:"submitted_by"`
	SubmittedAt          string `json:"submitted_at"`
	StageLabel           string `json:"stage_label"`
	QuorumProgress       string `json:"quorum_progress"`
	// StageKind (F8, spec.md §4/W4): review or approval, omitted when unknown
	// (defensive; the service always populates it from the stage snapshot).
	StageKind string `json:"stage_kind,omitempty"`
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
