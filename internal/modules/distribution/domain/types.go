// Package distributiondomain defines the shared data-transfer types for the
// distribution module. Placing them here (rather than in infra) breaks the
// cross-layer import: the delivery layer can depend on domain types without
// importing the infrastructure package.
package distributiondomain

// RecipientRow is the flat representation of an obligated reader. It carries
// the view columns plus the display name resolved via the iam port.
type RecipientRow struct {
	UserID   string
	Name     string
	AreaCode *string
	AreaName *string
	Source   string
}

// RecipientsPage is the result of a paginated Recipients call.
type RecipientsPage struct {
	Items      []RecipientRow
	HasMore    bool
	NextCursor string
}

// AreaCoverageRow is one row returned by Coverage.
type AreaCoverageRow struct {
	AreaCode string
	AreaName string
	Total    int
}
