package application

import (
	"context"

	iamdomain "metaldocs/internal/modules/iam/domain"
)

// actorCtx builds the request context every taxonomy governance mutation runs
// under in production: one that carries an authenticated principal.
//
// A3.3 made that a precondition rather than an assumption — the services now
// read the actor through authn.RequireUserID before they open a transaction and
// refuse the call when there is none, so a governance event can no longer be
// written with ActorUserID == "". Tests that exercise the SUCCESS path of a mutation
// therefore have to model the authenticated request the HTTP routes actually
// hand down; a context.Background() there described a call that cannot happen.
func actorCtx() context.Context {
	return iamdomain.WithAuthContext(context.Background(), "user-1", []iamdomain.Role{})
}
