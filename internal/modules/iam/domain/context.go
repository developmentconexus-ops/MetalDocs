package domain

import "context"

type authContextKey string

const (
	authUserIDKey authContextKey = "iam.user_id"
	authRolesKey  authContextKey = "iam.roles"
)

// WithAuthContext returns a copy of ctx carrying userID and roles, retrievable
// via UserIDFromContext / RolesFromContext. This is request-scoped identity
// context, distinct from the tx-local authz GUCs set by authz.SeedTxIdentity.
func WithAuthContext(ctx context.Context, userID string, roles []Role) context.Context {
	ctx = context.WithValue(ctx, authUserIDKey, userID)
	return context.WithValue(ctx, authRolesKey, roles)
}

// UserIDFromContext returns the user id stored by WithAuthContext, or "" if
// ctx is nil or carries none.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value(authUserIDKey).(string)
	return userID
}

// RolesFromContext returns the roles stored by WithAuthContext, or nil if ctx
// is nil or carries none.
func RolesFromContext(ctx context.Context) []Role {
	if ctx == nil {
		return nil
	}
	roles, _ := ctx.Value(authRolesKey).([]Role)
	return roles
}
