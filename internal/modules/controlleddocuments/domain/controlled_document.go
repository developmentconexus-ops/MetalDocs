// Package domain holds the controlled-documents module's aggregate
// (ControlledDocument), its sentinel errors, and the repository /
// sequence-allocator / cross-module ports the application layer depends
// on. It has no database/sql import (CI guard nosqltxindomain) —
// persistence concerns stay behind db.Tx and the repository interfaces
// defined here.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// CDStatus is the lifecycle state of a ControlledDocument: active,
// obsolete, or superseded. Transitions are one-way (active -> obsolete or
// active -> superseded) and are DB-enforced via UpdateStatus.
type CDStatus string

// CDStatus enum values. A ControlledDocument starts CDStatusActive and
// transitions to CDStatusObsolete or CDStatusSuperseded via changeStatus;
// there is no transition back to active.
const (
	CDStatusActive     CDStatus = "active"
	CDStatusObsolete   CDStatus = "obsolete"
	CDStatusSuperseded CDStatus = "superseded"
)

// ControlledDocument is a numbered slot in metaldocs.controlled_documents
// binding a (ProfileCode, ProcessAreaCode) pair to a chain of documents
// revisions. It carries no content of its own — Code is the stable,
// immutable, audit-traceable identity; the chain of documents rows holds
// the actual content.
type ControlledDocument struct {
	ID                        string     `json:"id"`
	TenantID                  string     `json:"tenant_id"`
	ProfileCode               string     `json:"profile_code"`
	ProcessAreaCode           string     `json:"process_area_code"`
	DepartmentCode            *string    `json:"department_code"`
	Code                      string     `json:"code"`
	SequenceNum               *int       `json:"sequence_num"`
	Title                     string     `json:"title"`
	OwnerUserID               string     `json:"owner_user_id"`
	OverrideTemplateVersionID *string    `json:"override_template_version_id"`
	Visibility                Visibility `json:"visibility"`
	Status                    CDStatus   `json:"status"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// Sentinel errors returned by the controlled-documents domain and
// application layers. Callers should compare with errors.Is.
var (
	ErrCDNotFound               = errors.New("controlled document not found")
	ErrNoActiveInstance         = errors.New("no active document instance for this controlled document")
	ErrCDCodeTaken              = errors.New("controlled document code already taken")
	ErrCDArchivedCodeReuse      = errors.New("cannot reuse code from archived controlled document")
	ErrSequenceCounterNotFound  = errors.New("sequence counter not initialized for profile")
	ErrCDNotActive              = errors.New("controlled document is not active")
	ErrActiveRevisionExists     = errors.New("controlled document already has an active revision")
	ErrManualCodeReasonRequired = errors.New("manual code reason is required")
	ErrOverrideReasonRequired   = errors.New("override reason is required")
	ErrCDTenantRequired         = errors.New("controlled document tenant must not be empty")
	ErrCDProfileRequired        = errors.New("controlled document profile must not be empty")
	ErrCDAreaRequired           = errors.New("controlled document process area must not be empty")
	ErrCDCodeRequired           = errors.New("controlled document code must not be empty")
	ErrCDTitleRequired          = errors.New("controlled document title must not be empty")
	ErrCDOwnerRequired          = errors.New("controlled document owner user id must not be empty")
	// ErrApprovalRouteMissing signals the profile a controlled document is
	// being created under has no active approval route, so the document
	// could never be submitted for approval. Creation fails CLOSED on this
	// (product decision D2): no implicit/fallback route, no soft warning.
	// The HTTP layer maps it to the SAME 409 problem+json code the submit
	// path already emits — "state.approval_route_missing" (approval/http
	// errors.go) — so both surfaces speak one wire contract.
	ErrApprovalRouteMissing = errors.New("controlled document profile has no active approval route")
)

// NewControlledDocument trims and validates input, returning a sentinel
// Err*Required error for the first missing mandatory field (tenant,
// profile, area, code, title, owner). It does not check code uniqueness
// or sequence allocation — those are repository/application concerns.
func NewControlledDocument(input ControlledDocument) (*ControlledDocument, error) {
	doc := ControlledDocument{
		ID:                        strings.TrimSpace(input.ID),
		TenantID:                  strings.TrimSpace(input.TenantID),
		ProfileCode:               strings.TrimSpace(input.ProfileCode),
		ProcessAreaCode:           strings.TrimSpace(input.ProcessAreaCode),
		DepartmentCode:            trimOptionalString(input.DepartmentCode),
		Code:                      strings.TrimSpace(input.Code),
		SequenceNum:               input.SequenceNum,
		Title:                     strings.TrimSpace(input.Title),
		OwnerUserID:               strings.TrimSpace(input.OwnerUserID),
		OverrideTemplateVersionID: trimOptionalString(input.OverrideTemplateVersionID),
		Visibility:                input.Visibility,
		Status:                    input.Status,
		CreatedAt:                 input.CreatedAt,
		UpdatedAt:                 input.UpdatedAt,
	}
	if doc.TenantID == "" {
		return nil, ErrCDTenantRequired
	}
	if doc.ProfileCode == "" {
		return nil, ErrCDProfileRequired
	}
	if doc.ProcessAreaCode == "" {
		return nil, ErrCDAreaRequired
	}
	if doc.Code == "" {
		return nil, ErrCDCodeRequired
	}
	if doc.Title == "" {
		return nil, ErrCDTitleRequired
	}
	if doc.OwnerUserID == "" {
		return nil, ErrCDOwnerRequired
	}
	return &doc, nil
}

func trimOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// IsActive reports whether d.Status is CDStatusActive.
func (d ControlledDocument) IsActive() bool {
	return d.Status == CDStatusActive
}

// AutoCode formats the system-generated code for a controlled document as
// "{PROFILE}-{AREA}-{NNN}" (profile and area upper-cased, seq zero-padded
// to 3 digits). Callers that supply a manual code bypass this formatter.
func AutoCode(profileCode, areaCode string, seq int) string {
	return fmt.Sprintf("%s-%s-%03d",
		strings.ToUpper(profileCode),
		strings.ToUpper(areaCode),
		seq)
}
