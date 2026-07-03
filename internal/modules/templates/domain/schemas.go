package domain

import "fmt"

// MetadataSchema describes the document metadata rules a template version
// imposes on documents created from it: the document code pattern, retention
// period, default distribution list, and required metadata fields.
type MetadataSchema struct {
	DocCodePattern      string   `json:"doc_code_pattern"`
	RetentionDays       int      `json:"retention_days"`
	DistributionDefault []string `json:"distribution_default"`
	RequiredMetadata    []string `json:"required_metadata"`
}

// NewMetadataSchema creates a new MetadataSchema, requiring retentionDays to
// be at least 1. Returns ErrInvalidConstraint (wrapped) otherwise.
func NewMetadataSchema(docCodePattern string, retentionDays int, distributionDefault, requiredMetadata []string) (MetadataSchema, error) {
	if retentionDays < 1 {
		return MetadataSchema{}, fmt.Errorf("retention_days must be at least 1: %w", ErrInvalidConstraint)
	}
	return MetadataSchema{
		DocCodePattern:      docCodePattern,
		RetentionDays:       retentionDays,
		DistributionDefault: distributionDefault,
		RequiredMetadata:    requiredMetadata,
	}, nil
}

// PlaceholderType identifies the data type/input kind of a template
// placeholder.
type PlaceholderType string

const (
	// PHText is a free-text placeholder.
	PHText PlaceholderType = "text"
	// PHDate is a date-valued placeholder.
	PHDate PlaceholderType = "date"
	// PHNumber is a numeric placeholder.
	PHNumber PlaceholderType = "number"
	// PHSelect is a placeholder whose value is chosen from a fixed set of options.
	PHSelect PlaceholderType = "select"
	// PHUser is a placeholder whose value identifies a user.
	PHUser PlaceholderType = "user"
	// PHPicture is a placeholder whose value is an image.
	PHPicture PlaceholderType = "picture"
	// PHComputed is a placeholder whose value is derived by a registered resolver.
	PHComputed PlaceholderType = "computed"
	// PHDictionary is a placeholder whose value comes from a native dictionary token.
	PHDictionary PlaceholderType = "dictionary"
)

// VisibilityOp identifies the comparison operator used by a
// VisibilityCondition to decide whether a placeholder is shown.
type VisibilityOp string

const (
	// VisibilityOpEq is the "equal to" comparison operator.
	VisibilityOpEq VisibilityOp = "eq"
	// VisibilityOpNe is the "not equal to" comparison operator.
	VisibilityOpNe VisibilityOp = "ne"
	// VisibilityOpGt is the "greater than" comparison operator.
	VisibilityOpGt VisibilityOp = "gt"
	// VisibilityOpGte is the "greater than or equal to" comparison operator.
	VisibilityOpGte VisibilityOp = "gte"
	// VisibilityOpLt is the "less than" comparison operator.
	VisibilityOpLt VisibilityOp = "lt"
	// VisibilityOpLte is the "less than or equal to" comparison operator.
	VisibilityOpLte VisibilityOp = "lte"
)

// Valid reports whether o is one of the recognized VisibilityOp values.
func (o VisibilityOp) Valid() bool {
	switch o {
	case VisibilityOpEq, VisibilityOpNe, VisibilityOpGt, VisibilityOpGte, VisibilityOpLt, VisibilityOpLte:
		return true
	}
	return false
}

// VisibilityCondition governs whether a placeholder is shown, based on
// comparing another placeholder's value (PlaceholderID) against Value using
// Op.
type VisibilityCondition struct {
	PlaceholderID string       `json:"placeholder_id"`
	Op            VisibilityOp `json:"op"`
	Value         any          `json:"value"`
}

// Placeholder defines a single fill-in field in a template version's
// placeholder schema, including its type, validation constraints, and
// optional visibility condition.
type Placeholder struct {
	ID       string          `json:"id"`
	Name     string          `json:"name,omitempty"`
	Label    string          `json:"label"`
	Type     PlaceholderType `json:"type"`
	Required bool            `json:"required"`
	Default  any             `json:"default,omitempty"`
	Options  []string        `json:"options,omitempty"`

	Regex       *string              `json:"regex,omitempty"`
	MinNumber   *float64             `json:"min_number,omitempty"`
	MaxNumber   *float64             `json:"max_number,omitempty"`
	MinDate     *string              `json:"min_date,omitempty"`
	MaxDate     *string              `json:"max_date,omitempty"`
	MaxLength   *int                 `json:"max_length,omitempty"`
	VisibleIf   *VisibilityCondition `json:"visible_if,omitempty"`
	Computed    bool                 `json:"computed,omitempty"`
	ResolverKey *string              `json:"resolver_key,omitempty"`
}
