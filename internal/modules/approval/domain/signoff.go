package domain

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var hashRegex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Decision represents an approve or reject vote.
type Decision string

// Decision values a Signoff may carry.
const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

// Signoff is an immutable value object. All fields are unexported.
type Signoff struct {
	id                       string
	approvalInstanceID       string
	stageInstanceID          string
	actorUserID              string
	actorTenantID            string
	decision                 Decision
	comment                  string
	signedAt                 time.Time
	signatureMethod          string
	signaturePayload         json.RawMessage
	contentHash              string // always lowercase hex sha-256
	actorDisplayNameSnapshot string
	signatureMeaning         string
	onBehalfOfUserID         string // F9/ADR 0077: delegator's user_id, "" when acting as self
}

// Getters — no setters exist; immutable after construction.

// ID returns the signoff's identifier.
func (s *Signoff) ID() string { return s.id }

// ApprovalInstanceID returns the parent approval instance's identifier.
func (s *Signoff) ApprovalInstanceID() string { return s.approvalInstanceID }

// StageInstanceID returns the stage instance this signoff was recorded against.
func (s *Signoff) StageInstanceID() string { return s.stageInstanceID }

// ActorUserID returns the signing user's identifier.
func (s *Signoff) ActorUserID() string { return s.actorUserID }

// ActorTenantID returns the signing user's tenant identifier.
func (s *Signoff) ActorTenantID() string { return s.actorTenantID }

// Decision returns the approve/reject vote.
func (s *Signoff) Decision() Decision { return s.decision }

// Comment returns the signer's optional comment.
func (s *Signoff) Comment() string { return s.comment }

// SignedAt returns the timestamp the signoff was recorded.
func (s *Signoff) SignedAt() time.Time { return s.signedAt }

// SignatureMethod returns the signature method used (e.g. password re-auth).
func (s *Signoff) SignatureMethod() string { return s.signatureMethod }

// SignaturePayload returns the method-specific signature evidence, opaque to callers.
func (s *Signoff) SignaturePayload() json.RawMessage { return s.signaturePayload }

// ContentHash returns the lowercase hex sha-256 of the document content at signing time.
func (s *Signoff) ContentHash() string { return s.contentHash }

// ActorDisplayNameSnapshot returns the signer's display name as captured at signing time.
func (s *Signoff) ActorDisplayNameSnapshot() string { return s.actorDisplayNameSnapshot }

// SignatureMeaning returns the meaning of this signature: "approval" or "rejection".
func (s *Signoff) SignatureMeaning() string { return s.signatureMeaning }

// OnBehalfOf returns the delegator's user_id when this signoff was recorded
// via an active delegation (F9/ADR 0077); "" when the actor signed as
// themselves.
func (s *Signoff) OnBehalfOf() string { return s.onBehalfOfUserID }

// SignoffParams holds constructor inputs.
type SignoffParams struct {
	ID                       string
	ApprovalInstanceID       string
	StageInstanceID          string
	ActorUserID              string
	ActorTenantID            string
	Decision                 Decision
	Comment                  string
	SignedAt                 time.Time
	SignatureMethod          string
	SignaturePayload         json.RawMessage
	ContentHash              string
	ActorDisplayNameSnapshot string
	SignatureMeaning         string
	OnBehalfOfUserID         string // F9/ADR 0077: delegator's user_id, "" when acting as self
}

// NewSignoff constructs an immutable Signoff value object.
// ContentHash is normalized to lowercase; uppercase hex is accepted.
func NewSignoff(p SignoffParams) (*Signoff, error) {
	if p.ID == "" {
		return nil, errors.New("signoff: ID is required")
	}
	if p.ApprovalInstanceID == "" {
		return nil, errors.New("signoff: ApprovalInstanceID is required")
	}
	if p.StageInstanceID == "" {
		return nil, errors.New("signoff: StageInstanceID is required")
	}
	if p.ActorUserID == "" {
		return nil, errors.New("signoff: ActorUserID is required")
	}
	if p.ActorTenantID == "" {
		return nil, errors.New("signoff: ActorTenantID is required")
	}
	if p.Decision != DecisionApprove && p.Decision != DecisionReject {
		return nil, errors.New("signoff: Decision must be 'approve' or 'reject'")
	}
	if p.SignedAt.IsZero() {
		return nil, errors.New("signoff: SignedAt must not be zero")
	}

	// Normalize hash to lowercase, then validate.
	hash := strings.ToLower(p.ContentHash)
	if !hashRegex.MatchString(hash) {
		return nil, errors.New("signoff: ContentHash must be 64 lowercase hex chars (sha-256)")
	}

	// SignatureMeaning defaults to "approval" (matches the DB column default)
	// so existing callers that don't set this field keep working unchanged.
	signatureMeaning := p.SignatureMeaning
	if signatureMeaning == "" {
		signatureMeaning = "approval"
	} else if signatureMeaning != "approval" && signatureMeaning != "rejection" {
		return nil, errors.New("signoff: SignatureMeaning must be 'approval' or 'rejection'")
	}

	return &Signoff{
		id:                       p.ID,
		approvalInstanceID:       p.ApprovalInstanceID,
		stageInstanceID:          p.StageInstanceID,
		actorUserID:              p.ActorUserID,
		actorTenantID:            p.ActorTenantID,
		decision:                 p.Decision,
		comment:                  p.Comment,
		signedAt:                 p.SignedAt,
		signatureMethod:          p.SignatureMethod,
		signaturePayload:         p.SignaturePayload,
		contentHash:              hash,
		actorDisplayNameSnapshot: p.ActorDisplayNameSnapshot,
		signatureMeaning:         signatureMeaning,
		onBehalfOfUserID:         p.OnBehalfOfUserID,
	}, nil
}

// MarshalJSON exposes Signoff for API responses.
func (s *Signoff) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"id":                          s.id,
		"approval_instance_id":        s.approvalInstanceID,
		"stage_instance_id":           s.stageInstanceID,
		"actor_user_id":               s.actorUserID,
		"actor_tenant_id":             s.actorTenantID,
		"decision":                    s.decision,
		"comment":                     s.comment,
		"signed_at":                   s.signedAt,
		"signature_method":            s.signatureMethod,
		"content_hash":                s.contentHash,
		"actor_display_name_snapshot": s.actorDisplayNameSnapshot,
		"signature_meaning":           s.signatureMeaning,
		"on_behalf_of_user_id":        s.onBehalfOfUserID,
	})
}
