package domain

// SelectorKind distinguishes the ways a Stage may name who is allowed to act
// on it (M4, unit 3.2 — ActorSelector discriminated union).
type SelectorKind string

// SelectorKind values understood by ActorSelector.Validate.
const (
	SelectorNamedUser          SelectorKind = "named_user"
	SelectorRoleInFixedArea    SelectorKind = "role_in_fixed_area"
	SelectorRoleInDocumentArea SelectorKind = "role_in_document_area"
	SelectorSubmitChoice       SelectorKind = "submit_choice"
)

// ActorSelector names one way a Stage may resolve its eligible actor pool.
// Kind discriminates which fields are meaningful; an empty string in
// UserID/Role/AreaCode mirrors an absent (nullable) DB column.
type ActorSelector struct {
	Kind     SelectorKind `json:"kind"`
	UserID   string       `json:"user_id,omitempty"`
	Role     string       `json:"role,omitempty"`
	AreaCode string       `json:"area_code,omitempty"`
}

// Validate enforces the exact field presence/absence contract for s.Kind:
//
//   - named_user: UserID set, Role and AreaCode absent.
//   - role_in_fixed_area: Role and AreaCode set, UserID absent.
//   - role_in_document_area: Role set, UserID and AreaCode absent.
//   - submit_choice: Role and AreaCode set, UserID absent.
//
// An unrecognized Kind returns ErrInvalidSelectorKind; a recognized Kind with
// the wrong fields present/absent returns ErrSelectorFieldsInvalid.
func (s ActorSelector) Validate() error {
	switch s.Kind {
	case SelectorNamedUser:
		if s.UserID != "" && s.Role == "" && s.AreaCode == "" {
			return nil
		}
	case SelectorRoleInFixedArea:
		if s.Role != "" && s.AreaCode != "" && s.UserID == "" {
			return nil
		}
	case SelectorRoleInDocumentArea:
		if s.Role != "" && s.AreaCode == "" && s.UserID == "" {
			return nil
		}
	case SelectorSubmitChoice:
		if s.Role != "" && s.AreaCode != "" && s.UserID == "" {
			return nil
		}
	default:
		return ErrInvalidSelectorKind
	}
	return ErrSelectorFieldsInvalid
}
