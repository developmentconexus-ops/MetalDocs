package domain

import "context"

type authContextKey string

const authPrincipalKey authContextKey = "auth.principal"

// WithCurrentUser returns a copy of ctx carrying user as the request principal.
func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, authPrincipalKey, user)
}

// CurrentUserFromContext returns the principal stored by WithCurrentUser, or
// ok=false when ctx is nil or carries none.
func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	if ctx == nil {
		return CurrentUser{}, false
	}
	user, ok := ctx.Value(authPrincipalKey).(CurrentUser)
	return user, ok
}
