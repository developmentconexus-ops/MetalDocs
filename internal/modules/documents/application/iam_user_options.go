package application

import "context"

type UserOption struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// IAMUserOptionsReader is the consumer-defined port for listing user options.
type IAMUserOptionsReader interface {
	ListUserOptions(ctx context.Context, tenantID string) ([]UserOption, error)
}
