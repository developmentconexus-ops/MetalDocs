package application

import "context"

// UserOption is a selectable user (id + display name) offered when validating
// a PHUser placeholder value.
type UserOption struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// IAMUserOptionsReader is the consumer-defined port for listing user options.
type IAMUserOptionsReader interface {
	ListUserOptions(ctx context.Context, tenantID string) ([]UserOption, error)
}
