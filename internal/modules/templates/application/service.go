// Package application is the use-case/orchestration layer for the templates
// module. It implements the template lifecycle (create, draft editing,
// submit/review/approve/publish, archive), autosave and schema management,
// and approval configuration, coordinating the Repository, Presigner, Clock,
// UUIDGen, and ResolverRegistryReader ports defined in this package. It owns
// authz enforcement (tier-2, in-tx) and audit trail emission for every
// mutating use case, and translates domain errors and CAS conflicts into the
// stable error vocabulary the HTTP layer maps to responses.
package application

import "metaldocs/internal/platform/db"

// Service is the application-layer facade for the templates module's use
// cases. It composes the Repository, Presigner, Clock, UUIDGen, and
// (optional) ResolverRegistryReader ports, plus a db.TxRunner for owning its
// write transactions.
type Service struct {
	repo      Repository
	presign   Presigner
	clock     Clock
	uuid      UUIDGen
	resolvers ResolverRegistryReader
	runner    db.TxRunner
}

// New constructs a Service with the given repository, presigner, clock, UUID
// generator, and an optional resolver registry reader (variadic so callers
// without native resolvers can omit it). The transaction runner must be
// supplied separately via WithRunner before the service can perform writes.
func New(repo Repository, presign Presigner, clock Clock, uuid UUIDGen, resolvers ...ResolverRegistryReader) *Service {
	var registry ResolverRegistryReader
	if len(resolvers) > 0 {
		registry = resolvers[0]
	}
	return &Service{repo: repo, presign: presign, clock: clock, uuid: uuid, resolvers: registry}
}

// WithRunner injects the transaction runner the service uses to own its
// write transactions (H-1d′). The application layer depends on the
// db.TxRunner port, not the concrete *sql.DB pool — the pool is wrapped into
// the port at the composition root.
func (s *Service) WithRunner(runner db.TxRunner) *Service {
	s.runner = runner
	return s
}
