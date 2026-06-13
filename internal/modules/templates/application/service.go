package application

import "metaldocs/internal/platform/db"

type Service struct {
	repo      Repository
	presign   Presigner
	clock     Clock
	uuid      UUIDGen
	resolvers ResolverRegistryReader
	runner    db.TxRunner
}

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
