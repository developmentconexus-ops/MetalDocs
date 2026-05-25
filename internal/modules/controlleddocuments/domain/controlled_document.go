package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type CDStatus string

const (
	CDStatusActive     CDStatus = "active"
	CDStatusObsolete   CDStatus = "obsolete"
	CDStatusSuperseded CDStatus = "superseded"
)

type ControlledDocument struct {
	ID                        string     `json:"id"`
	TenantID                  string     `json:"tenantId"`
	ProfileCode               string     `json:"profileCode"`
	ProcessAreaCode           string     `json:"processAreaCode"`
	DepartmentCode            *string    `json:"departmentCode"`
	Code                      string     `json:"code"`
	SequenceNum               *int       `json:"sequenceNum"`
	Title                     string     `json:"title"`
	OwnerUserID               string     `json:"ownerUserId"`
	OverrideTemplateVersionID *string    `json:"overrideTemplateVersionId"`
	Visibility                Visibility `json:"visibility"`
	Status                    CDStatus   `json:"status"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

var (
	ErrCDNotFound               = errors.New("controlled document not found")
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
)

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

func (d ControlledDocument) IsActive() bool {
	return d.Status == CDStatusActive
}

func AutoCode(profileCode, areaCode string, seq int) string {
	return fmt.Sprintf("%s-%s-%03d",
		strings.ToUpper(profileCode),
		strings.ToUpper(areaCode),
		seq)
}
