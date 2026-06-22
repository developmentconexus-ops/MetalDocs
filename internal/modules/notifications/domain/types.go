// Package notificationsdomain defines the shared data-transfer types for the
// notifications module. Placing them here (rather than in infra) breaks the
// cross-layer import: the delivery layer depends on domain types without
// importing the infrastructure package — the distribution module's precedent.
package notificationsdomain

import "time"

// NotificationRow is the flat representation of one per-recipient notification,
// the stored shape behind the generated api.Notification.
type NotificationRow struct {
	ID              string
	TenantID        string
	RecipientUserID string
	EventType       string
	ResourceType    string
	ResourceID      string
	Title           string
	Message         string
	Status          string // "PENDING" | "SENT" | "READ"
	CreatedAt       time.Time
	ReadAt          *time.Time
}

// NotificationsPage is the result of a paginated List call.
type NotificationsPage struct {
	Items      []NotificationRow
	HasMore    bool
	NextCursor string
}
