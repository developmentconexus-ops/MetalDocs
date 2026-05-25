package domain

type MetadataSchema struct {
	DocCodePattern      string   `json:"doc_code_pattern"`
	RetentionDays       int      `json:"retention_days"`
	DistributionDefault []string `json:"distribution_default"`
	RequiredMetadata    []string `json:"required_metadata"`
}

func NewMetadataSchema(docCodePattern string, retentionDays int, distributionDefault, requiredMetadata []string) (MetadataSchema, error) {
	return MetadataSchema{
		DocCodePattern:      docCodePattern,
		RetentionDays:       retentionDays,
		DistributionDefault: distributionDefault,
		RequiredMetadata:    requiredMetadata,
	}, nil
}

type PlaceholderType string

const (
	PHText     PlaceholderType = "text"
	PHDate     PlaceholderType = "date"
	PHNumber   PlaceholderType = "number"
	PHSelect   PlaceholderType = "select"
	PHUser     PlaceholderType = "user"
	PHPicture  PlaceholderType = "picture"
	PHComputed PlaceholderType = "computed"
)

type VisibilityOp string

const (
	VisibilityOpEq  VisibilityOp = "eq"
	VisibilityOpNe  VisibilityOp = "ne"
	VisibilityOpGt  VisibilityOp = "gt"
	VisibilityOpGte VisibilityOp = "gte"
	VisibilityOpLt  VisibilityOp = "lt"
	VisibilityOpLte VisibilityOp = "lte"
)

func (o VisibilityOp) Valid() bool {
	switch o {
	case VisibilityOpEq, VisibilityOpNe, VisibilityOpGt, VisibilityOpGte, VisibilityOpLt, VisibilityOpLte:
		return true
	}
	return false
}

type VisibilityCondition struct {
	PlaceholderID string       `json:"placeholder_id"`
	Op            VisibilityOp `json:"op"`
	Value         any          `json:"value"`
}

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

type CompositionConfig struct {
	HeaderSubBlocks []string                  `json:"header_sub_blocks"`
	FooterSubBlocks []string                  `json:"footer_sub_blocks"`
	SubBlockParams  map[string]map[string]any `json:"sub_block_params"`
}
