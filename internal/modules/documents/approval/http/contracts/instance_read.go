package contracts

type InstanceStatus string

const (
	InstanceStatusInProgress InstanceStatus = "in_progress"
	InstanceStatusApproved   InstanceStatus = "approved"
	InstanceStatusRejected   InstanceStatus = "rejected"
	InstanceStatusCancelled  InstanceStatus = "cancelled"
)

type SignoffDecision string

const (
	SignoffDecisionApprove SignoffDecision = "approve"
	SignoffDecisionReject  SignoffDecision = "reject"
)

type SignatureMethod string

const (
	SignatureMethodPasswordReauth SignatureMethod = "password_reauth"
)

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
}

type StageInstance struct {
	ID         string          `json:"id"`
	StageIndex int             `json:"stage_index"`
	Label      string          `json:"label"`
	Status     string          `json:"status"`
	Signoffs   []SignoffRecord `json:"signoffs"`
	Actors     []StageActor    `json:"actors"`
}

type StageActor struct {
	UserID      string           `json:"user_id"`
	DisplayName string           `json:"display_name"`
	Status      string           `json:"status"`
	Decision    *SignoffDecision `json:"decision,omitempty"`
}

type SignoffRecord struct {
	ID              string          `json:"id"`
	ActorUserID     string          `json:"actor_user_id"`
	Decision        SignoffDecision `json:"decision"`
	Reason          string          `json:"reason,omitempty"`
	SignatureMethod SignatureMethod `json:"signature_method"`
	SignedAt        string          `json:"signed_at"`
}

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
}

type InboxResponse struct {
	Items []InboxItem `json:"items"`
	Total int         `json:"total"`
}
