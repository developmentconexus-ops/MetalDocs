package apiv2

type ProfileResponse struct {
	Code                     string  `json:"code"`
	TenantID                 string  `json:"tenant_id"`
	FamilyCode               string  `json:"family_code"`
	Name                     string  `json:"name"`
	Description              string  `json:"description"`
	ReviewIntervalDays       int     `json:"review_interval_days"`
	DefaultTemplateVersionID *string `json:"default_template_version_id,omitempty"`
	OwnerUserID              *string `json:"owner_user_id,omitempty"`
	EditableByRole           string  `json:"editable_by_role"`
	ArchivedAt               *string `json:"archived_at,omitempty"`
	CreatedAt                string  `json:"created_at"`
}

type AreaResponse struct {
	Code                string  `json:"code"`
	TenantID            string  `json:"tenant_id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	ParentCode          *string `json:"parent_code,omitempty"`
	OwnerUserID         *string `json:"owner_user_id,omitempty"`
	DefaultApproverRole *string `json:"default_approver_role,omitempty"`
	ArchivedAt          *string `json:"archived_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

type ControlledDocumentResponse struct {
	ID                        string  `json:"id"`
	TenantID                  string  `json:"tenant_id"`
	ProfileCode               string  `json:"profile_code"`
	ProcessAreaCode           string  `json:"process_area_code"`
	DepartmentCode            *string `json:"department_code,omitempty"`
	Code                      string  `json:"code"`
	SequenceNum               *int    `json:"sequence_num,omitempty"`
	Title                     string  `json:"title"`
	OwnerUserID               string  `json:"owner_user_id"`
	OverrideTemplateVersionID *string `json:"override_template_version_id,omitempty"`
	Status                    string  `json:"status"`
	CreatedAt                 string  `json:"created_at"`
	UpdatedAt                 string  `json:"updated_at"`
}

type MembershipResponse struct {
	UserID        string  `json:"user_id"`
	TenantID      string  `json:"tenant_id"`
	AreaCode      string  `json:"area_code"`
	Role          string  `json:"role"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
	GrantedBy     *string `json:"granted_by,omitempty"`
}

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}
